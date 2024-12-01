package algorithm

import (
	"math/rand/v2"
	"net/url"
	"sort"
	"time"

	"github.com/ra-shree/prequal-server/common"
)

type UpstreamDecisionQueueItem struct {
	destination *url.URL
	insertTime  time.Time
}

var UpstreamDecisionQueue []UpstreamDecisionQueueItem
var queueSize = 0

// needed to partition into hot and cold probes
var hot_cold_quantile float64 = 0.6

func Enqueue(item *url.URL) {
	UpstreamDecisionQueue = append(UpstreamDecisionQueue, UpstreamDecisionQueueItem{
		destination: item,
		insertTime:  time.Now(),
	})
	queueSize++
}

func DeQueue() UpstreamDecisionQueueItem {
	item := UpstreamDecisionQueue[0]
	UpstreamDecisionQueue = UpstreamDecisionQueue[1:]
	queueSize--

	return item
}

func EmptyQueue() {
	UpstreamDecisionQueue = make([]UpstreamDecisionQueueItem, 0, 10)
}

func ProbeToReduceLatencyAndQueuingAlgorithm(r *common.Replica) {
	probes := common.ProbeQueue.Probes
	numberOfProbes := common.ProbeQueue.Size

	// for i := 0; i < len(r.Upstreams); i++ {
	// 	fmt.Printf("\n\nReplica:\t\t%v\n\n", r.Upstreams[i])
	// }

	if len(r.Upstreams) == 1 || numberOfProbes <= 0 {
		Enqueue(r.Upstreams[0])
		return
	}

	if numberOfProbes == 2 {
		Enqueue(r.Upstreams[rand.IntN(2)])
		return
		// return r.Upstreams[rand.IntN(2)]
		// if rand.IntN(2) == 0 {
		// 	return r.Upstreams[0]
		// }
		// return r.Upstreams[1]
	}

	sort.Slice(probes, func(i, j int) bool {
		return probes[i].RequestsInFlight < probes[j].RequestsInFlight
	})

	partiton_index := int(hot_cold_quantile * float64(numberOfProbes-1))

	if partiton_index == numberOfProbes-1 {
		lowestLatencyIndex := 0
		for i := 0; i < numberOfProbes; i++ {
			if probes[i].Latency < probes[lowestLatencyIndex].Latency {
				lowestLatencyIndex = i
			}
		}
		Enqueue(probes[lowestLatencyIndex].Upstream)
		return
	}

	minRIFIndex := 0
	for i := 0; i < partiton_index; i++ {
		if probes[i].RequestsInFlight < probes[minRIFIndex].RequestsInFlight {
			minRIFIndex = i
		}
	}
	Enqueue(probes[minRIFIndex].Upstream)
}

func ProbingToReduceLatencyAndQueuing(r *common.Replica) *url.URL {
	ProbeToReduceLatencyAndQueuingAlgorithm(r)

	return DeQueue().destination
}
