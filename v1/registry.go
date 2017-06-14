package havener

import (
	"fmt"
	"log"
	"strings"
	"time"
	"math/rand"
	"reflect"
	"sync"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/zpatrick/go-config"
)

type Registry map[string][]string

type SrvRegistry struct {
	Cfg *config.Config
	Cli *client.Client
	mutex sync.Mutex
	Registry Registry
	UriCfg map[string]map[string]string
	info types.Info
}

type Request struct {
	Mode 	string
	Stack   string
	Service string
}

func NewSrvRequest(stack, service string) Request {
	return Request{
		Mode: "service",
		Stack: stack,
		Service: service,
	}
}

func NewRegistryRequest() Request {
	return Request{
		Mode: "registry",
	}
}


func (rq *Request) String() string {
	return rq.Stack+"/"+rq.Service
}
func NewSrvRegistry(cfg *config.Config) SrvRegistry {
	sr := SrvRegistry{
		Cfg: cfg,
		Registry: Registry{},
		UriCfg: map[string]map[string]string{},
	}
	sr.ConnectDocker()
	return sr
}

func (sr *SrvRegistry) Log(level, msg string) {
	debug, _ := sr.Cfg.BoolOr("debug", false)
	level = strings.ToUpper(level)
	switch level {
	case "DEBUG":
		if debug {
			log.Printf("%s # %s", level, msg)
		}
	default:
		log.Printf("%s # %s", level, msg)
	}
}

func (sr *SrvRegistry) ConnectDocker() {
	dockerHost, err := sr.Cfg.String("docker-host")
	if err != nil {
		log.Panicf("Error parsing docker-host: %s", err.Error())
	}
	sr.Cli, err = client.NewClient(dockerHost, dockerAPI, nil, nil)
	if err != nil {
		sr.Log("error", fmt.Sprintf("Could not connect docker/docker/client to '%s': %v", dockerHost, err))
		return
	} else {
		sr.info, _ = sr.Cli.Info(ctx)
		sr.Log("info", fmt.Sprintf("Connected to docker-engine 'v%s'", sr.info.ServerVersion))
	}

}

func (sr *SrvRegistry) ChanHandler(req chan interface{}) {
	tickMs, _ := sr.Cfg.Int("service-tick-ms")
	ticker := time.NewTicker(time.Duration(tickMs) * time.Millisecond).C
	sr.BuildRegistry()
	for {
		select {
		case <-ticker:
			sr.BuildRegistry()
		case r := <-req:
			switch r.(type) {
			case Request:
				rq := r.(Request)
				switch rq.Mode {
				case "service":
					req <- sr.GetRedirect(rq.String())
				case "registry":
					req <- sr.GetRegistry()
				}
			}
		}
	}
}


func (sr *SrvRegistry) BuildRegistry() {
	baseUrl, err := sr.Cfg.StringOr("base-url", sr.info.Swarm.NodeAddr)
	if err != nil {
		sr.Log("error", fmt.Sprintf("Error parsing base-url: %s", err.Error()))
	}
	LabelPrefix, err := sr.Cfg.String("label-prefix")
	if err != nil {
		sr.Log("error", fmt.Sprintf("Error parsing label-prefix: %s", err.Error()))
	}
	lOpts := types.ServiceListOptions{}
	services, _ := sr.Cli.ServiceList(ctx, lOpts)
	uriCfg := map[string]map[string]string{}
	for _, srv := range services {
		skipSrv, err := sr.Cfg.String("skip-service")
		if err == nil && skipSrv == srv.Spec.Name {
			sr.Log("debug", fmt.Sprintf("Skip service %s", srv.Spec.Name))
			continue
		}
		srvName := strings.Replace(srv.Spec.Name, "_", "/", 1)
		for _, ports := range srv.Spec.EndpointSpec.Ports {
			key := fmt.Sprintf("%s:%d", srvName, ports.PublishedPort)
			if _, ok := uriCfg[key]; !ok {
				uriCfg[key] = map[string]string{
					"proto": "http",
					"uri":   "",
				}
			}
		}
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
					sr.Log("error", fmt.Sprintf("expected <int>[.(uri|proto)]: %v", tupel))
				}

				key := fmt.Sprintf("%s:%s", srvName, port)
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
	registry := Registry{}
	for key, uCfg := range uriCfg {
		tupel := strings.Split(key, ":")
		srv, port := tupel[0], tupel[1]
		srvUri := srv
		if _, ok := registry[srvUri]; !ok {
			registry[srvUri] = []string{}
		}
		uri := fmt.Sprintf("%s://%s:%s", uCfg["proto"], baseUrl, port)
		msg :=  fmt.Sprintf("Service: %s -> %s", key, uri)
		sr.Log("debug", msg)
		registry[srvUri] = append(registry[srvUri], uri)
	}
	sr.mutex.Lock()
	defer sr.mutex.Unlock()
	sr.UriCfg = uriCfg
	eq := reflect.DeepEqual(registry, sr.Registry)
	if len(sr.Registry) != 0 && eq {
		return
	}
	sr.Log("debug", "Update registry")
	if len(registry) == 0 {
		sr.Log("warn", "Registry is empty...?")

	}
	sr.Registry = registry
}

func (sr *SrvRegistry) GetRedirect(key string) string {
	sr.mutex.Lock()
	defer sr.mutex.Unlock()
	if rg, ok := sr.Registry[key]; ok {
		return rg[rand.Int()%len(sr.Registry[key])]
	}
	return ""
}

func (sr *SrvRegistry) GetRegistry() Registry {
	sr.mutex.Lock()
	defer sr.mutex.Unlock()
	reg := Registry{}
	for k, cfg := range sr.UriCfg {
		tupel := strings.Split(k, ":")
		srv, port := tupel[0], tupel[1]
		uri := fmt.Sprintf("%s://%s:%s", cfg["proto"],  sr.info.Swarm.NodeAddr, port)
		if _, ok := reg[srv]; !ok {
			reg[srv] = []string{}
		}
		switch {
		case cfg["proto"] == "http":
			reg[srv] = append(reg[srv], uri)
		case cfg["proto"] == "https":
			reg[srv] = append(reg[srv], uri)
		}
	}
	return reg
}
