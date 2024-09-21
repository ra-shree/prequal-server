package common

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/rand/v2"
	"net/http"
	"net/url"
	"os"
	"slices"
	"sync"
	"time"
)

var maxLifeTime time.Duration = 5 * time.Second
var poolSize = 16
var probeFactor = 0.5
var probeRemoveFactor = 1
var totalReplica = 1
var mu = 1

var denom = (1-poolSize/totalReplica)*int(probeFactor) - probeRemoveFactor
var reuseRate = max(1, ((1 + mu) / denom))

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
		Capacity: poolSize,
		Probes:   make([]ServerProbeItem, 0, 16),
	}
}

func (q *ServerProbeQueue) ProbesInQueue() []ServerProbeItem {
	return q.Probes
}

func (q *ServerProbeQueue) Add(probe *ServerProbe) {
	q.mutex.Lock()
	defer q.mutex.Unlock()

	fmt.Printf("\n\nAdded a probe and current size %d and capacity %d\n\n", q.Size, q.Capacity)
	if q.Size == q.Capacity {
		q.Start = (q.Start + 1) % q.Capacity
	} else {
		q.Size++
	}

	q.Probes = slices.Insert(q.Probes, q.End, *NewServerProbeItem(probe))
	q.End = (q.End + 1) % q.Capacity
}

func (q *ServerProbeQueue) Remove(index int) bool {
	q.mutex.Lock()
	defer q.mutex.Unlock()

	if q.Size == 0 {
		return false
	}

	newProbeList := append(q.Probes[:index], q.Probes[index+1:]...)
	q.Probes = newProbeList
	q.Size--
	q.End--

	return true
}

func (q *ServerProbeQueue) RemoveOldest() bool {
	q.mutex.Lock()
	defer q.mutex.Unlock()

	if q.Size == 0 {
		return false
	}

	q.Probes = q.Probes[1:]

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

	newProbeList := make([]ServerProbeItem, 0, 16)

	for i := 0; i < q.Size; i++ {
		index := (q.Start + i) % q.Capacity
		probe := q.Probes[index]

		// Skip probes that are still valid (i.e., keep them in the queue)
		if time.Since(probe.ReceiptTime) < maxLifeTime && probe.used < reuseRate {
			newProbeList = append(newProbeList, probe)
			continue
		}

		// Remove the probe by adjusting Start index (effectively deleting)
		removed++
	}

	q.Probes = newProbeList
	// Adjust the queue size and the start pointer
	q.Size -= removed
	q.Start = (q.Start + removed) % q.Capacity

	return true
}

func getProbe(url *url.URL) (*ServerProbe, error) {
	res, err := http.Get(fmt.Sprintf("%s://%s/%s", url.Scheme, url.Host, "ping"))
	if err != nil {
		log.Printf("error making get request %v", err)
		// do something about the replica like removing it from the pool
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		// do somethign about replica here too
		log.Printf("received non-200 response: %d", res.StatusCode)
	}

	body, err := io.ReadAll(res.Body)

	if err != nil {
		log.Printf("error reading response body %v", err)
	}

	var probeRes ProbeResponse

	err = json.Unmarshal([]byte(body), &probeRes)
	if err != nil {
		log.Printf("error parsing JSON: %v", err)
	}

	err = os.WriteFile("probe.log", body, 0644)
	if err != nil {
		log.Printf("error writing JSON to file: %v", err)
	}

	fmt.Printf("\n\nResponse JSON:::::::: \n %v %v %v", probeRes.ServerName, probeRes.RequestsInFlight, probeRes.Latency)

	newProbe := NewServerProbe(&probeRes, url)
	return newProbe, nil
}

func ProbeService(w http.ResponseWriter, r *http.Request, replicas []*Replica) {
	probeRate := randomRound(probeFactor)

	numUpstreams := len(replicas[0].Upstreams)
	perm := rand.Perm(numUpstreams)

	for i := 0; i < probeRate; i++ {
		newProbe, err := getProbe(replicas[0].Upstreams[perm[i]])
		if err != nil {
			fmt.Printf("error when getting probe %v", err)
			continue
		}
		ProbeQueue.Add(newProbe)
	}
}

func PeriodicProbeService(t time.Time, replicas []*Replica) {
	fmt.Printf("\nPeriodic Probe Request: %v\n", t)
	probeRate := 3

	numUpstreams := len(replicas[0].Upstreams)
	perm := rand.Perm(numUpstreams)

	for i := 0; i < probeRate; i++ {
		newProbe, err := getProbe(replicas[0].Upstreams[perm[i]])
		if err != nil {
			fmt.Printf("error when getting probe %v", err)
			continue
		}
		ProbeQueue.Add(newProbe)
	}
}

func ProbeCleanService(t time.Time) {
	fmt.Printf("\nCleaning Probe Queue: %v\n", t)
	ProbeQueue.RemoveProbes()
}
