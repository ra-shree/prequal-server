package prequal

import (
	"fmt"
	"math/rand/v2"
	"net/http"
	"sync"
	"time"

	"github.com/ra-shree/prequal-server/common"
)

type ServerProbeQueue struct {
	Probes            []ServerProbeItem
	start             int
	end               int
	Size              int
	capacity          int
	mutex             sync.Mutex
	prequalParameters PrequalParameters
}

func NewServerProbeQueue(capacity int, prequalParameters PrequalParameters) *ServerProbeQueue {
	return &ServerProbeQueue{
		start:             0,
		end:               0,
		Size:              0,
		capacity:          capacity,
		Probes:            make([]ServerProbeItem, 0, capacity),
		prequalParameters: prequalParameters,
	}
}

func (q *ServerProbeQueue) ProbesInQueue() []ServerProbeItem {
	return q.Probes
}

func (q *ServerProbeQueue) Add(probe *ServerProbe) {
	q.mutex.Lock()
	defer q.mutex.Unlock()

	fmt.Printf("\nAdded a probe, current size %d, capacity %d", q.Size, q.capacity)
	if q.Size == q.capacity {
		// Overwrite the oldest element
		q.Probes[q.start] = *NewServerProbeItem(probe)
		q.start = (q.start + 1) % q.capacity
	} else {
		q.Probes = append(q.Probes, *NewServerProbeItem(probe))
		q.Size++
	}

	// Always move End forward
	q.end = (q.end + 1) % q.capacity
}

func (q *ServerProbeQueue) Remove(index int) bool {
	q.mutex.Lock()
	defer q.mutex.Unlock()

	if q.Size == 0 || index < 0 || index >= q.Size {
		return false
	}

	// Compute the actual index in the circular queue
	actualIndex := (q.start + index) % q.capacity

	newProbeList := append(q.Probes[:actualIndex], q.Probes[actualIndex+1:]...)
	q.Probes = newProbeList

	// Adjust the end index and size
	q.end = (q.end - 1 + q.capacity) % q.capacity
	q.Size--

	return true
}

func (q *ServerProbeQueue) RemoveOldest() bool {
	q.mutex.Lock()
	defer q.mutex.Unlock()

	if q.Size == 0 {
		return false
	}

	// Increment the Start pointer to remove the oldest item
	q.start = (q.start + 1) % q.capacity
	q.Size--

	return true
}

func (q *ServerProbeQueue) RemoveProbes() bool {
	q.mutex.Lock()
	defer q.mutex.Unlock()

	if q.Size == 0 {
		return false
	}

	removed := 0
	newProbeList := make([]ServerProbeItem, 0, q.capacity)

	for i := 0; i < q.Size; i++ {
		index := (q.start + i) % q.capacity
		probe := q.Probes[index]

		// Skip probes that are still valid (i.e., keep them in the queue)
		if time.Since(probe.ReceiptTime) < q.prequalParameters.MaxLifeTime && probe.used < q.prequalParameters.ReuseRate {
			newProbeList = append(newProbeList, probe)
			continue
		}
		removed++
	}

	// Adjust the queue based on the elements kept
	q.Probes = newProbeList
	q.Size -= removed
	q.start = 0
	q.end = q.Size % q.capacity

	return true
}

func (q *ServerProbeQueue) ProbeCleanService() {
	q.RemoveProbes()
}

func (q *ServerProbeQueue) ProbeService(w http.ResponseWriter, r *http.Request, replicas []*Replica) []*common.UpdateStatisticsArg {
	probeRate := common.RandomRound(q.prequalParameters.ProbeFactor)

	numUpstreams := len(replicas[0].Upstreams)
	perm := rand.Perm(numUpstreams)

	latestProbes := make([]*common.UpdateStatisticsArg, 0, len(replicas[0].Upstreams))

	for i := range probeRate {
		newProbe, err := getProbe(replicas[0], perm[i])
		if err != nil {
			fmt.Printf("error when getting probe in probe service %v\n", err)
			continue
		}

		updateStat := common.NewUpdateStatisticsArg(
			replicas[0].Upstreams[perm[i]].String(),
			newProbe.RequestsInFlight,
			newProbe.Latency,
			newProbe.LastTenLatency,
		)
		latestProbes = append(latestProbes, updateStat)
		q.Add(newProbe)
	}

	return latestProbes
}

func (q *ServerProbeQueue) PeriodicProbeService(replica *Replica) []*common.UpdateStatisticsArg {
	probeRate := common.RandomRound(q.prequalParameters.ProbeFactor)

	numUpstreams := len(replica.Upstreams)
	perm := rand.Perm(numUpstreams)

	latestProbes := make([]*common.UpdateStatisticsArg, 0, len(replica.Upstreams))

	for i := range probeRate {
		newProbe, err := getProbe(replica, perm[i])
		if err != nil {
			fmt.Printf("error when getting probe %v\n", err)
			replica.RemoveUpstream(replica.Upstreams[perm[i]])

			FailedReplica = replica.Upstreams[perm[i]].String()

			continue
		}

		updateStat := common.NewUpdateStatisticsArg(
			replica.Upstreams[perm[i]].String(),
			newProbe.RequestsInFlight,
			newProbe.Latency,
			newProbe.LastTenLatency,
		)
		latestProbes = append(latestProbes, updateStat)

		q.Add(newProbe)
	}

	return latestProbes
}
