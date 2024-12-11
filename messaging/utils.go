package messaging

import (
	"encoding/json"
	"log"
	"net/url"

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
	err = json.Unmarshal(bodyBytes, &replica)

	switch {
	case msg.Name == ADD_REPLICA:
		// err := json.Unmarshal([]byte(msg.Body.(string)), &replica)
		if err != nil {
			ReplicaAddFailed(replica.Url)
			return
		}

		reverseproxy.Proxy.Replicas[0].AddUpstream(replica.Url)

		// send the message to the admin server
		ReplicaAdded(replica.Url)
	case msg.Name == REMOVE_REPLICA:
		url, err := url.Parse(replica.Url)

		if err != nil {
			log.Printf("error parsing url for removal in message consumer: %v", err)
		}

		reverseproxy.Proxy.Replicas[0].RemoveUpstream(url)
		ReplicaRemoved(replica.Url)
	}
	// log.Print("Received a message \n")
	// log.Printf("Name: %v", msg.Name)
	// log.Printf("Body: %v", msg.Body)
}
