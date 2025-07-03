package prequal

import (
	"math/rand/v2"
	"net/url"
	"sort"
	"time"
)

type UpstreamDecisionQueueItem struct {
	destination *url.URL
	insertTime  time.Time
}

type UpstreamDecisionQueue struct {
	UpstreamDecisionQueue []UpstreamDecisionQueueItem
	queueSize             uint64
	hotColdQuantile       float64
}

func (u *UpstreamDecisionQueue) Enqueue(item *url.URL) {
	u.UpstreamDecisionQueue = append(u.UpstreamDecisionQueue, UpstreamDecisionQueueItem{
		destination: item,
		insertTime:  time.Now(),
	})
	u.queueSize++
}

func (u *UpstreamDecisionQueue) DeQueue() UpstreamDecisionQueueItem {
	item := u.UpstreamDecisionQueue[0]
	u.UpstreamDecisionQueue = u.UpstreamDecisionQueue[1:]
	u.queueSize--

	return item
}

func (u *UpstreamDecisionQueue) EmptyQueue() {
	u.UpstreamDecisionQueue = make([]UpstreamDecisionQueueItem, 0, 10)
}

func (u *UpstreamDecisionQueue) ProbeToReduceLatencyAndQueuingAlgorithm(r *Replica, probeQueue *ServerProbeQueue) {
	probes := probeQueue.Probes
	numberOfProbes := probeQueue.Size
	if len(r.Upstreams) == 1 || numberOfProbes <= 0 {
		u.Enqueue(r.Upstreams[0])
		return
	}

	if numberOfProbes == 2 {
		u.Enqueue(r.Upstreams[rand.IntN(2)])
		return
	}

	sort.Slice(probes, func(i, j int) bool {
		return probes[i].RequestsInFlight < probes[j].RequestsInFlight
	})

	partiton_index := int(u.hotColdQuantile * float64(numberOfProbes-1))

	if partiton_index == numberOfProbes-1 {
		lowestLatencyIndex := 0
		for i := range numberOfProbes {
			if probes[i].Latency < probes[lowestLatencyIndex].Latency {
				lowestLatencyIndex = i
			}
		}
		u.Enqueue(probes[lowestLatencyIndex].Upstream)
		return
	}

	minRIFIndex := 0
	for i := range partiton_index {
		if probes[i].RequestsInFlight < probes[minRIFIndex].RequestsInFlight {
			minRIFIndex = i
		}
	}
	u.Enqueue(probes[minRIFIndex].Upstream)
}

func (u *UpstreamDecisionQueue) ProbingToReduceLatencyAndQueuing(r *Replica, probeQueue *ServerProbeQueue) *url.URL {
	u.ProbeToReduceLatencyAndQueuingAlgorithm(r, probeQueue)

	return u.DeQueue().destination
}
