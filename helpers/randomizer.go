package helpers

import "time"

type Randomizer interface {
	Rand(max int) int
	RandWeightedInt(weights []int) (int, error)
	RandIntPos(ints []int) (int, int, error)
	RandDuration(minDuration, maxDuration time.Duration) (time.Duration, error)
	Reset(instance, session uint64, onlyinstanceSeed bool)
}
