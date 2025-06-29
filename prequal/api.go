package prequal

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
)

type ProbeResponse struct {
	ServerName       string   `json:"serverName"`
	RequestsInFlight uint64   `json:"requestInFlight"`
	Latency          uint64   `json:"latency"`
	LastTenLatency   []uint64 `json:"last_ten_latency"`
}

func getProbe(replica *Replica, idx int) (*ServerProbe, error) {
	replica.Lock.RLock()
	defer replica.Lock.RUnlock()
	url := replica.Upstreams[idx]

	res, err := http.Get(fmt.Sprintf("%s://%s/%s", url.Scheme, url.Host, "ping"))
	if err != nil {
		log.Printf("error making get request %v", err)
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		log.Printf("received non-200 response: %d", res.StatusCode)
		return nil, err
	}

	body, err := io.ReadAll(res.Body)

	if err != nil {
		log.Printf("error reading response body %v", err)
		return nil, err
	}

	var probeRes ProbeResponse

	err = json.Unmarshal([]byte(body), &probeRes)
	if err != nil {
		log.Printf("error parsing JSON: %v", err)
		return nil, err
	}

	newProbe := NewServerProbe(&probeRes, url)
	return newProbe, nil
}
