package worker

import (
	"github.com/koding/kite"
	"gopkg.in/loom.v1/log"
	"os"
)

var logger = log.NewWithSync("loom-worker")

func Main(serverURL, topic, version string) {
	workerName := "loom-worker-" + topic

	k := kite.New(workerName, version)
	k.Config.DisableAuthentication = true

	w := NewWorker(serverURL, topic, k)
	err := w.Init()
	if err != nil {
		logger.Error("Worker init err: %v", err)
		os.Exit(1)
	}

	k.Run()
}
