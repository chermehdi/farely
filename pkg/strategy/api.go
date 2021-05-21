package strategy

import (
	"sync"
	"sync/atomic"

	"github.com/chermehdi/farely/pkg/domain"
	log "github.com/sirupsen/logrus"
)

// Known load balancing strategies, each entry in this block should correspond
// to a load balancing strategy with a concrete implementation.
const (
	kRoundRobin         = "RoundRobin"
	kWeightedRoundRobin = "WeightedRoundRobin"
	kUnknown            = "Unknown"
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
	strategies[kRoundRobin] = func() BalancingStrategy {
		return &RoundRobin{current: uint32(0)}
	}
	strategies[kWeightedRoundRobin] = func() BalancingStrategy {
		return &WeightedRoundRobin{mu: sync.Mutex{}}
	}
	// Add other load balancing strategies here
}

type RoundRobin struct {
	// The current server to forward the request to.
	// the next server should be (current + 1) % len(Servers)
	current uint32
}

func (r *RoundRobin) Next(servers []*domain.Server) (*domain.Server, error) {
	nxt := atomic.AddUint32(&r.current, uint32(1))
	lenS := uint32(len(servers))
	picked := servers[nxt%lenS]
	log.Infof("Strategy picked server '%s'", picked.Url.Host)
	return picked, nil
}

// WeightedRoundRobin is a strategy that is similar to the RoundRobin strategy,
// the only difference is that it takes server compute power into consideration.
// The compute power of a server is given as an integer, it represents the
// fraction of requests that one server can handle over another.
//
// A RoundRobin strategy is equivalent to a WeightedRoundRobin strategy with all
// weights = 1
type WeightedRoundRobin struct {
	// Any changes to the below field should only be done while holding the `mu`
	// lock.
	mu sync.Mutex
	// Note: This is making the assumption that the server list coming through the
	// Next function won't change between succesive calls.
	// Changing the server list would cause this strategy to break, panic, or not
	// route properly.
	//
	// count will keep track of the number of request server `i` processed.
	count []int
	// cur is the index of the last server that executed a request.
	cur int
}

func (r *WeightedRoundRobin) Next(servers []*domain.Server) (*domain.Server, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.count == nil {
		// First time using the strategy
		r.count = make([]int, len(servers))
		r.cur = 0
	}
	capacity := servers[r.cur].GetMetaOrDefaultInt("weight", 1)
	if r.count[r.cur] <= capacity {
		r.count[r.cur] += 1
		log.Infof("Strategy picked server '%s'", servers[r.cur].Url.Host)
		return servers[r.cur], nil
	}

	// server is at it's limit, reset the current one
	// and move on to the next server
	r.count[r.cur] = 0
	r.cur = (r.cur + 1) % len(servers)
	log.Infof("Strategy picked server '%s'", servers[r.cur].Url.Host)
	return servers[r.cur], nil
}

// LoadStrategy will try and resolve the balancing strategy based on the name,
// and will default to a 'RoundRobin' one if no strategy matched.
func LoadStrategy(name string) BalancingStrategy {
	st, ok := strategies[name]
	if !ok {
		log.Warnf("Strategy with name '%s' not found, falling back to a RoundRobin strategy", name)
		return strategies[kRoundRobin]()
	}
	log.Infof("Picked strategy '%s'", name)
	return st()
}
