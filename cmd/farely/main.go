package main

import (
	"flag"
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strings"

	"github.com/chermehdi/farely/pkg/config"
	"github.com/chermehdi/farely/pkg/domain"
	"github.com/chermehdi/farely/pkg/strategy"
	log "github.com/sirupsen/logrus"
)

var (
	port       = flag.Int("port", 8080, "where to start farely")
	configFile = flag.String("config-path", "", "The config file to supply to farely")
)

type Farely struct {
	// Config is the configuration loaded from a config file
	// TODO(chermehdi): This could be improved, as to fetch the configuration from
	// a more abstract concept (like ConfigSource) that can either be a file or
	// something else, and also should support hot reloading.
	Config *config.Config

	// ServerList will contain a mapping between matcher and replicas
	ServerList map[string]*config.ServerList
}

func NewFarely(conf *config.Config) *Farely {
	// TODO(chermehdi): prevent multiple or invalid matchers before creating the
	// server
	serverMap := make(map[string]*config.ServerList, 0)

	for _, service := range conf.Services {
		servers := make([]*domain.Server, 0)
		for _, replica := range service.Replicas {
			ur, err := url.Parse(replica.Url)
			if err != nil {
				log.Fatal(err)
			}
			proxy := httputil.NewSingleHostReverseProxy(ur)
			servers = append(servers, &domain.Server{
				Url:   ur,
				Proxy: proxy,
			})
		}
		serverMap[service.Matcher] = &config.ServerList{
			Servers:  servers,
			Name:     service.Name,
			Strategy: strategy.LoadStrategy(service.Strategy),
		}
	}
	return &Farely{
		Config:     conf,
		ServerList: serverMap,
	}
}

// Looks for the first server list that matches the reqPath (i.e matcher)
// Will return an error if no matcher have been found.
// TODO(chermehdi): Does it make sense to allow default responders?
func (f *Farely) findServiceList(reqPath string) (*config.ServerList, error) {
	log.Infof("Trying to find matcher for request '%s'", reqPath)
	for matcher, s := range f.ServerList {
		if strings.HasPrefix(reqPath, matcher) {
			log.Infof("Found service '%s' matching the request", s.Name)
			return s, nil
		}
	}
	return nil, fmt.Errorf("Could not find a matcher for url: '%s'", reqPath)
}

func (f *Farely) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	// TODO(chermehdi): We need to support per service forwarding, i.e this method
	// should read the request path, say host:port/serice/rest/of/url this should
	// be load balanced against service named "service" and url will be
	// "host{i}:port{i}/rest/of/url
	log.Infof("Received new request: url='%s'", req.Host)
	sl, err := f.findServiceList(req.URL.Path)
	if err != nil {
		log.Error(err)
		res.WriteHeader(http.StatusNotFound)
		return
	}

	next, err := sl.Strategy.Next(sl.Servers)

	if err != nil {
		log.Error(err)
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	log.Infof("Forwarding to the server='%s'", next.Url.RawPath)
	next.Forward(res, req)
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
