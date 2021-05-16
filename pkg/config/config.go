package config

import (
	"net/http"
	"net/http/httputil"
	"net/url"
	"sync/atomic"
)

type Service struct {
	Name     string   `yaml:"name"`
	Replicas []string `yaml:"replicas"`
}

// Config is a representation of the configuration
// given to farely from a config source.
type Config struct {
	Services []Service `yaml:"services"`

	// Name of the strategy to be used in load balancing between instances
	Strategy string `yaml:"strategy"`
}

// Server is an instance of a running server
type Server struct {
	Url   *url.URL
	Proxy *httputil.ReverseProxy
}

func (s *Server) Forward(res http.ResponseWriter, req *http.Request) {
	s.Proxy.ServeHTTP(res, req)
}

type ServerList struct {
	Servers []*Server

	// The current server to forward the request to.
	// the next server should be (current + 1) % len(Servers)
	current uint32
}

func (sl *ServerList) Next() uint32 {
	nxt := atomic.AddUint32(&sl.current, uint32(1))
	lenS := uint32(len(sl.Servers))
	return nxt % lenS
}
