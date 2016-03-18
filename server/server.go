package server

import (
	"github.com/koding/kite"
	"golang.org/x/net/context"
	"gopkg.in/loom.v1/expvar"
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

	//restRouter.DELETE("/v1/queues/:queue/:id", DeleteHandler)
	//TODO:  consider later.

	k.HandleHTTPFunc("/v1/queues/{queue}", PushHandler)
	k.HandleHTTPFunc("/v1/queues/{queue}/{id}", GetHandler)

	k.HandleHTTPFunc("/debug/vars", expvar.ExpvarHandler)
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
