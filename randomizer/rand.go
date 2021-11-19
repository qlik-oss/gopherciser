package randomizer

import (
	"math/rand"
	"time"

	"github.com/pkg/errors"
)

//Randomizer overlord randomizer
type Randomizer struct {
	r    *rand.Rand
	seed int64
}

// Weighted Interface used by random weighted functions
type Weighted interface {
	GetKey() string
	GetWeight() int
}

//NewSeededRandomizer new seeded randomizer
func NewSeededRandomizer(seed int64) *Randomizer {
	return newRnd(seed)
}

//NewRandomizer new randomizer seeded with time.now
func NewRandomizer() *Randomizer {
	return newRnd(time.Now().UnixNano())
}

func newRnd(seed int64) *Randomizer {
	r := rand.New(rand.NewSource(seed))
	return &Randomizer{
		r:    r,
		seed: seed,
	}
}

//GetPredictableSeed Get predictable seed for instance session combination,
//to be used for randomizers in testing and trend measurements
func GetPredictableSeed(instance, session int) int64 {
	if session < 0 {
		session = -session
	}
	if instance < 0 {
		instance = -instance
	}
	return GetPredictableSeedUInt64(uint64(instance), uint64(session))
}

//GetPredictableSeedUInt64 Get predictable seed for instance session combination,
//to be used for randomizers in testing and trend measurements
func GetPredictableSeedUInt64(instance, session uint64) int64 {

	//overflow protected primal seed calculation
	uInstance := instance % (^uint64(0) / 10000) //remainder of max uint64 to protect against overflow when multiplying  with 10000
	if session > ^uint64(0)-uInstance {          //max uint64 - instance to protect against overflow when adding session
		session = session - uInstance
	}
	innerSeed := int64(((uInstance * 10000) + session) % (^uint64(0) >> 1)) //remainder of max int64 to protect against overflow when converting to int64
	seeder := rand.New(rand.NewSource(innerSeed))

	return seeder.Int63()
}

//RandString return random string from list
func (rnd *Randomizer) RandString(strings []string) (string, error) {
	if len(strings) < 1 {
		return "", errors.New("Empty string array")
	}

	if rnd.r == nil {
		return "", errors.New("rand not initilized")
	}

	return strings[rnd.r.Intn(len(strings))], nil
}

//RandInt return random int from list
func (rnd *Randomizer) RandInt(ints []int) (int, error) {
	value, _, err := rnd.RandIntPos(ints)
	return value, err
}

// RandRune from rune array
func (rnd *Randomizer) RandRune(runes []rune) rune {
	pos := rnd.r.Intn(len(runes))
	return runes[pos]
}

//RandIntPos return random int value and position from list (value, pos)
func (rnd *Randomizer) RandIntPos(ints []int) (int, int, error) {
	if len(ints) < 1 {
		return -1, -1, errors.New("Empty integer array")
	}

	if rnd.r == nil {
		return -1, -1, errors.New("rand not initialized")
	}

	pos := rnd.r.Intn(len(ints))
	return ints[pos], pos, nil
}

//RandDuration random duration up to max
func (rnd *Randomizer) RandDuration(minDuration, maxDuration time.Duration) (time.Duration, error) {
	if minDuration == maxDuration {
		return minDuration, nil
	}

	if minDuration > maxDuration {
		return 0, errors.Errorf("Min duration(%d) greater than max(%d)", minDuration, maxDuration)
	}

	return time.Duration(rnd.r.Int63n(int64(maxDuration)-int64(minDuration)) + int64(minDuration)), nil
}

//Rand32 returns result from Int32n using current randomizer instance
func (rnd *Randomizer) Rand32(max int32) int32 {
	return rnd.r.Int31n(max)
}

//Rand returns result from Intn using current randomizer instance
func (rnd *Randomizer) Rand(max int) int {
	return rnd.r.Intn(max)
}

//RandWeightedInt randomize based on weight
func (rnd *Randomizer) RandWeightedInt(weights []int) (int, error) {
	sumWeights := 0
	for _, v := range weights {
		sumWeights += v
	}
	rVal := rnd.r.Intn(sumWeights)

	for i, v := range weights {
		if rVal < v {
			return i, nil
		}
		rVal -= v
	}
	return -1, errors.Errorf("Failed to randomize from weight %v", weights)
}

//RandWeighted randomize using Weighted interface
func (rnd *Randomizer) RandWeighted(w []Weighted) (string, error) {
	sumWeights := 0
	for _, v := range w {
		sumWeights += v.GetWeight()
	}

	if sumWeights <= 0 {
		return "", errors.Errorf("Illegal sum of weights: %v", sumWeights)
	}

	rVal := rnd.r.Intn(sumWeights)

	for _, v := range w {
		if rVal < v.GetWeight() {
			return v.GetKey(), nil
		}
		rVal -= v.GetWeight()
	}

	return "", errors.Errorf("Failed to randomize from weight %v", w)
}

//RandBytes random byte array
func (rnd *Randomizer) RandBytes(size int) ([]byte, error) {
	bytes := make([]byte, size)
	if readSize, err := rnd.r.Read(bytes); err != nil {
		return nil, errors.WithStack(err)
	} else if readSize != size {
		return nil, errors.Errorf("Randomized byte array size of incorrect length, expected<%d> got<%d>",
			size, readSize)
	}
	return bytes, nil
}
