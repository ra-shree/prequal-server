package common

import (
	"fmt"
	"log"
	"sync"
)

type UpdateStatisticsArg struct {
	Url              string
	RequestsInFlight uint64
	Latency          uint64
	LastTenLatency   []uint64
}

func NewUpdateStatisticsArg(url string, requestsInFlight, latency uint64, lastTenLatency []uint64) *UpdateStatisticsArg {
	return &UpdateStatisticsArg{
		Url:              url,
		RequestsInFlight: requestsInFlight,
		Latency:          latency,
		LastTenLatency:   lastTenLatency,
	}
}

type ReplicaStatisticsParameters struct {
	Name               string   `json:"name"`
	SuccessfulRequests int      `json:"successful_requests"`
	FailedRequests     int      `json:"failed_requests"`
	RequestsInFlight   uint64   `json:"requests_in_flight"`
	Latency            uint64   `json:"latency"`
	LastTenLatency     []uint64 `json:"last_ten_latency"`
	Status             string   `json:"status"`
}

type ReplicaStatistics struct {
	replicaStatistics map[string]ReplicaStatisticsParameters
	lock              sync.RWMutex
}

func InitializeStatistics(replicas []ReplicaConfig) *ReplicaStatistics {
	replicaStatisticsMap := make(map[string]ReplicaStatisticsParameters)
	for _, replica := range replicas {
		AddKey(replicaStatisticsMap, replica)
	}

	return &ReplicaStatistics{
		replicaStatistics: replicaStatisticsMap,
	}
}

func AddKey(replicaStatisticsMap map[string]ReplicaStatisticsParameters, replica ReplicaConfig) {
	log.Print("Appending new replica to stat: ", replica)
	replicaStatisticsMap[replica.URL] = ReplicaStatisticsParameters{
		SuccessfulRequests: 0,
		FailedRequests:     0,
		Name:               replica.Name,
		Status:             "active",
	}
}

func (r *ReplicaStatistics) UpdateStatistics(probes []*UpdateStatisticsArg) {
	r.lock.Lock()
	defer r.lock.Unlock()

	for i := range probes {
		stat := r.replicaStatistics[probes[i].Url]
		stat.Latency = probes[i].Latency
		stat.LastTenLatency = probes[i].LastTenLatency
		stat.RequestsInFlight = probes[i].RequestsInFlight

		r.replicaStatistics[probes[i].Url] = stat
	}
}

func (r *ReplicaStatistics) UpdateStatus(replica, status string) {
	stats := r.replicaStatistics[replica]
	stats.Status = status
	r.replicaStatistics[replica] = stats
}

func (r *ReplicaStatistics) IncrementSuccessfulRequests(replica string) {
	r.lock.Lock()
	defer r.lock.Unlock()

	if stats, exists := r.replicaStatistics[replica]; exists {
		stats.SuccessfulRequests += 1
		r.replicaStatistics[replica] = stats
	}
}

func (r *ReplicaStatistics) IncrementFailedRequests(replica string) {
	r.lock.Lock()
	defer r.lock.Unlock()

	if stats, exists := r.replicaStatistics[replica]; exists {
		stats.FailedRequests += 1
		r.replicaStatistics[replica] = stats
	}
}

func (r *ReplicaStatistics) PrintStatistics() {
	r.lock.RLock()
	defer r.lock.RUnlock()

	if len(r.replicaStatistics) == 0 {
		log.Println("No statistics data available.")
		return
	}

	fmt.Println("Replica Statistics:")
	for url, stats := range r.replicaStatistics {
		fmt.Printf("URL: %s, STATS: %+v\n", url, stats)
	}
}

func (r *ReplicaStatistics) GetStatistics() []ReplicaStatisticsParameters {
	r.lock.RLock()
	defer r.lock.RUnlock()

	statistics := make([]ReplicaStatisticsParameters, 0, len(r.replicaStatistics))

	for _, value := range r.replicaStatistics {
		statistics = append(statistics, value)
	}

	return statistics
}
