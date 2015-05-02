package server

import (
	"github.com/julienschmidt/httprouter"
	"github.com/koding/kite"
	"github.com/mgutz/logxi/v1"
	"os"
)

var logger = log.NewLogger(log.NewConcurrentWriter(os.Stdout), "loom-server")

var broker *Broker

func Main(port int, dbpath string, version string) {

	k := kite.New("loom-server", version)
	k.Config.DisableAuthentication = true

	broker = NewBroker(dbpath, k)
	err := broker.Init()
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}

	restRouter := httprouter.New()
	restRouter.POST("/v1/queues/:queue", PushHandler)
	restRouter.GET("/v1/queues/:queue/:id", GetHandler)
	restRouter.DELETE("/v1/queues/:queue/:id", DeleteHandler)

	k.HandleHTTP("/v1/queues/", restRouter)

	k.Config.Port = port
	k.Run()
}
