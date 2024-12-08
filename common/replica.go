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
	Lock      sync.RWMutex
}

// Any algorithm for selecting an upstream server needs to match this signature
type SelectionAlgorithm func(*Replica) *url.URL

func (t *Replica) SelectUpstream(upstreamSelector SelectionAlgorithm) *url.URL {
	upstream := upstreamSelector(t)

	return upstream
}

func (t *Replica) RemoveUpstream(faultyUpstream *url.URL) {
	t.Lock.Lock()
	defer t.Lock.Unlock()
	for i, upstream := range t.Upstreams {
		if upstream.String() == faultyUpstream.String() {
			t.Upstreams = append(t.Upstreams[:i], t.Upstreams[i+1:]...)
			fmt.Printf("Removed faulty upstream: %s\n", faultyUpstream.String())

			// sending message to admin server when removing replica``
			// msg := messaging.Message{
			// 	Name: messaging.REMOVE_REPLICA,
			// 	Body: faultyUpstream.String(),
			// }

			// messaging.PublishMessage(messaging.PUBLISHING_QUEUE, &msg)
			return
		}
	}
}

func (t *Replica) AddUpstream(upstream string) {
	url, err := url.Parse(upstream)

	if err != nil {
		fmt.Printf("error during parsing url %v", err)
	}
	t.Upstreams = append(t.Upstreams, url)
}
