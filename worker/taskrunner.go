package worker

import (
	"bytes"
	"errors"
	"github.com/looplab/fsm"
	"github.com/matryer/try"
	"gopkg.in/loom.v1/log"
	"gopkg.in/loom.v1/worker/config"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/http/httputil"
	"os"
	"os/exec"
	"strings"
	"time"
)

const (
	HTTP_GET  = "GET"
	HTTP_POST = "POST"
)

var (
	HTTPMethodNotSupport = errors.New("Http method is not supported")
)

type TaskRunner struct {
	job       *Job
	task      *config.Task
	err       error
	output    string
	fsm       *fsm.FSM
	eventC    chan string
	stateC    chan string
	logger    log.Logger
	startTime time.Time
	endTime   time.Time
}

func NewTaskRunner(job *Job, task *config.Task) *TaskRunner {
	tr := &TaskRunner{
		job:    job,
		task:   task,
		eventC: make(chan string),
		stateC: make(chan string),
		logger: log.New("taskrunner#" + task.TaskName()),
	}
	tr_fsm := fsm.NewFSM(
		"INIT",
		fsm.Events{
			{Name: "run", Src: []string{"INIT"}, Dst: "PROCESS"},
			{Name: "success", Src: []string{"PROCESS"}, Dst: "DONE"},
			{Name: "cancel", Src: []string{"INIT", "PROCESS"}, Dst: "CANCEL"},
			{Name: "error", Src: []string{"PROCESS"}, Dst: "ERROR"},
		},
		fsm.Callbacks{
			"enter_state": func(e *fsm.Event) {
				tr.logger.Debug("S:enter state:%v", e.Dst)
				tr.stateC <- e.Dst
				tr.logger.Debug("E:enter state:%v", e.Dst)
			},
		},
	)
	tr.fsm = tr_fsm

	go tr.stateListening()
	go tr.eventListening()

	tr.logger.Info("Start task runner")
	return tr
}

func (tr *TaskRunner) Run() {
	tr.eventC <- "run"
}

func (tr *TaskRunner) Cancel() {
	tr.eventC <- "cancel"
}

func (tr *TaskRunner) stateListening() {
	for {
		select {
		case state := <-tr.stateC:
			tr.logger.Debug("stateC:%v", state)
			if state == "PROCESS" {
				err := tr.processing()
				if err != nil {
					tr.eventC <- "error"
				} else {
					tr.eventC <- "success"
				}
			} else {
				tr.job.DoneTask(tr)
				tr.eventC <- "quit"
				break
			}
		}
	}
	tr.logger.Info("end stateListening")
}

func (tr *TaskRunner) eventListening() {
	tr.startTime = time.Now()
	for {
		select {
		case event := <-tr.eventC:
			if event == "quit" {
				break
			}
			err := tr.fsm.Event(event)
			if err != nil {
				//TODO:
				tr.logger.Error("fsm event error: %v", err)
			}
			if tr.logger.IsDebug() {
				tr.logger.Debug("eventListening event:%v", event)
			}
		}
	}

	tr.endTime = time.Now()
	tr.logger.Info("End taskrunner")
}

func (tr *TaskRunner) processing() error {
	var processFunc func() error

	if tr.task.Cmd != "" {
		processFunc = tr.cmd
	} else if tr.task.HTTP != nil {
		processFunc = tr.http
	}

	err := try.Do(func(attempt int) (bool, error) {
		err := processFunc()
		if err != nil {
			d, _ := tr.task.Retry.GetInterval()
			time.Sleep(d)
		}
		return attempt < tr.task.Retry.Number, err
	})
	if err != nil {
		tr.err = err
	}
	return err
}

func (tr *TaskRunner) cmd() error {
	cmdstr := tr.task.Cmd
	cmd := exec.Command("bash", "-c", cmdstr)
	out, err := cmd.CombinedOutput()
	tr.output = string(out)

	if tr.logger.IsDebug() {
		tr.logger.Debug("cmd: %s output: %v", cmdstr, tr.output)
	}

	return err
}

func (tr *TaskRunner) http() (err error) {
	method := strings.ToUpper(tr.task.HTTP.Method)

	switch method {
	case HTTP_GET:
		err = tr.httpGet()
	case HTTP_POST:
		err = tr.httpPost()
	default:
		err = HTTPMethodNotSupport
	}

	return nil
}

func (tr *TaskRunner) httpGet() error {
	httpcfg := tr.task.HTTP
	url := httpcfg.URL
	resp, err := http.Get(url)
	if err != nil {
		return err
	}

	body, err := httputil.DumpResponse(resp, true)
	if err != nil {
		return err
	}

	tr.output = string(body)
	return nil
}

func (tr *TaskRunner) httpPost() error {
	httpcfg := tr.task.HTTP
	url := httpcfg.URL
	params := httpcfg.Data
	files := httpcfg.Files

	body := new(bytes.Buffer)
	writer := multipart.NewWriter(body)

	//file handling
	if len(files) > 0 {
		for _, file := range files {
			if err := file.Err(); err != nil {
				return err
			}

			_file, err := os.Open(file.Path)
			if err != nil {
				return err
			}

			fi, _ := _file.Stat()
			part, err := writer.CreateFormFile(file.Filename, fi.Name())
			if err != nil {
				return err
			}
			fileContents, err := ioutil.ReadAll(_file)
			if err != nil {
				return err
			}
			part.Write(fileContents)
		}
	}

	//data handling
	for key, val := range params {
		_ = writer.WriteField(key, val)
	}

	err := writer.Close()
	if err != nil {
		return err
	}

	req, err := http.NewRequest(HTTP_POST, url, body)
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	dump_resp, err := httputil.DumpResponse(resp, true)
	if err != nil {
		return err
	}
	tr.output = string(dump_resp)

	return nil
}

//Implement Task interface

func (tr *TaskRunner) TaskName() string {
	return tr.task.Name
}

func (tr *TaskRunner) Ok() bool {
	if tr.err == nil {
		return true
	}
	return false
}

func (tr *TaskRunner) State() string {
	return tr.fsm.Current()
}

func (tr *TaskRunner) Err() error {
	return tr.err
}

func (tr *TaskRunner) Output() string {
	return tr.output
}

func (tr *TaskRunner) StartEndTimes() []*time.Time {
	return []*time.Time{&tr.startTime, &tr.endTime}
}
