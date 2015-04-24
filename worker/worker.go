package worker

import (
	"fmt"
	"github.com/koding/kite"
	"github.com/mgutz/logxi/v1"
	"gopkg.in/loom.v1/server"
	"time"
)

type Worker struct {
	ID        string
	ServerURL string
	kite      *kite.Kite
	Client    *kite.Client
	logger    log.Logger
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

	err := w.Client.Dial()
	if err != nil {
		return err
	}

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
		msg, err := w.Client.Tell("loom.worker.pop.message", "test")
		w.logger.Debug("pop.message", "msg", msg, ",err", err)
		time.Sleep(1 * time.Second)
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
