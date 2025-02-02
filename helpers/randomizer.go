package helpers

import "time"

type Randomizer interface {
	Rand(max int) int
	RandWeightedInt(weights []int) (int, error)
	RandIntPos(ints []int) (int, int, error)
	RandDuration(minDuration, maxDuration time.Duration) (time.Duration, error)
	Reset(instance, session uint64, onlyinstanceSeed bool)
	RandRune(runes []rune) rune
	// Float64 returns, as a float64, a pseudo-random number in the half-open interval [0.0,1.0).
	Float64() float64
}
