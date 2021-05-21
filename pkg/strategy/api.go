package strategy

import (
	"errors"
	"fmt"
	"sync"

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
		return &RoundRobin{
			mu:      sync.Mutex{},
			current: 0,
		}
	}
	strategies[kWeightedRoundRobin] = func() BalancingStrategy {
		return &WeightedRoundRobin{mu: sync.Mutex{}}
	}
	// Add other load balancing strategies here
}

type RoundRobin struct {
	// The current server to forward the request to.
	// the next server should be (current + 1) % len(Servers)
	mu      sync.Mutex
	current int
}

func (r *RoundRobin) Next(servers []*domain.Server) (*domain.Server, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	seen := 0
	var picked *domain.Server
	for seen < len(servers) {
		picked = servers[r.current]
		r.current = (r.current + 1) % len(servers)
		if picked.IsAlive() {
			break
		}
		seen += 1
	}
	if picked == nil || seen == len(servers) {
		log.Error("All servers are down")
		return nil, errors.New(fmt.Sprintf("Checked all the '%d' servers, none of them is available", seen))
	}
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

	seen := 0
	var picked *domain.Server
	for seen < len(servers) {
		picked = servers[r.cur]
		capacity := picked.GetMetaOrDefaultInt("weight", 1)
		if !picked.IsAlive() {
			seen += 1
			// Current server is not alive, so we reset the server's bucket count
			// and we try the next server in the next loop iteration
			r.count[r.cur] = 0
			r.cur = (r.cur + 1) % len(servers)
			continue
		}

		if r.count[r.cur] <= capacity {
			r.count[r.cur] += 1
			log.Infof("Strategy picked server '%s'", picked.Url.Host)
			return picked, nil
		}
		// server is at it's limit, reset the current one
		// and move on to the next server
		r.count[r.cur] = 0
		r.cur = (r.cur + 1) % len(servers)
	}

	if picked == nil || seen == len(servers) {
		log.Error("All servers are down")
		return nil, errors.New(fmt.Sprintf("Checked all the '%d' servers, none of them is available", seen))
	}

	return picked, nil
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
