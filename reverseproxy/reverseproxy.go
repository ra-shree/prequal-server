package reverseproxy

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"sync"

	"github.com/gorilla/mux"
	"github.com/ra-shree/prequal-server/algorithm"
	"github.com/ra-shree/prequal-server/common"
)

type ReverseProxy struct {
	listeners []Listener
	Proxy     *httputil.ReverseProxy
	servers   []*http.Server
	Replicas  []*common.Replica
	mutex     sync.RWMutex
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

type service func(http.ResponseWriter, *http.Request, []*common.Replica)

func (r *ReverseProxy) AddReplica(upstreams []string, router *mux.Router) error {
	var urls = []*url.URL{}

	for _, upstream := range upstreams {
		url, err := url.Parse(upstream)

		if err != nil {
			return err
		}

		if router == nil {
			router = mux.NewRouter()
			router.PathPrefix("/")
		}

		urls = append(urls, url)
	}

	r.Replicas = append(r.Replicas, &common.Replica{
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

func (r *ReverseProxy) Start(probeService service) error {
	r.Proxy = &httputil.ReverseProxy{
		Director:       r.Director(),
		ErrorHandler:   r.ErrorHandler(),
		ModifyResponse: r.ModifyResponse(),
	}

	handler := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		probeService(w, req, r.Replicas)
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
		if req.URL.Path == "/admin" || strings.HasPrefix(req.URL.Path, "/admin/") {
			req.URL.Scheme = "http"
			req.URL.Host = "localhost:8080"

			if _, ok := req.Header["User-Agent"]; !ok {
				req.Header.Set("User-Agent", "")
			}
			return
		}

		for _, s := range r.Replicas {
			match := &mux.RouteMatch{}

			if s.Router.Match(req, match) {
				upstream := s.SelectUpstream(algorithm.ProbingToReduceLatencyAndQueuing)
				// upstream := s.SelectUpstream(algorithm.RoundRobin)
				fmt.Printf("\nChose upstream %v\n\n", upstream)
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

		if err != nil && err.Error() == "upstream returned 500" {
			if len(r.Replicas[0].Upstreams) == 1 {
				http.Error(w, "upstream returned error or all upstreams failed", http.StatusBadGateway)
				return
			}

			r.Replicas[0].RemoveUpstream(currentUpstream)
			fmt.Printf("500 Error detected. Retrying with next upstream...")

			req.URL = r.Replicas[0].SelectUpstream(algorithm.ProbingToReduceLatencyAndQueuing)
			r.Proxy.ServeHTTP(w, req)
			if w.Header().Get("X-Success") == "true" {
				return
			}
		}
		http.Error(w, "Upstream returned error or all upstreams failed", http.StatusBadGateway)
	}
}
