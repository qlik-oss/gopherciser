package randomizer

import (
	"testing"
	"time"
)

var (
	intArray    = []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	stringArray = []string{"zero", "one", "one", "two", "three", "five", "eight"}
)

func TestRandSeeding(t *testing.T) {
	t.Parallel()

	seed := GetPredictableSeed(1, 1)
	if seed != 5217487135812689397 {
		t.Errorf("Expected seed 5217487135812689397 got %d", seed)
	}

	rnd := NewSeededRandomizer(seed)
	if rnd == nil {
		t.Fatal("Expected seeded randomizer object, got nil")
	}

	if duration, err := rnd.RandDuration(1, 200000); err != nil {
		t.Error(err)
	} else {
		if duration != 63663 {
			t.Errorf("Expected duration 63.663µs got %v", duration)
		}
	}

	if i, err := rnd.RandInt(intArray); err != nil {
		t.Error(err)
	} else {
		if i != 7 {
			t.Errorf("Expected int 7, got %d", i)
		}
	}

	if str, err := rnd.RandString(stringArray); err != nil {
		t.Error(err)
	} else {
		if str != "one" {
			t.Errorf("Expected one, got %v", str)
		}
	}

	if i := rnd.Rand(10); i != 4 {
		t.Errorf("Expected 4, got %v", i)
	}

	if i, err := rnd.RandWeightedInt([]int{3, 5, 7}); err != nil {
		t.Error(err)
	} else if i != 2 {
		t.Errorf("Expected 2, got %v", i)
	}

	expectedBytes := []byte{133, 232, 143, 244, 92, 168, 110, 69}
	if bytes, err := rnd.RandBytes(8); err != nil {
		t.Error(err)
	} else if len(bytes) != len(expectedBytes) {
		t.Errorf("Expected %v got %v", expectedBytes, bytes)
	} else {
		for i, v := range expectedBytes {
			if v != bytes[i] {
				t.Errorf("Expected %v got %v", expectedBytes, bytes)
				break
			}
		}
	}
}

type testWeightedStruct struct {
	MyKey    string
	MyWeight int
}

func (t testWeightedStruct) GetKey() string {
	return t.MyKey
}
func (t testWeightedStruct) GetWeight() int {
	return t.MyWeight
}

func TestRandWeighted(t *testing.T) {
	t.Parallel()

	seed := GetPredictableSeed(1, 1)
	if seed != 5217487135812689397 {
		t.Errorf("Expected seed 5217487135812689397 got %d", seed)
	}

	rnd := NewSeededRandomizer(seed)
	if rnd == nil {
		t.Fatal("Expected seeded randomizer object, got nil")
	}

	origArray := []testWeightedStruct{
		{
			MyKey:    "firstKey",
			MyWeight: 2,
		},
		{
			MyKey:    "SecondKey",
			MyWeight: 3,
		},
	}

	weightedInterface := make([]Weighted, len(origArray))
	for i, v := range origArray {
		weightedInterface[i] = v
	}

	entry, err := rnd.RandWeighted(weightedInterface)
	if err != nil {
		t.Fatal("RandWeighted Failed:", err)
	}

	if entry != "SecondKey" {
		t.Errorf("RandWeighted: Expected SecondKey, got %v", entry)
	}
}

func TestRand(t *testing.T) {
	t.Parallel()

	rnd := NewRandomizer()
	if rnd == nil {
		t.Fatal("Expected randomizer object, got nil")
	}

	if _, err := rnd.RandDuration(1, 200000); err != nil {
		t.Error(err)
	}

	if duration, err := rnd.RandDuration(616, 616); err != nil {
		t.Error(err)
	} else if duration != time.Duration(616) {
		t.Errorf("Expected duration 616ns got %v", duration)
	}

	if _, err := rnd.RandInt(intArray); err != nil {
		t.Error(err)
	}

	if _, err := rnd.RandString(stringArray); err != nil {
		t.Error(err)
	}

	if _, err := rnd.RandBytes(16); err != nil {
		t.Error(err)
	}
}

func TestNegative(t *testing.T) {
	t.Parallel()

	rnd := NewRandomizer()
	if rnd == nil {
		t.Fatal("Expected randomizer object, got nil")
		return // make linter not warn
	}

	if _, err := rnd.RandDuration(6, 2); err == nil {
		t.Error("Expected min > max error, got nothing")
	}

	if _, err := rnd.RandInt([]int{}); err == nil {
		t.Error("Expected empty integer array error, got nothing")
	}

	if _, err := rnd.RandString([]string{}); err == nil {
		t.Error("Expected empty string array error, got nothing")
	}

	rnd.r = nil
	if _, err := rnd.RandInt(intArray); err == nil {
		t.Error("Expected rand not initilized error, got nothing")
	}

	if _, err := rnd.RandString(stringArray); err == nil {
		t.Error("Expected rand not initilized error, got nothing")
	}
}

type wght struct {
	key    string
	weight int
}

func (w wght) GetKey() string {
	return w.key
}

func (w wght) GetWeight() int {
	return w.weight
}

func TestWeight(t *testing.T) {
	t.Parallel()

	rnd := NewSeededRandomizer(GetPredictableSeed(2, 42))
	one := 0
	thirty := 0
	sixtynine := 0

	wgthed := []Weighted{
		wght{
			key:    "one",
			weight: 1,
		},
		wght{
			key:    "thirty",
			weight: 30,
		},
		wght{
			key:    "sixtynine",
			weight: 69,
		},
	}

	iterations := 1000000

	for i := 0; i < iterations; i++ {
		v, err := rnd.RandWeighted(wgthed)
		if err != nil {
			t.Fatalf("Error randomizing weighted, err: %v", err)
		}

		switch v {
		case "one":
			one++
		case "thirty":
			thirty++
		case "sixtynine":
			sixtynine++
		}
	}

	if one != 10058 {
		t.Errorf("Unexpected random hits for 1%%, expected<10058/1000000> have<%d/%d>", one, iterations)
	}
	if thirty != 299347 {
		t.Errorf("Unexpected random hits for 30%%, expected<299347/1000000> have<%d/%d>", thirty, iterations)
	}
	if sixtynine != 690595 {
		t.Errorf("Unexpected random hits for 69%%, expected<690595/1000000> have<%d/%d>", sixtynine, iterations)
	}

	t.Logf("Weigted test\nTotal: %d\nOne: %d\nThirty: %d\nSixtynine: %d\n", iterations, one, thirty, sixtynine)
}

func TestPredicableSeed(t *testing.T) {
	t.Parallel()

	seed := GetPredictableSeed(2, 6)

	expected := int64(2734078272839172190)
	if seed != expected {
		t.Errorf("Unexpected seed<%d> expected<%d>", seed, expected)
	}

	maxuint64 := ^uint64(0)
	biginstance := maxuint64 - uint64(maxuint64/8)
	bigsession := maxuint64 - uint64(maxuint64/2)
	seed = GetPredictableSeedUInt64(biginstance, bigsession)
	expected = int64(3103636859239040441)
	if seed != expected {
		t.Errorf("Unexpected seed<%d> expected<%d>", seed, expected)
	}
}
