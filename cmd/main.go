package main

import (
	"log"
	"os"
	"os/signal"

	"github.com/gorilla/mux"
	"github.com/ra-shree/prequal-server/pkg/reverseproxy"
)

func main() {
	proxy := &reverseproxy.ReverseProxy{}

	r := mux.NewRouter()
	r.Host("localhost").PathPrefix("/api")

	proxy.AddReplica([]string{"http://localhost:9000"}, r)

	proxy.AddReplica([]string{
		"http://localhost:9001",
		"http://localhost:9002",
		"http://localhost:9003",
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
