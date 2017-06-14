package havener

import (
	"log"
	"net/url"
	"strings"
	"fmt"
	"net/http"
	"net/http/httputil"
	"github.com/zpatrick/go-config"

)

type ReverseProxy struct {
	Cfg *config.Config
	RegQuery chan interface{}
}

func NewReverseProxy(cfg *config.Config, rq chan interface{}) ReverseProxy {
	rp := ReverseProxy{
		Cfg: cfg,
		RegQuery: rq,
	}
	return rp
}

// NewMultipleHostReverseProxy creates a reverse proxy that will randomly
// select a host from the passed `targets`
func (rp *ReverseProxy) Handle() *httputil.ReverseProxy {
	director := func(req *http.Request) {
		stack, service, err := extractNameVersion(req.URL)
		if err != nil {
			log.Print(err)
			return
		}
		endpoint := rp.GetForward(stack, service)
		if endpoint == "" {
			log.Printf("'%s/%s' not found", stack, service)
			return
		} else {
			log.Printf("'%s/%s' forwards to '%s", stack, service, endpoint)
		}
		req.URL.Scheme = "http"
		req.URL.Host = endpoint
	}
	return &httputil.ReverseProxy{
		Director: director,
	}
}


func (rp *ReverseProxy) GetForward(stack, service string) string {
	rq := NewSrvRequest(stack, service)
	rp.RegQuery <- rq
	key := <- rp.RegQuery
	return key.(string)

}

func extractNameVersion(target *url.URL) (name, version string, err error) {
	path := target.Path
	// Trim the leading `/`
	if len(path) > 1 && path[0] == '/' {
		path = path[1:]
	}
	// Explode on `/` and make sure we have at least
	// 2 elements (service name and version)
	tmp := strings.Split(path, "/")
	if len(tmp) < 2 {
		return "", "", fmt.Errorf("Invalid path")
	}
	name, version = tmp[0], tmp[1]
	// Rewrite the request's path without the prefix.
	target.Path = "/" + strings.Join(tmp[2:], "/")
	return name, version, nil
}
