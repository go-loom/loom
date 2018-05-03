package worker

import (
	"github.com/go-loom/loom/pkg/config"
	"github.com/go-loom/loom/pkg/log"

	kitlog "github.com/go-kit/kit/log"
	"github.com/looplab/fsm"
	"github.com/matryer/try"

	"bytes"
	"errors"
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
	logger      kitlog.Logger
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
		logger:      log.With(log.Logger, "task", task.TaskName(), "job", job.ID),
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

				log.Debug(tr.logger).Log("msg", "enter_state", "src", e.Src, "dst", e.Dst)

			},
			"enter_PROCESS": func(e *fsm.Event) {
				tr.startTime = time.Now()
			},
			"leave_PROCESS": func(e *fsm.Event) {
				tr.endTime = time.Now()
			},
		},
	)
	tr.fsm = tr_fsm

	go tr.stateListening()
	go tr.eventListening()

	log.Info(tr.logger).Log("msg", "Start")
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
				log.Error(tr.logger).Log("msg", "fsm event", "err", err)
			}
			log.Debug(tr.logger).Log("msg", "eventListening", "event", event)
		}
	}

	log.Info(tr.logger).Log("msg", "end")
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
			log.Info(tr.logger).Log("msg", "Retry processing attempt", "num", attempt, "err", err)
			d, err := tr.task.Retry.GetDelayTime()
			if err != nil {
				log.Error(tr.logger).Log("msg", "retry delaytime", "err", err)
			}
			time.Sleep(d)
		}
		return attempt < tr.task.Retry.Number, err
	})
	if err != nil {
		tr.err = err
	}
	return err
}

func (tr *TaskRunner) cmd() (err error) {

	cmdstr, err := tr.task.Read(tr.task.Cmd, tr.templateCtx)
	if err != nil {
		log.Error(tr.logger).Log("msg", "cmd statement has wrong template", "err", err)
		return err
	}

	timeout, err := tr.task.Retry.GetTimeout()
	if err != nil {
		log.Error(tr.logger).Log("msg", "task retry timeout", "err", err)
		timeout = nil
	}

	cmd := exec.Command("bash", "-c", cmdstr)

	var b bytes.Buffer
	cmd.Stdout = &b
	cmd.Stderr = &b

	if err := cmd.Start(); err != nil {
		log.Error(tr.logger).Log("cmd", cmdstr, "err", err)
		return err
	}

	done := make(chan error, 1)
	go func() {
		done <- cmd.Wait()
	}()

	if timeout != nil {
		select {
		case <-time.After(*timeout):
			if err := cmd.Process.Kill(); err != nil {
				log.Error(tr.logger).Log("msg", "failed to kill", "err", err)
			}
			err = <-done
			if err != nil {
				log.Error(tr.logger).Log("msg", "Process timeout", "err", err)
			}
		case err = <-done:
			if err != nil {
				log.Error(tr.logger).Log("msg", "Process done", "err", err)
			}
		}
	} else {
		err = <-done
		if err != nil {
			log.Error(tr.logger).Log("msg", "Process done", "err", err)
		}
	}

	tr.output = b.String()

	log.Debug(tr.logger).Log("cmd", cmdstr, "output", tr.output)

	return
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
		log.Error(tr.logger).Log("msg", "http/get has an template err", "err", err)
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
			log.Error(tr.logger).Log("msg", "http:post has an template err", "err", err)
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
