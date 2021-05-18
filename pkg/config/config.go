package config

import (
	"github.com/chermehdi/farely/pkg/domain"
	"github.com/chermehdi/farely/pkg/strategy"
)

// Config is a representation of the configuration
// given to farely from a config source.
type Config struct {
	Services []domain.Service `yaml:"services"`

	// TODO(chermehdi): remove this.
	// Name of the strategy to be used in load balancing between instances
	Strategy string `yaml:"strategy"`
}

type ServerList struct {
	// Servers are the replicas
	Servers []*domain.Server

	// Name of the service
	Name string

	// Strategy defines how the server list is load balanced.
	// It can never be 'nil', it should always default to a 'RoundRobin' version.
	Strategy strategy.BalancingStrategy
}
