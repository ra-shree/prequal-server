package reverseproxy

import (
	"errors"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/gorilla/mux"
	"github.com/ra-shree/prequal-server/pkg/algorithm"
	"github.com/ra-shree/prequal-server/pkg/common"
	"github.com/ra-shree/prequal-server/utils"
)

type ReverseProxy struct {
	listeners []Listener
	proxy     *httputil.ReverseProxy
	servers   []*http.Server
	replicas  []*common.Replica
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

	r.replicas = append(r.replicas, &common.Replica{
		Router:    router,
		Upstreams: urls,
	})

	return nil
}

func (r *ReverseProxy) Start() error {
	r.proxy = &httputil.ReverseProxy{
		Director: r.Director(),
	}

	for _, l := range r.listeners {
		listener, err := l.NewListener()

		if err != nil {
			return err
		}

		srv := &http.Server{Handler: r.proxy}

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
		for _, s := range r.replicas {
			match := &mux.RouteMatch{}

			if s.Router.Match(req, match) {
				upstream := s.SelectUpstream(algorithm.RandomDChoice)
				targetQuery := upstream.RawQuery

				req.URL.Scheme = upstream.Scheme
				req.URL.Host = upstream.Host
				req.URL.Path, req.URL.RawPath = utils.JoinURLPath(upstream, req.URL)
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
