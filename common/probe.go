package common

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/rand/v2"
	"net/http"
	"net/url"
	"sync"
	"time"
)

type PrequalParameters struct {
	MaxLifeTime       time.Duration
	PoolSize          int
	ProbeFactor       float64
	ProbeRemoveFactor int
	TotalReplica      int
	Mu                int
	Denominator       int
	ReuseRate         int
}

func NewPrequalParameters(maxLifeTime int, poolSize int, probeFactor float64, probeRemoveFactor int, totalReplica int, mu int) *PrequalParameters {
	return &PrequalParameters{
		MaxLifeTime:       time.Duration(maxLifeTime) * time.Second,
		PoolSize:          poolSize,
		ProbeFactor:       probeFactor,
		ProbeRemoveFactor: probeRemoveFactor,
		TotalReplica:      totalReplica,
		Mu:                mu,
		Denominator:       (1-poolSize/totalReplica)*int(probeFactor) - probeRemoveFactor,
		ReuseRate:         max(1, ((1+mu)/(1-poolSize/totalReplica)*int(probeFactor) - probeRemoveFactor)),
	}
}

var CurrentPrequalParameters PrequalParameters

// var maxLifeTime time.Duration = 5 * time.Second
// var poolSize = 16
// var probeFactor = 1.2
// var probeRemoveFactor = 1
// var totalReplica = 1
// var mu = 1

// var denom = (1-poolSize/totalReplica)*int(probeFactor) - probeRemoveFactor
// var reuseRate = max(1, ((1 + mu) / denom))

var ProbeQueue = NewServerProbeQueue()

type ServerProbe struct {
	Name             string
	RequestsInFlight int
	Latency          int
	Upstream         *url.URL
}

type ServerProbeItem struct {
	used        int
	ReceiptTime time.Time
	ServerProbe
}

type ServerProbeQueue struct {
	Probes   []ServerProbeItem
	Start    int
	End      int
	Size     int
	Capacity int
	mutex    sync.Mutex
}

type ProbeResponse struct {
	ServerName       string `json:"serverName"`
	RequestsInFlight uint64 `json:"requestInFlight"`
	Latency          uint64 `json:"latency"`
}

func NewServerProbe(s *ProbeResponse, u *url.URL) *ServerProbe {
	return &ServerProbe{
		Name:             s.ServerName,
		RequestsInFlight: int(s.RequestsInFlight),
		Latency:          int(s.Latency),
		Upstream:         u,
	}
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
		Start:    0,
		End:      0,
		Size:     0,
		Capacity: 16,
		Probes:   make([]ServerProbeItem, 0, 16),
	}
}

func (q *ServerProbeQueue) ProbesInQueue() []ServerProbeItem {
	return q.Probes
}

func (q *ServerProbeQueue) Add(probe *ServerProbe) {
	q.mutex.Lock()
	defer q.mutex.Unlock()

	fmt.Printf("\n\nAdded a probe, current size %d, capacity %d\n\n", q.Size, q.Capacity)
	if q.Size == q.Capacity {
		// Overwrite the oldest element
		q.Probes[q.Start] = *NewServerProbeItem(probe)
		q.Start = (q.Start + 1) % q.Capacity
	} else {
		q.Probes = append(q.Probes, *NewServerProbeItem(probe))
		q.Size++
	}

	// Always move End forward
	q.End = (q.End + 1) % q.Capacity
}

func (q *ServerProbeQueue) Remove(index int) bool {
	q.mutex.Lock()
	defer q.mutex.Unlock()

	if q.Size == 0 || index < 0 || index >= q.Size {
		return false
	}

	// Compute the actual index in the circular queue
	actualIndex := (q.Start + index) % q.Capacity

	newProbeList := append(q.Probes[:actualIndex], q.Probes[actualIndex+1:]...)
	q.Probes = newProbeList
	// Shift elements after the removed index
	// for i := actualIndex; i != q.End; i = (i + 1) % q.Capacity {
	// 	next := (i + 1) % q.Capacity
	// 	q.Probes[i] = q.Probes[next]
	// }

	// Adjust the end index and size
	q.End = (q.End - 1 + q.Capacity) % q.Capacity
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
	q.Start = (q.Start + 1) % q.Capacity
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
	newProbeList := make([]ServerProbeItem, 0, q.Capacity)

	for i := 0; i < q.Size; i++ {
		index := (q.Start + i) % q.Capacity
		probe := q.Probes[index]

		// Skip probes that are still valid (i.e., keep them in the queue)
		if time.Since(probe.ReceiptTime) < CurrentPrequalParameters.MaxLifeTime && probe.used < CurrentPrequalParameters.ReuseRate {
			newProbeList = append(newProbeList, probe)
			continue
		}
		removed++
	}

	// Adjust the queue based on the elements kept
	q.Probes = newProbeList
	q.Size -= removed
	q.Start = 0
	q.End = q.Size % q.Capacity

	return true
}

func getProbe(replica *Replica, idx int) (*ServerProbe, error) {
	replica.Lock.RLock()
	defer replica.Lock.RUnlock()
	url := replica.Upstreams[idx]

	res, err := http.Get(fmt.Sprintf("%s://%s/%s", url.Scheme, url.Host, "ping"))
	if err != nil {
		log.Printf("error making get request %v", err)
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		log.Printf("received non-200 response: %d", res.StatusCode)
		return nil, err
	}

	body, err := io.ReadAll(res.Body)

	if err != nil {
		log.Printf("error reading response body %v", err)
		return nil, err
	}

	var probeRes ProbeResponse

	err = json.Unmarshal([]byte(body), &probeRes)
	if err != nil {
		log.Printf("error parsing JSON: %v", err)
		return nil, err
	}

	// err = os.WriteFile("probe.log", body, 0644)
	// if err != nil {
	// 	log.Printf("error writing JSON to file: %v", err)
	// 	return nil, err
	// }

	// fmt.Printf("\n\nResponse JSON:::::::: \n %v %v %v", probeRes.ServerName, probeRes.RequestsInFlight, probeRes.Latency)

	newProbe := NewServerProbe(&probeRes, url)
	return newProbe, nil
}

func ProbeService(w http.ResponseWriter, r *http.Request, replicas []*Replica) {
	probeRate := randomRound(CurrentPrequalParameters.ProbeFactor)

	numUpstreams := len(replicas[0].Upstreams)
	perm := rand.Perm(numUpstreams)

	for i := 0; i < probeRate; i++ {
		newProbe, err := getProbe(replicas[0], perm[i])
		if err != nil {
			fmt.Printf("error when getting probe in probe service %v\n", err)
			continue
		}
		ProbeQueue.Add(newProbe)
	}
}

func PeriodicProbeService(replica *Replica) {
	// fmt.Printf("\nPeriodic Probe Request: %v\n", t)
	probeRate := randomRound(CurrentPrequalParameters.ProbeFactor)

	numUpstreams := len(replica.Upstreams)
	perm := rand.Perm(numUpstreams)

	for i := 0; i < probeRate; i++ {
		newProbe, err := getProbe(replica, perm[i])
		if err != nil {
			fmt.Printf("error when getting probe %v\n", err)
			//publish message
			// messaging.ReplicaFailed(replica.Upstreams[perm[i]].String())

			// remove replica
			replica.RemoveUpstream(replica.Upstreams[perm[i]])

			FailedReplica = replica.Upstreams[perm[i]].String()

			continue
		}
		ProbeQueue.Add(newProbe)
	}
}

func ProbeCleanService() {
	// fmt.Printf("\nCleaning Probe Queue: %v\n", t)
	ProbeQueue.RemoveProbes()
}
