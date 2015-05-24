package main

import (
	"github.com/codegangsta/cli"
	"gopkg.in/loom.v1/server"
	"gopkg.in/loom.v1/worker"
	"log"
	"os"
	"os/signal"
	"syscall"

	//"github.com/seanpont/assert"
)

func main() {
	app := cli.NewApp()
	app.Name = "loom"
	app.Usage = "Distributed task processing tool"
	app.Version = VERSION

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
					Value:  "http://localhost:7000/kite",
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
			},
		},
	}

	app.Run(os.Args)
}

func ServerAction(c *cli.Context) {
	dbpath := c.String("dbpath")
	port := c.Int("port")
	s := server.NewServer(port, dbpath, VERSION)
	shutdown(s.Close)
	s.Run()
}

func WorkerAction(c *cli.Context) {
	serverURL := c.String("serverurl")
	topic := c.String("topic")
	maxJobSize := c.Int("maxJobSize")
	w := worker.New(serverURL, topic, maxJobSize, VERSION)
	shutdown(w.Close)
	w.Run()
}

func shutdown(callback func() error) {
	c := make(chan os.Signal, 2)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	go func() {
		s := <-c
		log.Println("Got signal: ", s)
		err := callback()
		if err != nil {
			log.Print("Error shutdown: ", err)
		}
	}()
}
