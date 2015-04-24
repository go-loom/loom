package worker

import (
	"fmt"
	"github.com/koding/kite"
	"github.com/mgutz/logxi/v1"
	"gopkg.in/loom.v1/server"
	"sync"
)

type Worker struct {
	ID             string
	ServerURL      string
	kite           *kite.Kite
	Client         *kite.Client
	logger         log.Logger
	connected      bool
	connectedMutex sync.RWMutex
}

func NewWorker(serverURL string, k *kite.Kite) *Worker {
	w := &Worker{
		ID:        k.Id,
		ServerURL: serverURL,
		kite:      k,
		logger:    log.New("worker"),
	}

	return w
}

func (w *Worker) Init() error {
	w.Client = w.kite.NewClient(w.ServerURL)

	logger.Info("worker", "id", w.ID)

	connected, err := w.Client.DialForever()
	<-connected
	if err != nil {
		return err
	}

	w.connectedMutex.Lock()
	w.connected = true
	w.connectedMutex.Unlock()

	w.Client.OnConnect(func() {
		w.connectedMutex.Lock()
		w.connected = true
		w.connectedMutex.Unlock()
	})
	w.Client.OnDisconnect(func() {
		w.connectedMutex.Lock()
		w.connected = false
		w.connectedMutex.Unlock()
	})

	response, err := w.Client.Tell("loom.worker.init", w.ID, "test")
	if err != nil {
		return err
	}

	ok := response.MustBool()
	if !ok {
		return fmt.Errorf("registering worker to server is failed")
	}

	w.kite.HandleFunc("loom.worker.pop", w.HandleMessagePop)

	go w.msgPump()

	return nil
}

func (w *Worker) msgPump() {
	for {
		w.connectedMutex.RLock()
		connected := w.connected
		w.connectedMutex.RUnlock()

		if connected {
			resp, err := w.Client.Tell("loom.worker.pop.message", "test")
			if err != nil {
				w.logger.Debug("pop.message", "err", err, "connected", connected)
			} else {
				msg := resp.MustMap()
				w.logger.Debug("pop.message", "msg", msg, "err", err, "connected", connected)
			}
		}
	}
}

func (w *Worker) HandleMessagePop(r *kite.Request) (interface{}, error) {
	var msg server.Message
	err := r.Args.One().Unmarshal(msg)
	if err != nil {
		return nil, err
	}

	w.logger.Debug("popmessage", "msg", msg)

	return w.ID, nil
}
