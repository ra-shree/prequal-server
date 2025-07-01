package reverseproxy

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"sync"

	"github.com/gorilla/mux"
	"github.com/ra-shree/prequal-server/common"
	"github.com/ra-shree/prequal-server/prequal"
)

var Proxy *ReverseProxy

type ReverseProxy struct {
	listeners             []Listener
	Proxy                 *httputil.ReverseProxy
	servers               []*http.Server
	Replicas              []*prequal.Replica
	mutex                 sync.RWMutex
	ProbeQueue            *prequal.ServerProbeQueue
	ReplicaStatitics      *common.ReplicaStatistics
	UpstreamDecisionQueue *prequal.UpstreamDecisionQueue
}

func (r *ReverseProxy) AddListener(address string) {
	l := Listener{
		Addr: address,
	}

	r.listeners = append(r.listeners, l)
}

func (r *ReverseProxy) AddListenerTLS(address, tlsCert, tlsKey string) {
	l := Listener{
		Addr:    address,
		TLSCert: tlsCert,
		TLSKey:  tlsKey,
	}

	r.listeners = append(r.listeners, l)
}

type service func(http.ResponseWriter, *http.Request, []*prequal.Replica) []*common.UpdateStatisticsArg

func (r *ReverseProxy) AddReplica(upstreams []string, router *mux.Router) error {
	var urls = []*url.URL{}

	for _, upstream := range upstreams {
		url, err := url.Parse(upstream)

		if err != nil {
			return err
		}

		urls = append(urls, url)
	}

	r.Replicas = append(r.Replicas, &prequal.Replica{
		Router:    router,
		Upstreams: urls,
	})

	return nil
}

func (r *ReverseProxy) RemoveReplica(upstream string) error {
	for i, replica := range r.Replicas {
		for _, url := range replica.Upstreams {
			if url.String() == upstream {
				r.Replicas = append(r.Replicas[:i], r.Replicas[i+1:]...)
				return nil
			}
		}
	}
	return fmt.Errorf("replica with the specified upstream URL not found")
}

func (r *ReverseProxy) Start(probeService service, config common.Config) error {
	r.Proxy = &httputil.ReverseProxy{
		Director:       r.Director(),
		ErrorHandler:   r.ErrorHandler(),
		ModifyResponse: r.ModifyResponse(),
	}

	handler := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if req.URL.Path == fmt.Sprintf("/%v", config.Server.StatRoute) {
			fmt.Println("\nMatched route for path:", req.URL.Path)
			common.MuxRouter.ServeHTTP(w, req)
			return
		}

		fmt.Println("\nNo route matched, proxying request:", req.URL.Path)
		newProbes := probeService(w, req, r.Replicas)

		if newProbes != nil {
			r.ReplicaStatitics.UpdateStatistics(newProbes)
		}

		r.Proxy.ServeHTTP(w, req)
	})

	for _, l := range r.listeners {
		listener, err := l.NewListener()

		if err != nil {
			return err
		}

		srv := &http.Server{Handler: handler}

		r.servers = append(r.servers, srv)

		if l.ServesTLS() {
			go func() {
				if err := srv.ServeTLS(listener, l.TLSCert, l.TLSKey); err != nil && !errors.Is(err, http.ErrServerClosed) {
					log.Println(err)
				}
			}()
		} else {
			go func() {
				if err := srv.Serve(listener); err != nil && !errors.Is(err, http.ErrServerClosed) {
					log.Println(err)
				}
			}()
		}
	}
	return nil
}

func (r *ReverseProxy) Director() func(req *http.Request) {
	return func(req *http.Request) {

		for _, s := range r.Replicas {
			match := &mux.RouteMatch{}

			if s.Router.Match(req, match) {
				upstream := s.SelectUpstream(r.UpstreamDecisionQueue.ProbingToReduceLatencyAndQueuing, r.ProbeQueue)

				// log.Printf("Selected upstream: %v\n", upstream.String())
				r.ReplicaStatitics.IncrementSuccessfulRequests(upstream.String())

				// upstream := s.SelectUpstream(algorithm.RoundRobin)
				fmt.Printf("\nChose upstream %v", upstream)
				req = req.WithContext(context.WithValue(req.Context(), "current_upstream", upstream))
				targetQuery := upstream.RawQuery

				req.URL.Scheme = upstream.Scheme
				req.URL.Host = upstream.Host
				req.URL.Path, req.URL.RawPath = common.JoinURLPath(upstream, req.URL)
				if targetQuery == "" || req.URL.RawQuery == "" {
					req.URL.RawQuery = targetQuery + req.URL.RawQuery
				} else {
					req.URL.RawQuery = targetQuery + "&" + req.URL.RawQuery
				}

				if _, ok := req.Header["User-Agent"]; !ok {
					req.Header.Set("User-Agent", "")
				}
				break
			}
		}
	}
}

func (r *ReverseProxy) ModifyResponse() func(*http.Response) error {
	return func(res *http.Response) error {
		if res.StatusCode == http.StatusInternalServerError || res.StatusCode == http.StatusBadGateway || res.StatusCode == http.StatusRequestTimeout || res.StatusCode == http.StatusGatewayTimeout {
			return fmt.Errorf("upstream returned server error")
		}
		return nil
	}
}

func (r *ReverseProxy) ErrorHandler() func(http.ResponseWriter, *http.Request, error) {
	return func(w http.ResponseWriter, req *http.Request, err error) {
		r.mutex.Lock()
		defer r.mutex.Unlock()

		currentUpstream, ok := req.Context().Value("current_upstream").(*url.URL)
		if !ok || currentUpstream == nil {
			http.Error(w, "unable to determine upstream", http.StatusInternalServerError)
			return
		}

		r.ReplicaStatitics.IncrementFailedRequests(currentUpstream.String())

		if err != nil && err.Error() == "upstream returned 500" {
			if len(r.Replicas[0].Upstreams) == 1 {
				http.Error(w, "upstream returned error or all upstreams failed", http.StatusBadGateway)
				return
			}

			// r.Replicas[0].RemoveUpstream(currentUpstream)
			fmt.Printf("500 Error detected. Retrying with next upstream...")

			req.URL = r.Replicas[0].SelectUpstream(r.UpstreamDecisionQueue.ProbingToReduceLatencyAndQueuing, r.ProbeQueue)
			r.Proxy.ServeHTTP(w, req)
			if w.Header().Get("X-Success") == "true" {
				r.ReplicaStatitics.IncrementSuccessfulRequests(req.URL.String())
				return
			}
		}
		http.Error(w, "Upstream returned error or all upstreams failed", http.StatusBadGateway)
	}
}
