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

var maxLifeTime time.Duration = 5 * time.Second
var poolSize = 16
var probeFactor = 0.5
var probeRemoveFactor = 1
var totalReplica = 1
var mu = 1

var denom = (1-poolSize/totalReplica)*int(probeFactor) - probeRemoveFactor
var reuseRate = max(1, ((1 + mu) / denom))

var probeQueue = NewServerProbeQueue()

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
		upstream:         u,
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
		Probes:   make([]ServerProbeItem, poolSize),
	}
}

func (q *ServerProbeQueue) Add(probe *ServerProbe) {
	if q.Size == q.Capacity {
		q.Start = (q.Start + 1) % q.Capacity
	} else {
		q.Size++
	}

	q.Probes[q.End] = *NewServerProbeItem(probe)
	q.End = (q.End + 1) % q.Capacity
}

func (q *ServerProbeQueue) Remove(index int) bool {
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
	if q.Size == 0 {
		return false
	}

	q.Start = (q.Start + 1) % q.Capacity
	q.Size--

	return true
}

func (q *ServerProbeQueue) RemoveProbes() bool {
	if q.Size == 0 {
		return false
	}

	var newProbeList []ServerProbeItem

	count := 0
	for _, probe := range q.Probes {
		if time.Since(probe.ReceiptTime) < maxLifeTime || probe.used > reuseRate {
			continue
		}
		count++
		newProbeList = append(newProbeList, probe)
	}

	q.Probes = newProbeList
	q.Size -= count
	q.End -= count
	return true
}

func getProbe(url *url.URL) (*ServerProbe, error) {
	url.Path = url.Path + "/ping"

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

	// fmt.Printf("\n\nResponse JSON:::::::: \n %v", probeRes.ServerName)

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
		probeQueue.mutex.Lock()
		probeQueue.Add(newProbe)
		probeQueue.mutex.Unlock()
	}
}

func ProbeCleanService(t time.Time) {
	fmt.Printf("Cleaning Probe Queue: %v\n", t)
	probeQueue.mutex.Lock()
	defer probeQueue.mutex.Unlock()
	probeQueue.RemoveProbes()
}
