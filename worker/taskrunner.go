package worker

import (
	"bytes"
	"errors"
	"github.com/looplab/fsm"
	"github.com/matryer/try"
	"gopkg.in/loom.v1/config"
	"gopkg.in/loom.v1/log"
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

const (
	TASK_STATE_INIT    = "INIT"
	TASK_STATE_PROCESS = "PROCESS"
	TASK_STATE_DONE    = "DONE"
	TASK_STATE_CANCEL  = "CANCEL"
	TASK_STATE_ERROR   = "ERROR"

	TASK_EVENT_RUN     = "run"
	TASK_EVENT_SUCCESS = "success"
	TASK_EVENT_CANCEL  = "cancel"
	TASK_EVENT_ERROR   = "error"

	TASK_QUIT = "quit"
)

var (
	HTTPMethodNotSupport = errors.New("Http method is not supported")
)

type TaskRunner struct {
	job         *Job
	task        *config.Task
	err         error
	output      string
	fsm         *fsm.FSM
	eventC      chan string
	stateC      chan string
	logger      log.Logger
	startTime   time.Time
	endTime     time.Time
	templateCtx map[string]interface{}
}

func NewTaskRunner(job *Job, task *config.Task, templateCtx map[string]interface{}) *TaskRunner {
	tr := &TaskRunner{
		job:         job,
		task:        task,
		eventC:      make(chan string),
		stateC:      make(chan string),
		templateCtx: templateCtx,
		logger:      log.New("taskr/" + task.TaskName() + "#" + job.ID),
	}
	tr_fsm := fsm.NewFSM(
		TASK_STATE_INIT,
		fsm.Events{
			{
				Name: TASK_EVENT_RUN,
				Src:  []string{TASK_STATE_INIT},
				Dst:  TASK_STATE_PROCESS,
			},
			{
				Name: TASK_EVENT_SUCCESS,
				Src:  []string{TASK_STATE_PROCESS},
				Dst:  TASK_STATE_DONE,
			},
			{
				Name: TASK_EVENT_CANCEL,
				Src:  []string{TASK_STATE_INIT, TASK_STATE_PROCESS},
				Dst:  TASK_STATE_CANCEL,
			},
			{
				Name: TASK_EVENT_ERROR,
				Src:  []string{TASK_STATE_PROCESS},
				Dst:  TASK_STATE_ERROR,
			},
		},
		fsm.Callbacks{
			"enter_state": func(e *fsm.Event) {

				tr.stateC <- e.Dst

				if tr.logger.IsDebug() {
					tr.logger.Debug("Enter state src:%v dst:%v", e.Src, e.Dst)
				}
			},
		},
	)
	tr.fsm = tr_fsm

	go tr.stateListening()
	go tr.eventListening()

	tr.logger.Info("Start")
	return tr
}

func (tr *TaskRunner) Run() {
	tr.eventC <- TASK_EVENT_RUN
}

func (tr *TaskRunner) Cancel() {
	tr.eventC <- TASK_EVENT_CANCEL
}

func (tr *TaskRunner) stateListening() {
L:
	for {
		select {
		case state := <-tr.stateC:

			//Call state change handlers
			tr.job.OnTaskChanged(tr)

			if state == TASK_STATE_PROCESS {
				err := tr.processing()
				if err != nil {
					tr.eventC <- TASK_EVENT_ERROR
				} else {
					tr.eventC <- TASK_EVENT_SUCCESS
				}
			} else {
				tr.job.OnTaskDone(tr)
				tr.eventC <- TASK_QUIT
				break L
			}
		}
	}
}

func (tr *TaskRunner) eventListening() {
	tr.startTime = time.Now()
L:
	for {
		select {
		case event := <-tr.eventC:
			if event == TASK_QUIT {
				break L
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
	tr.logger.Info("End")
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

	cmdstr, err := tr.task.Read(tr.task.Cmd, tr.templateCtx)
	if err != nil {
		tr.logger.Error("cmd statement has template error:%v", err)
		return err
	}
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

	return
}

func (tr *TaskRunner) httpGet() error {
	httpcfg := tr.task.HTTP
	url := httpcfg.URL
	url, err := tr.task.Read(url, tr.templateCtx)
	if err != nil {
		tr.logger.Error("http/get has an template error:%v", err)
		return err
	}

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
	tctx := tr.templateCtx

	httpcfg := tr.task.HTTP

	url := httpcfg.URL
	url, err := tr.task.Read(url, tctx)
	if err != nil {
		return err
	}

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

			path, err := tr.task.Read(file.Path, tctx)
			if err != nil {
				return err
			}

			_file, err := os.Open(path)
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
		val, err = tr.task.Read(val, tctx)
		if err != nil {
			tr.logger.Error("http:post has an template error:%v", err)
			return err
		}

		_ = writer.WriteField(key, val)
	}

	err = writer.Close()
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
