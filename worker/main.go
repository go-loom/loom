package worker

import (
	"github.com/koding/kite"
	"github.com/mgutz/logxi/v1"
	"os"
)

var logger = log.NewLogger(log.NewConcurrentWriter(os.Stdout), "loom-worker")

func Main(serverURL string, version string) {
	k := kite.New("loom-worker", version)
	k.Config.DisableAuthentication = true

	w := NewWorker(serverURL, k)
	err := w.Init()
	if err != nil {
		logger.Error("worker", "err", err)
		os.Exit(1)
	}

	k.Run()
}
