package main

import (
	"log"
	"os"
	"os/signal"
	"time"

	// "github.com/gorilla/mux"
	"github.com/ra-shree/prequal-server/pkg/common"
	"github.com/ra-shree/prequal-server/pkg/reverseproxy"
)

func main() {
	proxy := &reverseproxy.ReverseProxy{}

	probeCleanTimer := time.NewTicker(5 * time.Second)
	go func() {
		for i := range probeCleanTimer.C {
			common.ProbeCleanService(i)
		}
	}()

	// r := mux.NewRouter()
	// r.Host("localhost").PathPrefix("/api")

	// proxy.AddReplica([]string{"http://localhost:9000"}, r)

	proxy.AddReplica([]string{
		"http://localhost:1233",
	}, nil)

	// proxy.AddReplica([]string{"http://localhost:8000"}, r)
	proxy.AddListener(":8080")
	if err := proxy.Start(common.ProbeService); err != nil {
		log.Fatal(err)
	}

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c
}
