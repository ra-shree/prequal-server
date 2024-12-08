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
	file, err := os.OpenFile("app.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatal(err)
	}
	log.SetOutput(file)

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
	reverseproxy.Proxy = &reverseproxy.ReverseProxy{}

	common.MuxRouter = mux.NewRouter()
	common.MuxRouter.Host("localhost").PathPrefix("/")

	log.Print("Registering replica servers from database...")
	replicas := api.GetReplicas()

	if len(replicas) == 0 {
		log.Panicf("No replicas found in database. Exiting...")
	}

	// for _, replica := range replicas {
	// 	log.Printf("Adding replica with url: %v", replica)
	// }

	reverseproxy.Proxy.AddReplica(replicas, common.MuxRouter)
	log.Print("Replicas added successfully")

	log.Print("Configuring algorithm parameters from database...")
	prequalParameters := api.GetPrequalParameters()
	common.CurrentPrequalParameters = *common.NewPrequalParameters(
		prequalParameters["max_life_time"].(int),
		prequalParameters["pool_size"].(int),
		prequalParameters["probe_factor"].(float64),
		prequalParameters["probe_remove_factor"].(int),
		len(replicas),
		prequalParameters["mu"].(int))

	common.InitializeStatistics(replicas)

	// proxy.AddReplica([]string{
	// 	"http://localhost:9001",
	// 	"http://localhost:9002",
	// 	"http://localhost:9003",
	// 	"http://localhost:9004",
	// }, r)

	log.Print("Starting probe service...")
	periodicProbetime := time.NewTicker(500 * time.Millisecond)
	go func() {
		for range periodicProbetime.C {
			common.PeriodicProbeService(reverseproxy.Proxy.Replicas[0])
			algorithm.ProbeToReduceLatencyAndQueuingAlgorithm(reverseproxy.Proxy.Replicas[0])
		}
	}()

	log.Print("Starting probe cleaner service...")
	probeCleanTimer := time.NewTicker(2 * time.Second)
	go func() {
		for range probeCleanTimer.C {
			common.ProbeCleanService()
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

	log.Print("Start statistics service...")
	statisticsTimer := time.NewTicker(10 * time.Second)
	go func() {
		for range statisticsTimer.C {
			messaging.StatisticsUpdated()
		}
	}()

	log.Print("Starting reverse proxy...")
	reverseproxy.Proxy.AddListener(":8000")
	if err := reverseproxy.Proxy.Start(common.ProbeService); err != nil {
		log.Fatal(err)
	}

	log.Print("Reverse proxy server is ready to serve requests. Press Ctrl+C to stop the server.")
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c
}
