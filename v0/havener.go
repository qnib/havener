package havener

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"

	"github.com/codegangsta/cli"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/zpatrick/go-config"
)

const (
	dockerAPI = "v1.29"
)

var (
	ctx = context.Background()
)

/**** http://blog.charmes.net/2015/07/reverse-proxy-in-go.html
 */

type Registry map[string][]string

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

// NewMultipleHostReverseProxy creates a reverse proxy that will randomly
// select a host from the passed `targets`
func NewMultipleHostReverseProxy(reg Registry) *httputil.ReverseProxy {
	director := func(req *http.Request) {
		name, version, err := extractNameVersion(req.URL)
		if err != nil {
			log.Print(err)
			return
		}
		endpoints := reg[name+"/"+version]
		if len(endpoints) == 0 {
			log.Printf("Service/Version not found")
			return
		}
		req.URL.Scheme = "http"
		req.URL.Host = endpoints[rand.Int()%len(endpoints)]
	}
	return &httputil.ReverseProxy{
		Director: director,
	}
}

func Run(appCtx *cli.Context) {
	cfg := config.NewConfig([]config.Provider{config.NewCLI(appCtx, true)})
	baseUrl, err := cfg.String("base-url")
	if err != nil {
		log.Panicf("Error parsing base-url: %s", err.Error())
	}
	dockerHost, err := cfg.String("docker-host")
	if err != nil {
		log.Panicf("Error parsing docker-host: %s", err.Error())
	}
	bindHost, err := cfg.String("bind-host")
	if err != nil {
		log.Panicf("Error parsing bind-host: %s", err.Error())
	}
	bindPort, err := cfg.Int("bind-port")
	if err != nil {
		log.Panicf("Error parsing bind-port: %s", err.Error())
	}
	LabelPrefix, err := cfg.String("label-prefix")
	if err != nil {
		log.Panicf("Error parsing label-prefix: %s", err.Error())
	}
	cli, err := client.NewClient(dockerHost, dockerAPI, nil, nil)
	if err != nil {
		log.Printf("Could not connect docker/docker/client to '%s': %v", dockerHost, err)
		return
	} else {
		info, _ := cli.Info(ctx)
		log.Printf("Connected to docker-engine 'v%s'", info.ServerVersion)
	}

	lOpts := types.ServiceListOptions{}
	services, _ := cli.ServiceList(ctx, lOpts)
	_ = baseUrl
	registry := Registry{}
	uriCfg := map[string]map[string]string{}
	for _, srv := range services {
		for k, v := range srv.Spec.TaskTemplate.ContainerSpec.Labels {
			pPre := LabelPrefix + ".port"
			if strings.HasPrefix(k, pPre) {
				sufKey := k[len(pPre)+1:]
				tupel := strings.Split(sufKey, ".")
				var suffix, port string
				switch len(tupel) {
				case 1:
					port = tupel[0]
				case 2:
					port, suffix = tupel[0], tupel[1]
				default:
					log.Panicf("expected <int>[.(uri|proto)]: %v", tupel)
				}

				key := strings.Replace(srv.Spec.Name, "_", "/", 1) + ":" + port
				if _, ok := uriCfg[key]; !ok {
					uriCfg[key] = map[string]string{
						"proto": "http",
						"uri":   "",
					}
				}
				if suffix != "" {
					uriCfg[key][suffix] = v
				}
			}
		}
	}
	for key, uCfg := range uriCfg {
		_ = uCfg
		tupel := strings.Split(key, ":")
		srv, port := tupel[0], tupel[1]
		srvUri := srv //+ uCfg["uri"]
		log.Printf("Add URI '%s' -> '%s:%s'", srvUri, baseUrl, port)
		if _, ok := registry[srvUri]; !ok {
			registry[srvUri] = []string{}
		}
		registry[srvUri] = append(registry[srvUri], fmt.Sprintf("%s:%s", baseUrl, port))
	}
	proxy := NewMultipleHostReverseProxy(registry)
	addr := fmt.Sprintf("%s:%d", bindHost, bindPort)
	log.Printf("Start Listening on port '%s", addr)
	log.Fatal(http.ListenAndServe(addr, proxy))
}
