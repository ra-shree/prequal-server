package messaging

import (
	"encoding/json"
	"log"

	"github.com/ra-shree/prequal-server/common"
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
	case msg.Name == "replica-added":
		message := Message{
			Name: "replica added successfully",
			Body: "",
		}

		PublishMessage(common.PUBLISHING_QUEUE, &message)

	}
	log.Print("Received a message \n")
	log.Printf("Name: %v", msg.Name)
	log.Printf("Body: %v", msg.Body)
}
