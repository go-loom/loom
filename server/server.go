package server

import (
	"github.com/julienschmidt/httprouter"
	"github.com/koding/kite"
	"golang.org/x/net/context"
	"gopkg.in/loom.v1/log"
	"os"
)

var logger = log.NewWithSync("loom-server")
var broker *Broker //Use for http handlers

type Server struct {
	broker  *Broker
	kite    *kite.Kite
	ctx     context.Context
	cancelF context.CancelFunc
}

func NewServer(port int, dbpath string, version string) *Server {
	ctx, cancelF := context.WithCancel(context.Background())
	k := kite.New("loom-server", version)
	k.Config.DisableAuthentication = true

	b := NewBroker(ctx, dbpath, k)
	broker = b //TODO:
	err := b.Init()
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}

	s := &Server{
		broker:  b,
		kite:    k,
		ctx:     ctx,
		cancelF: cancelF,
	}

	restRouter := httprouter.New()
	restRouter.POST("/v1/queues/:queue", PushHandler)
	restRouter.GET("/v1/queues/:queue/:id", GetHandler)
	//restRouter.DELETE("/v1/queues/:queue/:id", DeleteHandler)
	//TODO:  consider later.

	k.HandleHTTP("/v1/queues/", restRouter)
	k.Config.Port = port

	return s
}

func (s *Server) Run() {
	s.kite.Run()
}

func (s *Server) Close() error {
	s.cancelF()
	s.kite.Close()
	s.broker.Done()
	return nil
}
