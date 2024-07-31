package server

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/ra-shree/dynamic-load-balancer/internal/configs"
)

func Run() error {
	config, err := configs.NewConfiguration()
	if err != nil {
		return fmt.Errorf("could not load configuration: %v", err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/ping", ping)

	for _, resource := range config.Resources {
		destinationUrl, _ := url.Parse(resource.Destination_Url)
		proxy := NewProxy(destinationUrl)
		mux.HandleFunc(resource.Endpoint, ProxyRequestHandler(proxy, destinationUrl, resource.Endpoint))
	}

	if err := http.ListenAndServe(config.Server.Host+":"+config.Server.Listen_Port, mux); err != nil {
		return fmt.Errorf("could not start the server: %v", err)
	}

	return nil
}
