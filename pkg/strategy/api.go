package strategy

import (
	"github.com/chermehdi/farely/pkg/domain"
	log "github.com/sirupsen/logrus"
	"sync/atomic"
)

// Known load balancing strategies, each entry in this block should correspond
// to a load balancing strategy with a concrete implementation.
const (
	RoundRobin         = "RoundRobin"
	WeightedRoundRobin = "WeightedRoundRobin"
	Unknown            = "Unknown"
)

// BalancingStrategy is the load balancing abstraction that every algorithm
// should implement.
type BalancingStrategy interface {
	Next([]*domain.Server) (*domain.Server, error)
}

// Map of BalancingStrategy factories
var strategies map[string]func() BalancingStrategy

func init() {
	strategies = make(map[string]func() BalancingStrategy, 0)
	strategies[RoundRobin] = func() BalancingStrategy {
		return &RoundRobing{current: uint32(0)}
	}
	// Add other load balancing strategies here
}

type RoundRobing struct {
	// The current server to forward the request to.
	// the next server should be (current + 1) % len(Servers)
	current uint32
}

func (r *RoundRobing) Next(servers []*domain.Server) (*domain.Server, error) {
	nxt := atomic.AddUint32(&r.current, uint32(1))
	lenS := uint32(len(servers))
	picked := servers[nxt%lenS]
	log.Infof("Strategy picked server '%s'", picked.Url.Host)
	return picked, nil
}

// LoadStrategy will try and resolve the balancing strategy based on the name,
// and will default to a 'RoundRobin' one if no strategy matched.
func LoadStrategy(name string) BalancingStrategy {
	st, ok := strategies[name]
	if !ok {
		return strategies[RoundRobin]()
	}
	return st()
}
