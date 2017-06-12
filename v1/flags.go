package havener

import "github.com/codegangsta/cli"

var Flags = []cli.Flag{
	cli.BoolFlag{
		Name:   "redirect",
		Usage:  "Instead of reverse proxy, redirect the request",
		EnvVar: "HAVENER_REDIRECT",
	},
	cli.StringFlag{
		Name:   "label-prefix",
		Value:  "org.qnib.havener",
		Usage:  "Label prefix.",
		EnvVar: "HAVENER_LABEL_PREFIX",
	},
	cli.StringFlag{
		Name:   "base-url",
		Value:  "localhost",
		Usage:  "Base URL of proxy",
		EnvVar: "HAVENER_BASE_URL",
	},
	cli.StringFlag{
		Name:   "docker-host",
		Value:  "unix:///var/run/docker.sock",
		Usage:  "DOCKER_HOST variable to connect to docker-engine",
		EnvVar: "DOCKER_HOST",
	},
	cli.StringFlag{
		Name:   "bind-host",
		Value:  "0.0.0.0",
		Usage:  "Bind host for proxy",
		EnvVar: "HAVENER_BIND_HOST",
	},
	cli.IntFlag{
		Name:   "bind-port",
		Value:  9090,
		Usage:  "Bind port for proxy",
		EnvVar: "HAVENER_BIND_PORT",
	},

}
