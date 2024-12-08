package common

import (
	"encoding/json"
	"log"
	"sync"
)

type ReplicaStatisticsParameters struct {
	SuccessfulRequests int
	FailedRequests     int
}

var ReplicaStatistics map[string]ReplicaStatisticsParameters
var lock = sync.RWMutex{}

func InitializeStatistics(replicas []string) {
	lock.Lock()
	defer lock.Unlock()
	ReplicaStatistics = make(map[string]ReplicaStatisticsParameters)
	for _, replica := range replicas {
		ReplicaStatistics[replica] = ReplicaStatisticsParameters{
			SuccessfulRequests: 0,
			FailedRequests:     0,
		}
	}
}

func AppendNewReplica(replica string) {
	lock.Lock()
	defer lock.Unlock()
	ReplicaStatistics[replica] = ReplicaStatisticsParameters{
		SuccessfulRequests: 0,
		FailedRequests:     0,
	}
}

func RemoveReplicaKey(replica string) {
	lock.Lock()
	defer lock.Unlock()
	delete(ReplicaStatistics, replica)
}

func IncrementSuccessfulRequests(replica string) {
	lock.Lock()
	defer lock.Unlock()
	ReplicaStatistics[replica] = ReplicaStatisticsParameters{
		SuccessfulRequests: ReplicaStatistics[replica].SuccessfulRequests + 1,
		FailedRequests:     ReplicaStatistics[replica].FailedRequests,
	}
}

func IncrementFailedRequests(replica string) {
	lock.Lock()
	defer lock.Unlock()
	ReplicaStatistics[replica] = ReplicaStatisticsParameters{
		SuccessfulRequests: ReplicaStatistics[replica].SuccessfulRequests,
		FailedRequests:     ReplicaStatistics[replica].FailedRequests + 1,
	}
}

func TransformMapToJson() []byte {
	var jsonArray []struct {
		ReplicaName string                      `json:"replica_name"`
		Statistics  ReplicaStatisticsParameters `json:"statistics"`
	}
	lock.RLock()
	defer lock.RUnlock()
	for name, stats := range ReplicaStatistics {
		jsonArray = append(jsonArray, struct {
			ReplicaName string                      `json:"replica_name"`
			Statistics  ReplicaStatisticsParameters `json:"statistics"`
		}{
			ReplicaName: name,
			Statistics:  stats,
		})
	}

	jsonData, err := json.Marshal(jsonArray)
	if err != nil {
		log.Fatalf("Failed to marshal JSON: %v", err)
	}

	return jsonData
}
