package utils

import (
	"github.com/lucasepe/codename"

	"math/rand"
)

var r *rand.Rand

func init() {
	rng, err := codename.DefaultRNG()
	if err != nil {
		panic(err)
	}

	r = rng

}

func RandomName() string {
	return codename.Generate(r, 4)
}
