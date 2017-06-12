package havener

import (
	"fmt"
	"log"
	"strings"
	"time"
	"math/rand"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/zpatrick/go-config"
	"reflect"
	"sync"
)

type Registry map[string][]string

type SrvRegistry struct {
	Cfg *config.Config
	Cli *client.Client
	mutex sync.Mutex
	Registry Registry
}

type Request struct {
	Stack   string
	Service string
}

func (rq *Request) String() string {
	return rq.Stack+"/"+rq.Service
}
func NewSrvRegistry(cfg *config.Config) SrvRegistry {
	sr := SrvRegistry{
		Cfg: cfg,
		Registry: Registry{},
	}
	sr.ConnectDocker()
	return sr
}

func (sr *SrvRegistry) ConnectDocker() {
	dockerHost, err := sr.Cfg.String("docker-host")
	if err != nil {
		log.Panicf("Error parsing docker-host: %s", err.Error())
	}
	sr.Cli, err = client.NewClient(dockerHost, dockerAPI, nil, nil)
	if err != nil {
		log.Printf("Could not connect docker/docker/client to '%s': %v", dockerHost, err)
		return
	} else {
		info, _ := sr.Cli.Info(ctx)
		log.Printf("Connected to docker-engine 'v%s'", info.ServerVersion)
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
				req <- sr.GetRedirect(rq.String())
			}
		}
	}
}


func (sr *SrvRegistry) BuildRegistry() {
	baseUrl, err := sr.Cfg.String("base-url")
	if err != nil {
		log.Panicf("Error parsing base-url: %s", err.Error())
	}
	LabelPrefix, err := sr.Cfg.String("label-prefix")
	if err != nil {
		log.Panicf("Error parsing label-prefix: %s", err.Error())
	}
	lOpts := types.ServiceListOptions{}
	services, _ := sr.Cli.ServiceList(ctx, lOpts)
	_ = baseUrl
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
	registry := Registry{}
	for key, uCfg := range uriCfg {
		tupel := strings.Split(key, ":")
		srv, port := tupel[0], tupel[1]
		srvUri := srv
		if _, ok := registry[srvUri]; !ok {
			registry[srvUri] = []string{}
		}
		registry[srvUri] = append(registry[srvUri], fmt.Sprintf("%s://%s:%s", uCfg["proto"], baseUrl, port))
	}
	eq := reflect.DeepEqual(registry, sr.Registry)
	if eq {
		return
	}
	sr.mutex.Lock()
	defer sr.mutex.Unlock()
	log.Println("Update registry")
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
