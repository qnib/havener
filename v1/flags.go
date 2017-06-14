package havener

import "github.com/codegangsta/cli"

var Flags = []cli.Flag{
	cli.BoolFlag{
		Name:   "redirect-disable",
		Usage:  "Disable redirect server",
		EnvVar: "HAVENER_REDIRECT_DISABLE",
	},
	cli.BoolFlag{
		Name:   "debug",
		Usage:  "Enable debugging",
		EnvVar: "HAVENER_DEBUG",
	},
	cli.BoolFlag{
		Name:   "proxy-disable",
		Usage:  "Disable proxy server",
		EnvVar: "HAVENER_PROXY_DISABLE",
	},
	cli.StringFlag{
		Name:   "label-prefix",
		Value:  "org.qnib.havener",
		Usage:  "Label prefix.",
		EnvVar: "HAVENER_LABEL_PREFIX",
	},
	cli.StringFlag{
		Name:   "base-url",
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
	cli.StringFlag{
		Name:   "skip-service",
		Usage:  "Bind host for proxy",
		EnvVar: "HAVENER_SKIP_SERVICE",
	},
	cli.IntFlag{
		Name:   "redirect-port",
		Value:  9090,
		Usage:  "Bind port for redirect server",
		EnvVar: "HAVENER_REDIRECT_PORT",
	},
	cli.IntFlag{
		Name:   "proxy-port",
		Value:  9091,
		Usage:  "Bind port for proxy server",
		EnvVar: "HAVENER_PROXY_PORT",
	},
	cli.IntFlag{
		Name:   "service-tick-ms",
		Value:  2000,
		Usage:  "Update interval of service-registry",
		EnvVar: "HAVENER_SERVICE_TICK_MS",
	},
}
