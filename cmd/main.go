package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"github.com/ra-shree/prequal-server/common"
	"github.com/ra-shree/prequal-server/prequal"
	"github.com/ra-shree/prequal-server/reverseproxy"
)

func main() {
	config := common.LoadConfig("config.yaml")
	fmt.Printf("Parsed Config: %+v\n", config)

	log.Print("Initializing routes...")
	common.MuxRouter = mux.NewRouter()

	log.Printf("Register statistics route at /%v ...", config.Server.StatRoute)
	common.MuxRouter.HandleFunc(fmt.Sprintf("/%v", config.Server.StatRoute), func(w http.ResponseWriter, r *http.Request) {
		stats := reverseproxy.Proxy.ReplicaStatitics.GetStatistics()

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(stats); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})

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

	log.Print("Configuring algorithm parameters from database...")
	prequalParameters := *prequal.NewPrequalParameters(
		config.Algorithm.MaxLifeTime,
		config.Algorithm.PoolSize,
		config.Algorithm.ProbeFactor,
		int(config.Algorithm.ProbeRemoveFactor),
		len(replicaList),
		int(config.Algorithm.Mu))

	log.Print("Initializing the reverse proxy...")
	reverseproxy.Proxy = &reverseproxy.ReverseProxy{
		ProbeQueue:            prequal.NewServerProbeQueue(config.Algorithm.PoolSize, prequalParameters),
		ReplicaStatitics:      common.InitializeStatistics(config.Replicas),
		UpstreamDecisionQueue: &prequal.UpstreamDecisionQueue{},
	}

	common.MuxRouter.PathPrefix("/")
	reverseproxy.Proxy.AddReplica(replicaList, common.MuxRouter)

	log.Print("All instances added successfully.")

	log.Print("Starting probe service...")
	periodicProbetime := time.NewTicker(500 * time.Millisecond)
	go func() {
		for range periodicProbetime.C {
			newProbes := reverseproxy.Proxy.ProbeQueue.PeriodicProbeService(reverseproxy.Proxy.Replicas[0])

			if newProbes != nil {
				reverseproxy.Proxy.ReplicaStatitics.UpdateStatistics(newProbes)
			}

			reverseproxy.Proxy.UpstreamDecisionQueue.ProbeToReduceLatencyAndQueuingAlgorithm(reverseproxy.Proxy.Replicas[0], reverseproxy.Proxy.ProbeQueue)
		}
	}()

	log.Print("Starting probe cleaner service...")
	probeCleanTimer := time.NewTicker(2 * time.Second)
	go func() {
		for range probeCleanTimer.C {
			reverseproxy.Proxy.ProbeQueue.ProbeCleanService()
			reverseproxy.Proxy.UpstreamDecisionQueue.EmptyQueue()
		}
	}()
	log.Print("Starting reverse proxy...")

	port := fmt.Sprintf(":%s", strconv.Itoa(config.Server.Port))
	reverseproxy.Proxy.AddListener(port)
	if err := reverseproxy.Proxy.Start(reverseproxy.Proxy.ProbeQueue.ProbeService, config); err != nil {
		log.Fatal(err)
	}

	log.Print("Reverse proxy server is ready to serve requests. Press Ctrl+C to stop the server.")
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c
}
