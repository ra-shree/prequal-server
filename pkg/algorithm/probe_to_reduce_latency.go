package algorithm

import (
	"math/rand/v2"
	"net/url"
	"sort"

	"github.com/ra-shree/prequal-server/pkg/common"
)

// needed to partition into hot and cold probes
var hot_cold_quantile float64 = 0.6

func ProbeToReduceLatencyAndQueuing(r *common.Replica) *url.URL {
	probes := common.ProbeQueue.Probes
	numberOfProbes := len(probes)

	// for i := 0; i < len(r.Upstreams); i++ {
	// 	fmt.Printf("\n\nReplica:\t\t%v\n\n", r.Upstreams[i])
	// }
	if numberOfProbes < 2 {
		if rand.IntN(2) == 0 {
			return r.Upstreams[0]
		}
		return r.Upstreams[1]
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
