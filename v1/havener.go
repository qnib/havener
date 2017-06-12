package havener

import (
	"context"
	"fmt"
	"log"

	"github.com/codegangsta/negroni"
	"github.com/codegangsta/cli"
	"github.com/zpatrick/go-config"
	"net/http"
	"sync"
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
	redirectPort, err := cfg.Int("redirect-port")
	if err != nil {
		log.Panicf("Error parsing redirect-port: %s", err.Error())
	}
	rdAddr := fmt.Sprintf("%s:%d", bindHost, redirectPort)
	proxyPort, err := cfg.Int("proxy-port")
	if err != nil {
		log.Panicf("Error parsing redirect-port: %s", err.Error())
	}
	proxyAddr := fmt.Sprintf("%s:%d", bindHost, proxyPort)
	n := negroni.Classic()
	rdDisable, _ := cfg.BoolOr("redirect-disable", false)
	proxyDisable, _ := cfg.BoolOr("proxy-disable", false)
	rq := make(chan interface{})
	reg := NewSrvRegistry(cfg)
	go reg.ChanHandler(rq)
	var wg sync.WaitGroup
	if ! rdDisable {
		mux := RedirectMux(cfg, rq)
		n.UseHandler(mux)
		log.Printf("Start Listening on port '%s", rdAddr)
		wg.Add(1)
		go n.Run(rdAddr)
	}
	if ! proxyDisable {
		proxy := NewReverseProxy(cfg, rq)
		log.Printf("Start Listening on port '%s", proxyAddr)
		wg.Add(1)
		go http.ListenAndServe(proxyAddr, proxy.Handle())
	}
	wg.Wait()
}
