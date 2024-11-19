package algorithm

import (
	"net/url"

	"math/rand"

	"github.com/ra-shree/prequal-server/common"
)

func RandomDChoice(r *common.Replica) *url.URL {
	random := rand.Intn(len(r.Upstreams))

	return r.Upstreams[random]
}
