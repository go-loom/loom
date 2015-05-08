package main

import (
	"github.com/codegangsta/cli"
	"gopkg.in/loom.v1/server"
	"gopkg.in/loom.v1/worker"
	"os"

	//	"github.com/looplab/fsm"
	//	"github.com/seanpont/assert"
)

func main() {
	app := cli.NewApp()
	app.Name = "loom"
	app.Usage = "Distributed task processing tool"
	app.Version = VERSION
	//app.Action = WorkerAction

	app.Flags = []cli.Flag{}

	app.Commands = []cli.Command{
		{
			Name:      "server",
			ShortName: "s",
			Usage:     "Run message queues server and worker manager",
			Action:    ServerAction,
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "dbpath,d",
					Value: "/var/data/loom",
					Usage: "path for message db",
				},
				cli.IntFlag{
					Name:  "port,p",
					Value: 7000,
					Usage: "port for the server",
				},
			},
		},
		{
			Name:      "worker",
			ShortName: "w",
			Usage:     "",
			Action:    WorkerAction,
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "serverurl,s",
					Value: "http://localhost:7000/kite",
				},
				cli.StringFlag{
					Name:  "config,c",
					Value: "./worker.yaml",
				},
			},
		},
	}

	app.Run(os.Args)
}

func ServerAction(c *cli.Context) {
	dbpath := c.String("dbpath")
	port := c.Int("port")
	server.Main(port, dbpath, VERSION)
}

func WorkerAction(c *cli.Context) {
	serverURL := c.String("serverurl")
	workerConfigFile := c.String("config")
	worker.Main(serverURL, workerConfigFile, VERSION)

}
