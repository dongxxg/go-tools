package util

import (
	"math/rand"
	"time"
)

func init() {
	seed := time.Now().Unix()
	rand.Seed(seed)
}

func RandomNum(min, max int) int {
	return min + rand.Intn(max-min+1)
}
