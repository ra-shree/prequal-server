package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"time"

	"github.com/gorilla/mux"
	"github.com/ra-shree/prequal-server/algorithm"
	"github.com/ra-shree/prequal-server/common"
	"github.com/ra-shree/prequal-server/messaging"
	"github.com/ra-shree/prequal-server/reverseproxy"
)

func main() {
	go func() {
		messaging.SetupPublisher()
	}()

	proxy := &reverseproxy.ReverseProxy{}
	r := mux.NewRouter()
	r.Host("localhost").PathPrefix("/")

	proxy.AddReplica([]string{
		"http://localhost:9001",
		"http://localhost:9002",
		"http://localhost:9003",
		"http://localhost:9004",
	}, r)

	periodicProbetime := time.NewTicker(500 * time.Millisecond)
	go func() {
		for i := range periodicProbetime.C {
			common.PeriodicProbeService(i, proxy.Replicas)
			algorithm.ProbeToReduceLatencyAndQueuingAlgorithm(proxy.Replicas[0])
		}
	}()

	probeCleanTimer := time.NewTicker(2000 * time.Millisecond)
	go func() {
		for i := range probeCleanTimer.C {
			common.ProbeCleanService(i)
			algorithm.EmptyQueue()
		}
	}()

	probeSliceChecker := time.NewTicker(500 * time.Millisecond)
	go func() {
		for j := range probeSliceChecker.C {
			fmt.Print(j)
			fmt.Printf("\n\nProbe Number %v\t\t", len(common.ProbeQueue.Probes))
		}
	}()

	proxy.AddListener(":8000")
	if err := proxy.Start(common.ProbeService); err != nil {
		log.Fatal(err)
	}

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c
}
