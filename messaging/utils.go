package messaging

import (
	"encoding/json"
	"log"

	"github.com/gorilla/mux"
	"github.com/ra-shree/prequal-server/common"
	"github.com/ra-shree/prequal-server/reverseproxy"
)

type Message struct {
	Name string      `json:"name"`
	Body interface{} `json:"body"`
}

func failOnError(err error, msg string) {
	if err != nil {
		log.Panicf("%s: %s", msg, err)
	}
}

func processMessage(body []byte) {
	var msg Message
	err := json.Unmarshal(body, &msg)
	if err != nil {
		log.Printf("Failed to decode message: %v", err)
		return
	}

	switch {
	case msg.Name == ADD_REPLICA:
		type ReplicaMessage struct {
			Name                string `json:"name"`
			Url                 string `json:"url"`
			Status              string `json:"status"`
			HealthCheckEndpoint string `json:"healthcheckendpoint"`
		}
		bodyBytes, err := json.Marshal(msg.Body)
		if err != nil {
			log.Printf("Failed to marshal message body: %v", err)
			return
		}

		var replica ReplicaMessage
		// err := json.Unmarshal([]byte(msg.Body.(string)), &replica)
		err = json.Unmarshal(bodyBytes, &replica)
		if err != nil {
			message := Message{
				Name: REPLICA_ADD_FAILED,
				Body: msg.Body,
			}

			PublishMessage(PUBLISHING_QUEUE, &message)
			return
		}

		reverseproxy.Proxy.Replicas[0].AddUpstream(replica.Url)
		// reverseproxy.Proxy.
		message := Message{
			Name: ADDED_REPLICA,
			Body: msg.Body,
		}
		common.AppendNewReplica(replica.Url)
		// AddUpstream(replica.Url, common.MuxRouter)
		PublishMessage(PUBLISHING_QUEUE, &message)
		return
		// addReplica(&replica)

	}
	log.Print("Received a message \n")
	log.Printf("Name: %v", msg.Name)
	log.Printf("Body: %v", msg.Body)
}

func AddUpstream(upstream string, router *mux.Router) {
	// url, err := url.Parse(upstream)

	// if err != nil {
	// 	return nil, err
	// }
	// router := mux.NewRouter()

	router.Host("localhost").PathPrefix("/")

	reverseproxy.Proxy.Replicas[0].AddUpstream(upstream)
	log.Printf("Added upstream: %s\n", upstream)
	// t.Upstreams = append(t.Upstreams, url)

	// return t.Upstreams, nil
}
