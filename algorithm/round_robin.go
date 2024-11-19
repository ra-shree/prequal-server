package algorithm

import (
	"net/url"

	"github.com/ra-shree/prequal-server/load-balancer/common"
)

var lastUpstream int

func RoundRobin(r *common.Replica) *url.URL {
	count := len(r.Upstreams)
	if count == 1 {
		return r.Upstreams[0]
	}

	r.Lock.Lock()
	defer r.Lock.Unlock()

	next := lastUpstream + 1
	if next >= count {
		next = 0
	}

	lastUpstream = next

	return r.Upstreams[next]
}
