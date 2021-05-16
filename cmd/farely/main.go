package main

import (
	"flag"
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"

	"github.com/chermehdi/farely/pkg/config"
	log "github.com/sirupsen/logrus"
)

var (
	port       = flag.Int("port", 8080, "where to start farely")
	configFile = flag.String("config-path", "", "The config file to supply to farely")
)

type Farely struct {
	Config     *config.Config
	ServerList *config.ServerList
}

func NewFarely(conf *config.Config) *Farely {
	servers := make([]*config.Server, 0)
	for _, service := range conf.Services {
		// TODO(chermehdi): Don't ignore the names
		for _, replica := range service.Replicas {
			ur, err := url.Parse(replica)
			if err != nil {
				log.Fatal(err)
			}
			proxy := httputil.NewSingleHostReverseProxy(ur)
			servers = append(servers, &config.Server{
				Url:   ur,
				Proxy: proxy,
			})
		}
	}
	return &Farely{
		Config: conf,
		ServerList: &config.ServerList{
			Servers: servers,
		},
	}
}

func (f *Farely) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	// TODO(chermehdi): We need to support per service forwarding, i.e this method
	// should read the request path, say host:port/serice/rest/of/url this should
	// be load balanced against service named "service" and url will be
	// "host{i}:port{i}/rest/of/url
	log.Infof("Received new request: url='%s'", req.Host)

	next := f.ServerList.Next()
	log.Infof("Forwarding to the server number='%d'", next)
	// Forwarding the request to the proxy
	f.ServerList.Servers[next].Forward(res, req)
}

func main() {
	flag.Parse()
	file, err := os.Open(*configFile)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()
	conf, err := config.LoadConfig(file)

	if err != nil {
		log.Fatal(err)
	}

	farely := NewFarely(conf)

	server := http.Server{
		Addr:    fmt.Sprintf(":%d", *port),
		Handler: farely,
	}

	if err := server.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}
