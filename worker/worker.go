package worker

import (
	"encoding/json"
	"fmt"
	"github.com/koding/kite"
	"golang.org/x/net/context"
	"gopkg.in/loom.v1/log"
	"gopkg.in/loom.v1/worker/config"
	"sync"
	"time"
)

type Worker struct {
	ID             string
	ServerURL      string
	kite           *kite.Kite
	Client         *kite.Client
	workerConfig   *config.Worker
	works          map[string]*Work
	worksMutex     sync.RWMutex
	logger         log.Logger
	connected      bool
	connectedMutex sync.RWMutex
	ctx            context.Context
}

func NewWorker(serverURL string, workerConfig *config.Worker, k *kite.Kite) *Worker {
	w := &Worker{
		ID:           k.Id,
		ServerURL:    serverURL,
		workerConfig: workerConfig,
		works:        make(map[string]*Work, 0),
		kite:         k,
		logger:       log.New("worker"),
		ctx:          context.Background(), //TODO:
	}

	return w
}

func (w *Worker) Init() error {
	w.Client = w.kite.NewClient(w.ServerURL)

	logger.Info("Worker ID: %v", w.ID)

	connected, err := w.Client.DialForever()
	<-connected
	if err != nil {
		return err
	}

	w.connectedMutex.Lock()
	w.connected = true
	w.connectedMutex.Unlock()

	w.Client.OnConnect(func() {
		err := w.tellHelloToServer()
		if err != nil {
			return
		}

		w.connectedMutex.Lock()
		w.connected = true
		w.connectedMutex.Unlock()
	})
	w.Client.OnDisconnect(func() {
		w.connectedMutex.Lock()
		w.connected = false
		w.connectedMutex.Unlock()
	})

	w.tellHelloToServer()
	w.kite.HandleFunc("loom.worker:message.pop", w.HandleMessagePop)

	return nil
}

func (w *Worker) tellHelloToServer() error {
	response, err := w.Client.Tell("loom.server:worker.connect", w.ID, w.workerConfig.Topic)
	if err != nil {
		return err
	}

	ok := response.MustBool()
	if !ok {
		return fmt.Errorf("registering worker to server is failed")
	}

	return nil
}

func (w *Worker) tellJobDoneToServer(work *Work) error {
	_, err := w.Client.Tell("loom.server:worker.job", work.Job["ID"], w.workerConfig.Topic, "done", work.tasks.MapInfo(), w.ID)
	if err != nil {
		logger.Error("tellJobDoneToServer err: %v", err)
		return err
	}
	return nil
}

func (w *Worker) HandleMessagePop(r *kite.Request) (interface{}, error) {
	msg, err := r.Args.One().Map()
	if err != nil {
		return false, err
	}

	value := msg["value"].MustString()
	var valueMap map[string]interface{}
	err = json.Unmarshal([]byte(value), &valueMap)
	if err != nil {
		logger.Error("json err: %v", err)
		return false, err
	}

	if logger.IsDebug() {
		logger.Debug("message id: %v value: %v", msg["id"], valueMap)
	}

	job := NewJob(msg["id"].MustString(), valueMap)

	work := NewWork(w.ctx, job, w.workerConfig)
	work.OnEnd(func(work *Work) {
		if w.logger.IsDebug() {
			w.logger.Debug("workdone tasks: %v", work.tasks)
		}
		for err := w.tellJobDoneToServer(work); err != nil; {
			time.Sleep(1 * time.Second)
		}

		w.worksMutex.Lock()
		delete(w.works, work.ID)
		w.worksMutex.Unlock()

	})
	work.Run() //async.

	w.worksMutex.Lock()
	w.works[work.ID] = work
	w.worksMutex.Unlock()

	if w.logger.IsDebug() {
		w.logger.Debug("Pop message: %v job: %v", msg, job)
	}

	return true, nil
}
