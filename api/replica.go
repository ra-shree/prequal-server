package api

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
)

type GetReplicaResponse struct {
	Success bool
	Data    []ReplicaResponse
}

func GetReplicas() []string {
	res, err := http.Get(fmt.Sprintf("%s/%s", adminUrl, "get-replica"))

	if err != nil {
		panic(err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		log.Panicf("received non-200 response: %d", res.StatusCode)
	}

	body, err := io.ReadAll(res.Body)

	if err != nil {
		log.Panicf("error reading response body %v", err)
	}

	var replicaResponse GetReplicaResponse

	err = json.Unmarshal([]byte(body), &replicaResponse)
	if err != nil {
		log.Panicf("error parsing JSON: %v", err)
	}

	// get the url from the response body and put it in an array
	var replicaUrls []string
	for _, replica := range replicaResponse.Data {
		if replica.Status == "active" {
			replicaUrls = append(replicaUrls, replica.Url)
		}
	}
	return replicaUrls
}
