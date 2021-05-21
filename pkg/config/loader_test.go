package config

import (
	"strings"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	conf, err := LoadConfig(strings.NewReader(`
services:
- name: "test service" 
  strategy: RoundRobin 
  matcher: "/api/v1"
  replicas:
  - url: localhost:8081
  - url: localhost:8082
strategy: RoundRobin
`))
	if err != nil {
		t.Errorf("Error should be nil: '%s'", err)
	}
	if conf.Strategy != "RoundRobin" {
		t.Errorf("Strategy expected to equal 'RoundRobin' got '%s' instead", conf.Strategy)
	}
	if len(conf.Services) != 1 {
		t.Errorf("Expected service count to be 1 got '%d'", len(conf.Services))
	}
	if conf.Services[0].Matcher != "/api/v1" {
		t.Errorf("Expected the matcher to be '/api/v1' go '%s' instead", conf.Services[0].Matcher)
	}
	if conf.Services[0].Name != "test service" {
		t.Errorf("Expected service name to be equal to 'test service' got '%s'", conf.Services[0].Name)
	}
	if conf.Services[0].Strategy != "RoundRobin" {
		t.Errorf("Expected service name to be equal to 'RoundRobin' got '%s'", conf.Services[0].Strategy)
	}
	if len(conf.Services[0].Replicas) != 2 {
		t.Errorf("Expected replica count to be 2 got '%d'", len(conf.Services[0].Replicas))
	}
	if conf.Services[0].Replicas[0].Url != "localhost:8081" {
		t.Errorf("Expected first replica to be 'locahost:8081', got '%s'", conf.Services[0].Replicas[0])
	}
	if conf.Services[0].Replicas[1].Url != "localhost:8082" {
		t.Errorf("Expected second replica to be 'locahost:8082', got '%s'", conf.Services[0].Replicas[1])
	}
}

func TestLoadConfigWeightedRoundRobin(t *testing.T) {
	conf, err := LoadConfig(strings.NewReader(`
services:
- name: "test service" 
  strategy: WeightedRoundRobin 
  matcher: "/api/v1"
  replicas:
  - url: localhost:8081
    metadata:
      weight: 10
  - url: localhost:8082
    metadata:
      weight: 5 
strategy: RoundRobin
`))
	if err != nil {
		t.Errorf("Error should be nil: '%s'", err)
	}
	if conf.Strategy != "RoundRobin" {
		t.Errorf("Strategy expected to equal 'RoundRobin' got '%s' instead", conf.Strategy)
	}
	if len(conf.Services) != 1 {
		t.Errorf("Expected service count to be 1 got '%d'", len(conf.Services))
	}
	if conf.Services[0].Matcher != "/api/v1" {
		t.Errorf("Expected the matcher to be '/api/v1' go '%s' instead", conf.Services[0].Matcher)
	}
	if conf.Services[0].Name != "test service" {
		t.Errorf("Expected service name to be equal to 'test service' got '%s'", conf.Services[0].Name)
	}
	if conf.Services[0].Strategy != "WeightedRoundRobin" {
		t.Errorf("Expected service name to be equal to 'RoundRobin' got '%s'", conf.Services[0].Strategy)
	}
	if len(conf.Services[0].Replicas) != 2 {
		t.Errorf("Expected replica count to be 2 got '%d'", len(conf.Services[0].Replicas))
	}
	if conf.Services[0].Replicas[0].Url != "localhost:8081" {
		t.Errorf("Expected first replica to be 'locahost:8081', got '%s'", conf.Services[0].Replicas[0])
	}
	if conf.Services[0].Replicas[1].Url != "localhost:8082" {
		t.Errorf("Expected second replica to be 'locahost:8082', got '%s'", conf.Services[0].Replicas[1])
	}
}
