package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"time"

	// "github.com/gorilla/mux"
	"github.com/gorilla/mux"
	"github.com/ra-shree/prequal-server/pkg/common"
	"github.com/ra-shree/prequal-server/pkg/reverseproxy"
)

func main() {
	proxy := &reverseproxy.ReverseProxy{}
	r := mux.NewRouter()
	r.Host("localhost").PathPrefix("/")

	// proxy.AddReplica([]string{"http://localhost:9000"}, r)

	proxy.AddReplica([]string{
		"http://localhost:9001",
		"http://localhost:9002",
		"http://localhost:9003",
		"http://localhost:9004",
	}, r)

	periodicProbetime := time.NewTicker(3 * time.Second)

	go func() {
		for i := range periodicProbetime.C {
			common.PeriodicProbeService(i, proxy.Replicas)
		}
	}()

	probeCleanTimer := time.NewTicker(5 * time.Second)
	go func() {
		for i := range probeCleanTimer.C {
			common.ProbeCleanService(i)
		}
	}()

	probeSliceChecker := time.NewTicker(1 * time.Second)
	go func() {
		for j := range probeSliceChecker.C {
			fmt.Print(j)
			fmt.Printf("\n\nProbe Number %v\t\t", len(common.ProbeQueue.Probes))
		}
	}()
	// proxy.AddReplica([]string{"http://localhost:8000"}, r)
	proxy.AddListener(":8080")
	if err := proxy.Start(common.ProbeService); err != nil {
		log.Fatal(err)
	}

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c
}
