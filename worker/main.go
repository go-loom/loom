package worker

import (
	"github.com/koding/kite"
	"github.com/go-loom/loom/expvar"
	"github.com/go-loom/loom/log"
	"os"
)

var logger = log.NewWithSync("loom-worker")

func New(serverURL string, topic string, maxJobSize int, version string) *Worker {
	workerName := "loom-worker#" + topic

	k := kite.New(workerName, version)
	k.Config.DisableAuthentication = true

	k.HandleHTTPFunc("/debug/vars", expvar.ExpvarHandler)

	w := NewWorker(serverURL, topic, maxJobSize, k)
	err := w.Init()
	if err != nil {
		logger.Error("Worker init err: %v", err)
		os.Exit(1)
	}

	return w
}
