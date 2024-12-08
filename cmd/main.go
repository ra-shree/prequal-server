package main

import (
	"log"
	"os"
	"os/signal"
	"time"

	"github.com/gorilla/mux"
	"github.com/ra-shree/prequal-server/algorithm"
	"github.com/ra-shree/prequal-server/api"
	"github.com/ra-shree/prequal-server/common"
	"github.com/ra-shree/prequal-server/messaging"
	"github.com/ra-shree/prequal-server/reverseproxy"
)

func main() {
	log.Print("Establishing connection with publishing queue...")
	messaging.InitializePublisher()
	defer messaging.CleanupPublisher()

	log.Print("Starting consumer service...")
	go func() {
		messaging.SetupConsumer()
	}()

	// msg := messaging.Message{
	// 	Name: "example",
	// 	Body: []string{"item1", "item2", "item3"},
	// }
	// messaging.PublishMessage("reverseproxy-to-admin", &msg)

	log.Print("Initializing reverse proxy...")
	proxy := &reverseproxy.ReverseProxy{}
	r := mux.NewRouter()
	r.Host("localhost").PathPrefix("/")

	log.Print("Registering replica servers from database...")
	replicas := api.GetReplicas()

	if len(replicas) == 0 {
		log.Panicf("No replicas found in database. Exiting...")
	}

	for _, replica := range replicas {
		log.Printf("Adding replica with url: %v", replica)
	}

	proxy.AddReplica(replicas, r)

	// proxy.AddReplica([]string{
	// 	"http://localhost:9001",
	// 	"http://localhost:9002",
	// 	"http://localhost:9003",
	// 	"http://localhost:9004",
	// }, r)

	log.Print("Replicas added successfully")

	log.Print("Starting probe service...")
	periodicProbetime := time.NewTicker(500 * time.Millisecond)
	go func() {
		for i := range periodicProbetime.C {
			common.PeriodicProbeService(i, proxy.Replicas)
			algorithm.ProbeToReduceLatencyAndQueuingAlgorithm(proxy.Replicas[0])
		}
	}()

	log.Print("Starting probe cleaner service...")
	probeCleanTimer := time.NewTicker(2000 * time.Millisecond)
	go func() {
		for i := range probeCleanTimer.C {
			common.ProbeCleanService(i)
			algorithm.EmptyQueue()
		}
	}()

	// Checks the number of probes currently in queue
	// probeSliceChecker := time.NewTicker(500 * time.Millisecond)
	// go func() {
	// 	for j := range probeSliceChecker.C {
	// 		fmt.Print(j)
	// 		fmt.Printf("\n\nProbe Number %v\t\t", len(common.ProbeQueue.Probes))
	// 	}
	// }()

	log.Print("Starting reverse proxy...")
	proxy.AddListener(":8000")
	if err := proxy.Start(common.ProbeService); err != nil {
		log.Fatal(err)
	}

	log.Print("Reverse proxy server is ready to serve requests. Press Ctrl+C to stop the server.")
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c
}
