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
	rq := NewSrvRequest(stack, service)
	rh.RegQuery <- rq
	val := <- rh.RegQuery
	switch val.(type) {
	case string:
		return val.(string)
	}
	return ""

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
	si := NewSrvIndex(cfg, rq)
	mux := bone.New()
	mux.Get("/", http.HandlerFunc(si.Handler))
	mux.Get("/static/css/bootstrap.min.css",  http.Handler(http.StripPrefix("/static/css", http.FileServer(http.Dir("/usr/share/havener/static/css")))))
	mux.Get("/static/js/bootstrap.min.js",  http.Handler(http.StripPrefix("/static/js", http.FileServer(http.Dir("/usr/share/havener/static/js")))))
	mux.Get("/static/js/jquery.min.js",  http.Handler(http.StripPrefix("/static/js", http.FileServer(http.Dir("/usr/share/havener/static/js")))))
	mux.Get("/:stack/:service", http.HandlerFunc(rh.Handler))
	return mux
}