package havener

import (
	"net/http"
	"github.com/go-zoo/bone"
	"github.com/zpatrick/go-config"
	"fmt"
)

type Redirector struct {
	Cfg *config.Config
	RegQuery chan interface{}
}

func NewRedirector(cfg *config.Config, rq chan interface{}) Redirector {
	rd := Redirector{
		Cfg: cfg,
		RegQuery: rq,
	}
	return rd
}

func (rh *Redirector) GetRedirect(r *http.Request) string {
	stack := bone.GetValue(r, "stack")
	service := bone.GetValue(r, "service")
	rq := Request{Stack: stack, Service: service}
	rh.RegQuery <- rq
	key := <- rh.RegQuery
	return key.(string)

}

func (rh *Redirector) Handler(w http.ResponseWriter, r *http.Request) {
	url := rh.GetRedirect(r)
	if url != "" {
		http.Redirect(w, r, url, 301)
	} else {
		msg := fmt.Sprintf("No such stack/service '%s/%s' in SWARM cluster", bone.GetValue(r, "stack"), bone.GetValue(r,"service"))
		http.Error(w, msg, 404)
	}

}

func RedirectMux(cfg *config.Config, rq chan interface{}) *bone.Mux {
	rh := NewRedirector(cfg, rq)
	mux := bone.New()
	mux.Get("/:stack/:service", http.HandlerFunc(rh.Handler))
	return mux
}