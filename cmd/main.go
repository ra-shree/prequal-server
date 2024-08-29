package main

import (
	"log"
	"net/http"

	"github.com/ra-shree/prequal-server/utils"
)

var probeResponse = make(chan utils.ServerProbe)

// check if there is a value in the channel
// if yes, then add it to my server probe queue
// otherwise, keep going

var serverMap = map[string]string{
	"server1": "http://localhost:9000",
	"server2": "http://localhost:9001",
	"server3": "http://localhost:9002",
}

func reverseProxyHandler() {

}

func main() {
	mux := http.NewServeMux()
	// mux.HandleFunc("/", reverseProxyHandler)

	log.Println("Starting reverse proxy")

	if err := http.ListenAndServe(":8000", mux); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
