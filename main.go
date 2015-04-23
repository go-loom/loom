package main

import (
	"github.com/codegangsta/cli"
	"gopkg.in/loom.v1/server"
	"os"
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
				cli.StringFlag{
					Name:  "host,h",
					Value: "",
					Usage: "host address for the server",
				},
			},
		},
	}

	app.Run(os.Args)
}

func ServerAction(c *cli.Context) {
	dbpath := c.String("dbpath")
	port := c.Int("port")
	host := c.String("host")

	server.Main(host, port, dbpath, VERSION)
}
