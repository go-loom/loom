package worker

import (
	"github.com/koding/kite"
	"github.com/mgutz/logxi/v1"
	"gopkg.in/loom.v1/worker/config"
	"os"
)

var logger = log.NewLogger(log.NewConcurrentWriter(os.Stdout), "loom-worker")

func Main(serverURL, workerConfigFile string, version string) {
	workerConfig, err := config.ReadFile(workerConfigFile)
	if err != nil {
		logger.Error("worker config", "err", err)
		os.Exit(1)
	}
	logger.Info("read", "config", workerConfigFile)

	workerName := "loom-worker-" + workerConfig.Name

	k := kite.New(workerName, version)
	k.Config.DisableAuthentication = true

	w := NewWorker(serverURL, workerConfig, k)
	err = w.Init()
	if err != nil {
		logger.Error("worker", "err", err)
		os.Exit(1)
	}

	k.Run()
}
