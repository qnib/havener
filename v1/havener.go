package havener

import (
	"context"
	"fmt"
	"log"

	"github.com/codegangsta/negroni"
	"github.com/codegangsta/cli"
	"github.com/zpatrick/go-config"
)

const (
	dockerAPI = "v1.29"
)

var (
	ctx = context.Background()
)



func Run(appCtx *cli.Context) {
	cfg := config.NewConfig([]config.Provider{config.NewCLI(appCtx, true)})

	bindHost, err := cfg.String("bind-host")
	if err != nil {
		log.Panicf("Error parsing bind-host: %s", err.Error())
	}
	bindPort, err := cfg.Int("bind-port")
	if err != nil {
		log.Panicf("Error parsing bind-port: %s", err.Error())
	}
	addr := fmt.Sprintf("%s:%d", bindHost, bindPort)
	log.Printf("Start Listening on port '%s", addr)
	n := negroni.Classic()
	redirect, _ := cfg.BoolOr("redirect", false)
	switch {
	case redirect:
		mux := RedirectMux(cfg)
		n.UseHandler(mux)
    /*default:
		proxy := NewReverseProxy(cfg)
		n.UseHandler(proxy.Handle)
	*/
	}
	n.Run(addr)
}

