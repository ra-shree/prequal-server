package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"github.com/ra-shree/prequal-server/algorithm"
	"github.com/ra-shree/prequal-server/common"
	"github.com/ra-shree/prequal-server/reverseproxy"
)

func main() {
	config := common.LoadConfig("config.yaml")

	fmt.Printf("Parsed Config: %+v\n", config)

	log.Print("Initializing reverse proxy...")
	probeQueue := common.NewServerProbeQueue(config.Algorithm.PoolSize)

	reverseproxy.Proxy = &reverseproxy.ReverseProxy{
		ProbeQueue: probeQueue,
	}

	common.MuxRouter = mux.NewRouter()
	common.MuxRouter.PathPrefix("/")

	log.Print("Registering server instances...")

	if len(config.Replicas) == 0 {
		log.Panic("no instances specified in config file.")
	}

	replicaList := []string{}
	for _, replica := range config.Replicas {
		log.Printf("Adding replica with url: %v", replica)
		_, err := http.Get(fmt.Sprintf("%s/%s", replica.URL, replica.Healthcheck))

		if err != nil {
			log.Printf("The server %v failed healthcheck", replica)
			continue
		}

		replicaList = append(replicaList, replica.URL)
	}

	if len(replicaList) == 0 {
		log.Fatal("none of the instances passed healthcheck. Exiting...")
	}

	reverseproxy.Proxy.AddReplica(replicaList, common.MuxRouter)
	log.Print("Instances added successfully.")

	log.Print("Configuring algorithm parameters from database...")
	common.CurrentPrequalParameters = *common.NewPrequalParameters(
		config.Algorithm.MaxLifeTime,
		config.Algorithm.PoolSize,
		config.Algorithm.ProbeFactor,
		int(config.Algorithm.ProbeRemoveFactor),
		len(replicaList),
		int(config.Algorithm.Mu))

	common.InitializeStatistics(replicaList)

	log.Print("Starting probe service...")
	periodicProbetime := time.NewTicker(500 * time.Millisecond)
	go func() {
		for range periodicProbetime.C {
			reverseproxy.Proxy.ProbeQueue.PeriodicProbeService(reverseproxy.Proxy.Replicas[0])
			algorithm.ProbeToReduceLatencyAndQueuingAlgorithm(reverseproxy.Proxy.Replicas[0], reverseproxy.Proxy.ProbeQueue)
		}
	}()

	log.Print("Starting probe cleaner service...")
	probeCleanTimer := time.NewTicker(2 * time.Second)
	go func() {
		for range probeCleanTimer.C {
			reverseproxy.Proxy.ProbeQueue.ProbeCleanService()
			algorithm.EmptyQueue()
		}
	}()

	// Checks the number of probes currently in queue
	probeSliceChecker := time.NewTicker(500 * time.Millisecond)
	go func() {
		for j := range probeSliceChecker.C {
			fmt.Print(j)
			fmt.Printf("\n\nProbe Number %v\t\t", len(reverseproxy.Proxy.ProbeQueue.Probes))
		}
	}()

	// log.Print("Start statistics service...")
	// statisticsTimer := time.NewTicker(10 * time.Second)
	// go func() {
	// 	for range statisticsTimer.C {
	// 		messaging.StatisticsUpdated()
	// 	}
	// }()

	log.Print("Starting reverse proxy...")

	port := fmt.Sprintf(":%s", strconv.Itoa(config.Server.Port))
	reverseproxy.Proxy.AddListener(port)
	if err := reverseproxy.Proxy.Start(reverseproxy.Proxy.ProbeQueue.ProbeService); err != nil {
		log.Fatal(err)
	}

	log.Print("Reverse proxy server is ready to serve requests. Press Ctrl+C to stop the server.")
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c
}
