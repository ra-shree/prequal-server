package common

import (
	"net/url"
	"time"
)

var maxLifeTime time.Duration

type ServerProbe struct {
	Name             string
	RequestsInFlight int
	Latency          int
	upstream         *url.URL
}

type ServerProbeItem struct {
	used        int
	ReceiptTime time.Time
	ServerProbe
}

type ServerProbeQueue struct {
	probes   []ServerProbeItem
	start    int
	end      int
	size     int
	capacity int
}

func NewServerProbeItem(s *ServerProbe) *ServerProbeItem {
	return &ServerProbeItem{
		ServerProbe: *s,
		used:        0,
		ReceiptTime: time.Now(),
	}
}

func NewServerProbeQueue() *ServerProbeQueue {
	return &ServerProbeQueue{
		start:    0,
		end:      0,
		size:     0,
		capacity: 16,
	}
}

func (q *ServerProbeQueue) Add(probe *ServerProbe) {
	if q.size == q.capacity {
		q.start = (q.start + 1) % q.capacity
	} else {
		q.size++
	}

	q.probes[q.end] = *NewServerProbeItem(probe)
	q.end = (q.end + 1) % q.capacity
}

func (q *ServerProbeQueue) Remove(index int) bool {
	if q.size == 0 {
		return false
	}

	newProbeList := append(q.probes[:index], q.probes[index+1:]...)
	q.probes = newProbeList
	q.size--
	q.end--

	return true
}

func (q *ServerProbeQueue) RemoveOldest() bool {
	if q.size == 0 {
		return false
	}

	q.start = (q.start + 1) % q.capacity
	q.size--

	return true
}

func (q *ServerProbeQueue) RemoveOldProbes() bool {
	if q.size == 0 {
		return false
	}

	var newProbeList []ServerProbeItem

	count := 0
	for _, probe := range q.probes {
		if time.Since(probe.ReceiptTime) < maxLifeTime {
			continue
		}
		count++
		newProbeList = append(newProbeList, probe)
	}

	q.probes = newProbeList
	q.size -= count
	q.end -= count
	return true
}
