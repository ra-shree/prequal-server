package reverseproxy

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/gorilla/mux"
	"github.com/ra-shree/prequal-server/pkg/algorithm"
	"github.com/ra-shree/prequal-server/pkg/common"
)

type ReverseProxy struct {
	listeners []Listener
	Proxy     *httputil.ReverseProxy
	servers   []*http.Server
	Replicas  []*common.Replica
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

func (r *ReverseProxy) Start(probeService service) error {
	r.Proxy = &httputil.ReverseProxy{
		Director: r.Director(),
	}

	handler := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		probeService(w, req, r.Replicas)
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
				upstream := s.SelectUpstream(algorithm.RandomDChoice)

				fmt.Printf("chose upstream %v", upstream)

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
