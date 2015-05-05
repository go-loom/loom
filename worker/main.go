package worker

import (
	"github.com/koding/kite"
	"gopkg.in/loom.v1/log"
	"gopkg.in/loom.v1/worker/config"
	"os"
)

var logger = log.NewWithSync("loom-worker")

func Main(serverURL, workerConfigFile string, version string) {
	workerConfig, err := config.ReadFile(workerConfigFile)
	if err != nil {
		logger.Error("worker config err: %v", err)
		os.Exit(1)
	}
	logger.Info("Read config file: %v", workerConfigFile)

	workerName := "loom-worker-" + workerConfig.Name

	k := kite.New(workerName, version)
	k.Config.DisableAuthentication = true

	w := NewWorker(serverURL, workerConfig, k)
	err = w.Init()
	if err != nil {
		logger.Error("Worker init err: %v", err)
		os.Exit(1)
	}

	k.Run()
}
