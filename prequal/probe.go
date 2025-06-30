package prequal

import (
	"net/url"
	"time"
)

type PrequalParameters struct {
	MaxLifeTime       time.Duration
	PoolSize          int
	ProbeFactor       float64
	ProbeRemoveFactor int
	TotalReplica      int
	Mu                int
	Denominator       int
	ReuseRate         int
}

type ServerProbe struct {
	Name             string
	RequestsInFlight uint64
	Latency          uint64
	Upstream         *url.URL
	LastTenLatency   []uint64
}

type ServerProbeItem struct {
	used        int
	ReceiptTime time.Time
	ServerProbe
}

func NewPrequalParameters(maxLifeTime int, poolSize int, probeFactor float64, probeRemoveFactor int, totalReplica int, mu int) *PrequalParameters {
	return &PrequalParameters{
		MaxLifeTime:       time.Duration(maxLifeTime) * time.Second,
		PoolSize:          poolSize,
		ProbeFactor:       probeFactor,
		ProbeRemoveFactor: probeRemoveFactor,
		TotalReplica:      totalReplica,
		Mu:                mu,
		Denominator:       (1-poolSize/totalReplica)*int(probeFactor) - probeRemoveFactor,
		ReuseRate:         max(1, ((1+mu)/(1-poolSize/totalReplica)*int(probeFactor) - probeRemoveFactor)),
	}
}

func NewServerProbe(s *ProbeResponse, u *url.URL) *ServerProbe {
	return &ServerProbe{
		Name:             s.ServerName,
		RequestsInFlight: s.RequestsInFlight,
		Latency:          s.Latency,
		LastTenLatency:   s.LastTenLatency,
		Upstream:         u,
	}
}

func NewServerProbeItem(s *ServerProbe) *ServerProbeItem {
	return &ServerProbeItem{
		ServerProbe: *s,
		used:        0,
		ReceiptTime: time.Now(),
	}
}
