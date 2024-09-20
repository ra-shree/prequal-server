package algorithm

import (
	"math/rand/v2"
	"net/url"
	"sort"
	"sync"

	"github.com/ra-shree/prequal-server/pkg/common"
)

// needed to partition into hot and cold probes
var hot_cold_quantile float64 = 0.6
var lock sync.RWMutex

func ProbeToReduceLatencyAndQueuing(r *common.Replica) *url.URL {
	lock.RLock()
	defer lock.RUnlock()
	probes := common.ProbeQueue.ProbesInQueue()
	numberOfProbes := len(probes)

	if numberOfProbes < 2 {
		return r.Upstreams[rand.IntN(16)]
	}

	sort.Slice(probes, func(i, j int) bool {
		return probes[i].RequestsInFlight < probes[j].RequestsInFlight
	})

	partiton_index := int(hot_cold_quantile * float64(16-1))

	if partiton_index == 16 {
		lowestLatencyIndex := 0
		for i := 0; i < 16; i++ {
			if probes[i].Latency < probes[lowestLatencyIndex].Latency {
				lowestLatencyIndex = i
			}
		}
		return probes[lowestLatencyIndex].Upstream
	}

	minRIFIndex := 0
	for i := 0; i < partiton_index; i++ {
		if probes[i].RequestsInFlight < probes[minRIFIndex].RequestsInFlight {
			minRIFIndex = i
		}
	}
	return probes[minRIFIndex].Upstream
}
