package reverseproxy

import (
	"net/url"
	"sync"

	"github.com/gorilla/mux"
)

// represents a replica of the backend server that requests are forwarded to
type Replica struct {
	router       *mux.Router
	upstreams    []*url.URL
	lastUpstream int
	lock         sync.Mutex
}

func (t *Replica) SelectUpstream() *url.URL {
	count := len(t.upstreams)
	if count == 1 {
		return t.upstreams[0]
	}

	t.lock.Lock()
	defer t.lock.Unlock()

	next := t.lastUpstream + 1
	if next >= count {
		next = 0
	}

	t.lastUpstream = next

	return t.upstreams[next]
}
