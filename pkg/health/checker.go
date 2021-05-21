package health

import (
	"errors"
	"net"
	"time"

	"github.com/chermehdi/farely/pkg/domain"
	log "github.com/sirupsen/logrus"
)

type HealthChecker struct {
	servers []*domain.Server

	// TODO(chermehdi): configure the period based on the config file.
	period int
}

// NewChecker will create a new HealthChecker.
func NewChecker(_conf *domain.Config, servers []*domain.Server) (*HealthChecker, error) {
	if len(servers) == 0 {
		return nil, errors.New("A server list expected, gotten an empty list")
	}
	return &HealthChecker{
		servers: servers,
	}, nil
}

// Start keeps looping indefinitly try to check the health of every server
// the caller is responsible of creating the goroutine when this should run
func (hc *HealthChecker) Start() {
	log.Info("Starting the health checker...")
	ticker := time.NewTicker(time.Second * 10)
	defer ticker.Stop()
	for {
		select {
		case _ = <-ticker.C:
			for _, server := range hc.servers {
				go checkHealth(server)
			}
		}
	}
}

// changes the liveness of the server (either from live to dead or the other way
// around)
func checkHealth(server *domain.Server) {
	// We will consider a server to be healthy if we can open a tcp connection
	// to the host:port of the server within a reasonable time frame.
	_, err := net.DialTimeout("tcp", server.Url.Host, time.Second*5)
	if err != nil {
		log.Errorf("Could not connect to the server at '%s'", server.Url.Host)
		old := server.SetLiveness(false)
		if old {
			log.Warnf("Transitioning server '%s' from Live to Unavailable state", server.Url.Host)
		}
		return
	}
	old := server.SetLiveness(true)
	if !old {
		log.Infof("Transitioning server '%s' from Unavailable to Live state", server.Url.Host)
	}
}
