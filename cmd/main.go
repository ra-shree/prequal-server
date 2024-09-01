package main

import (
	"log"
	"os"
	"os/signal"

	"github.com/gorilla/mux"
	"github.com/ra-shree/prequal-server/reverseproxy"
)

func main() {
	proxy := &reverseproxy.ReverseProxy{}

	r := mux.NewRouter()
	r.Host("localhost").PathPrefix("/api")

	proxy.AddReplica([]string{"http://localhost:8000"}, r)

	proxy.AddReplica([]string{
		"http://localhost:8001",
		"http://localhost:8002",
		"http://localhost:8003",
	}, nil)

	proxy.AddReplica([]string{"http://localhost:8000"}, r)
	proxy.AddListener(":8080")
	if err := proxy.Start(); err != nil {
		log.Fatal(err)
	}

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c
}
