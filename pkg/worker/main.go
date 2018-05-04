package worker

import (
	"github.com/go-loom/loom/pkg/expvar"
	"github.com/go-loom/loom/pkg/log"
	"github.com/go-loom/loom/pkg/util"
	"github.com/go-loom/loom/pkg/version"

	"github.com/oklog/run"

	"context"
	"fmt"
	"net"
	"net/http"
)

func Main(serverURL, topic string, maxJobSize int, workerName string, workerPort int) error {
	ctx, cancel := context.WithCancel(context.Background())

	apiListener, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", workerPort))
	if err != nil {
		log.Error(log.Logger).Log("err", err)
		return err
	}

	var g run.Group
	{
		g.Add(func() error {
			worker := NewWorker(ctx, workerName, serverURL, topic, maxJobSize)
			if err := worker.Init(); err != nil {
				log.Error(log.Logger).Log("err", err)
				return err
			}
			worker.Run()
			return nil
		}, func(error) {
			cancel()
		})
	}
	{
		g.Add(func() error {
			mux := http.NewServeMux()
			mux.HandleFunc("/debug/vars", expvar.ExpvarHandler)

			return http.Serve(apiListener, mux)

		}, func(error) {
			apiListener.Close()
		})
	}
	{
		cancel := make(chan struct{})
		g.Add(func() error {
			return util.Interrupt(cancel)
		}, func(error) {
			close(cancel)
		})
	}

	log.Logger.Log("worker", "started", "name", workerName, "version", version.Version, "commit", version.GitCommit, "build", version.BuildDate)
	return g.Run()
}
