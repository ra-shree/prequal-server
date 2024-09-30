package common

import (
	"fmt"
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

func (t *Replica) RemoveUpstream(faultyUpstream *url.URL) {
	for i, upstream := range t.Upstreams {
		if upstream.String() == faultyUpstream.String() {
			t.Upstreams = append(t.Upstreams[:i], t.Upstreams[i+1:]...)
			fmt.Printf("Removed faulty upstream: %s\n", faultyUpstream.String())
			return
		}
	}
}
