package domain

import (
	"net/http"
	"net/http/httputil"
	"net/url"
)

type Service struct {
	Name string `yaml:"name"`

	// A prefix matcher to select service based on the path part of the url
	// Note(self): The matcher could be more sophisticated (i.e Regex based,
	// subdomain based), but for the purposes of simplicity let's think about this
	// later, and it could be a nice contribution to the project.
	Matcher string `yaml:"matcher"`

	// Strategy is the load balancing strategy used for this service.
	Strategy string `yaml:"strategy"`

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
