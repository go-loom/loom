package main

import (
	"github.com/go-loom/loom/pkg/server"
	"github.com/go-loom/loom/pkg/version"
	"github.com/go-loom/loom/pkg/worker"

	"github.com/codegangsta/cli"

	"fmt"
	_ "github.com/seanpont/assert"
	"os"
)

func main() {
	app := cli.NewApp()
	app.Name = "loom"
	app.Usage = "Distributed task processing tool"
	app.Version = fmt.Sprintf("%v git=%v build=%v", version.Version, version.GitCommit, version.BuildDate)
	app.Flags = []cli.Flag{}

	app.Commands = []cli.Command{
		{
			Name:      "server",
			ShortName: "s",
			Usage:     "Run message queues server and worker manager",
			Action:    ServerAction,
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:   "dbpath,d",
					Value:  "/var/data/loom",
					Usage:  "path for message db",
					EnvVar: "DBPATH",
				},
				cli.IntFlag{
					Name:   "port,p",
					Value:  7000,
					Usage:  "port for the server",
					EnvVar: "PORT",
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
					Name:   "serverurl,s",
					Value:  "http://localhost:7000",
					EnvVar: "SERVER_URL",
				},
				cli.StringFlag{
					Name:   "topic,t",
					Value:  "",
					EnvVar: "TOPIC",
				},
				cli.IntFlag{
					Name:   "maxJobSize,m",
					Value:  10,
					Usage:  "max job size per worker",
					EnvVar: "MAX_JOB_SIZE",
				},
				cli.StringFlag{
					Name:   "name,n",
					Value:  "worker1",
					EnvVar: "WORKER_NAME",
				},
				cli.IntFlag{
					Name:   "port,p",
					Value:  7001,
					Usage:  "worker port",
					EnvVar: "WORKER_PORT",
				},
			},
		},
	}

	app.Run(os.Args)
}

func ServerAction(c *cli.Context) error {
	dbpath := c.String("dbpath")
	port := c.Int("port")
	return server.Main(port, dbpath)
}

func WorkerAction(c *cli.Context) error {
	serverURL := c.String("serverurl")
	topic := c.String("topic")
	maxJobSize := c.Int("maxJobSize")
	workerName := c.String("name")
	workerPort := c.Int("port")
	return worker.Main(serverURL, topic, maxJobSize, workerName, workerPort)
}
