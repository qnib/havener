package main

import (
	"os"
	"github.com/codegangsta/cli"
	"github.com/qnib/havener/v1"
)

func main() {
	app := cli.NewApp()
	app.Name = "Connection Proxy for Docker SWARM"
	app.Usage = "havener [options]"
	app.Version = havener.Version
	app.Author = "Christian Kniep"
	app.Email = "christian@qnib.org"
	app.Flags = havener.Flags
	app.Action = havener.Run
	app.Run(os.Args)
}
