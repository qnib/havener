package havener

import (
	"net/http"
	"github.com/go-zoo/bone"
	"github.com/zpatrick/go-config"
)

type Redirector struct {
	Cfg *config.Config
	Reg SrvRegistry
}

func NewRedirector(cfg *config.Config) Redirector {
	reg := NewSrvRegistry(cfg)
	return Redirector{
		Cfg: cfg,
		Reg: reg,
	}
}

func (rh *Redirector) GetRedirect(r *http.Request) string {
	stack := bone.GetValue(r, "stack")
	service := bone.GetValue(r, "service")
	return rh.Reg.GetRedirect(stack+"/"+service)
}

func (rh *Redirector) Handler(w http.ResponseWriter, r *http.Request) {
	url := rh.GetRedirect(r)
	http.Redirect(w, r, url, 301)
}

func RedirectMux(cfg *config.Config) *bone.Mux {
	rh := NewRedirector(cfg)
	rh.Reg.BuildRegistry()
	mux := bone.New()
	mux.Get("/:stack/:service", http.HandlerFunc(rh.Handler))
	return mux
}