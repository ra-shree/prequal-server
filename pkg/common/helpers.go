package common

import (
	"math"
	"math/rand/v2"
)

func randomRound(num float64) int {
	if rand.IntN(2) == 0 {
		return int(math.Ceil(num))
	} else {
		return int(math.Floor(num))
	}
}
