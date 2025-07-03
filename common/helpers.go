package common

import (
	"log"
	"math"
	"math/rand/v2"
	"net/url"
	"path"
)

func RandomRound(num float64) int {
	if rand.IntN(2) == 0 {
		return int(math.Ceil(num))
	} else {
		return int(math.Floor(num))
	}
}

func JoinURLPath(base, u *url.URL) (joinedPath, rawJoinedPath string) {

	joinedPath = path.Join(base.Path, u.Path)

	if u.RawPath != "" {
		rawJoinedPath = path.Join(base.Path, u.RawPath)
	} else {
		rawJoinedPath = joinedPath
	}

	if base.Path == "" && len(joinedPath) > 0 && joinedPath[0] != '/' {
		joinedPath = "/" + joinedPath
	}

	if base.Path == "" && len(rawJoinedPath) > 0 && rawJoinedPath[0] != '/' {
		rawJoinedPath = "/" + rawJoinedPath
	}

	return joinedPath, rawJoinedPath
}

func FailOnError(err error, msg string) {
	if err != nil {
		log.Panicf("%s: %s", msg, err)
	}
}
