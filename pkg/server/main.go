package server

import (
	"github.com/go-loom/loom/pkg/expvar"
	"github.com/go-loom/loom/pkg/log"
	"github.com/go-loom/loom/pkg/rpc/pb"
	"github.com/go-loom/loom/pkg/util"
	"github.com/go-loom/loom/pkg/version"

	"github.com/gorilla/mux"
	"github.com/oklog/run"

	"context"
	"fmt"
	"net"
	"net/http"
)

func Main(port int, dbpath string) error {

	apiNet := "tcp"
	apiAddr := fmt.Sprintf("0.0.0.0:%d", port)
	apiListener, err := net.Listen(apiNet, apiAddr)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithCancel(context.Background())
	broker := NewBroker(ctx, dbpath)
	if err := broker.Init(); err != nil {
		return err
	}

	var g run.Group
	{
		g.Add(func() error {
			r := mux.NewRouter()

			twirpHandler := pb.NewLoomServer(broker, nil)
			r.Handle(pb.LoomPathPrefix, twirpHandler)

			httpApiHandler := &httpApiHandler{
				broker: broker,
			}

			r.HandleFunc("/v1/queues/{queue}", httpApiHandler.PushHandler)
			r.HandleFunc("/v1/queues/{queue}/{id}", httpApiHandler.GetHandler)
			r.HandleFunc("/debug/vars", expvar.ExpvarHandler)

			return http.Serve(apiListener, r)

		}, func(error) {
			apiListener.Close()
			cancel()
			broker.Done()
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

	log.Logger.Log("server", "started", "version", version.Version, "commit", version.GitCommit, "build", version.BuildDate)
	return g.Run()
}
