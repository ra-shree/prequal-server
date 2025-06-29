package common

import (
	"fmt"
	"log"
	"sync"
)

type ReplicaStatisticsParameters struct {
	SuccessfulRequests int
	FailedRequests     int
}

type ReplicaStatistics struct {
	replicaStatistics map[string]ReplicaStatisticsParameters
	lock              sync.RWMutex
}

func InitializeStatistics(replicas []string) *ReplicaStatistics {
	replicaStatisticsMap := make(map[string]ReplicaStatisticsParameters)
	for _, replica := range replicas {
		AppendReplicaKey(replicaStatisticsMap, replica)
	}

	return &ReplicaStatistics{
		replicaStatistics: replicaStatisticsMap,
	}
}

func AppendReplicaKey(replicaStatisticsMap map[string]ReplicaStatisticsParameters, replica string) {
	log.Print("Appending new replica to stat: ", replica)
	replicaStatisticsMap[replica] = ReplicaStatisticsParameters{
		SuccessfulRequests: 0,
		FailedRequests:     0,
	}
}

func (r *ReplicaStatistics) IncrementSuccessfulRequests(replica string) {
	if stats, exists := r.replicaStatistics[replica]; exists {
		r.replicaStatistics[replica] = ReplicaStatisticsParameters{
			SuccessfulRequests: stats.SuccessfulRequests + 1,
			FailedRequests:     stats.FailedRequests,
		}
	} else {
		r.replicaStatistics[replica] = ReplicaStatisticsParameters{
			SuccessfulRequests: 0,
			FailedRequests:     0,
		}
	}
}

func (r *ReplicaStatistics) IncrementFailedRequests(replica string) {
	if stats, exists := r.replicaStatistics[replica]; exists {
		r.replicaStatistics[replica] = ReplicaStatisticsParameters{
			SuccessfulRequests: stats.SuccessfulRequests,
			FailedRequests:     stats.FailedRequests + 1,
		}
	} else {
		r.replicaStatistics[replica] = ReplicaStatisticsParameters{
			SuccessfulRequests: 0,
			FailedRequests:     0,
		}
	}

}

type statDataArr struct {
	ReplicaName string                      `json:"replica_name"`
	Statistics  ReplicaStatisticsParameters `json:"statistics"`
}

func (r *ReplicaStatistics) TransformMapToJson() []statDataArr {
	var jsonArray []statDataArr

	r.lock.RLock()
	defer r.lock.RUnlock()
	for name, stats := range r.replicaStatistics {
		// fmt.Printf("Converting to JSON - Name: %s, Stats: %+v\n", name, stats)
		jsonArray = append(jsonArray, statDataArr{
			ReplicaName: name,
			Statistics:  stats,
		})
	}
	// fmt.Printf("Final JSON Array: %+v\n", jsonArray)

	return jsonArray
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
		fmt.Printf("URL: %s\n", url)
		fmt.Printf("  Successful Requests: %d\n", stats.SuccessfulRequests)
		fmt.Printf("  Failed Requests: %d\n", stats.FailedRequests)
		fmt.Println("---")
	}
}
