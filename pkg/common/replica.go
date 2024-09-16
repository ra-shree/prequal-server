package common

import (
	"net/url"
	"sync"

	"github.com/gorilla/mux"
)

// represents a replica of the backend server that requests are forwarded to
type Replica struct {
	Router    *mux.Router
	Upstreams []*url.URL
	Lock      sync.Mutex
}

// Any algorithm for selecting an upstream server needs to match this signature
type SelectionAlgorithm func(*Replica) *url.URL

func (t *Replica) SelectUpstream(upstreamSelector SelectionAlgorithm) *url.URL {
	upstream := upstreamSelector(t)

	return upstream
}
