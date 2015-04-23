package server

import (
	"github.com/koding/kite"
	"github.com/koding/logging"
)

var logger = logging.NewLogger("loom-server")

var broker *Broker

func Main(host string, port int, dbpath string, version string) {

	broker = NewBroker(dbpath)

	k := kite.New("loom-server", version)

	k.Config.Port = port
	k.Run()
}
