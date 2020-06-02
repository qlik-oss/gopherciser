package helpers

import (
	"sort"
	"testing"
)

func TestUniqueInts(t *testing.T) {
	m := make(UniqueInts)
	a := m.Array()
	if m.Array() == nil {
		t.Fatalf("uniqueInts array was nil")
	}
	validateIntArray(t, "uniqueInts", a, []int{})

	m.AddValue(5)
	m.AddValue(5)
	m.AddValue(22)
	a = m.Array()
	if a == nil {
		t.Fatalf("uniqueInts array was nil")
	}

	validateIntArraySortable(t, "uniqueInts", a, []int{5, 22}, true)

	if !m.HasValue(5) {
		t.Error("uniqueInts collection doesn't contain expected value: 5")
	}

	if m.HasValue(10) {
		t.Error("uniqueInts collection contains unexpected value: 10")
	}
}

func validateIntArray(t *testing.T, key string, value, expected []int) {
	t.Helper()

	validateIntArraySortable(t, key, value, expected, false)
}

func validateIntArraySortable(t *testing.T, key string, value, expected []int, doSort bool) {
	t.Helper()

	if len(value) != len(expected) {
		t.Errorf("%s: Unexpected array length<%d> expected<%d> array<%+v>", key, len(value), len(expected), value)
		return
	}

	if doSort { //Order not important
		sort.Ints(value)
		sort.Ints(expected)
	}

	for i, v := range value {
		if v != expected[i] {
			t.Errorf("%s: Unexpected value<%d> pos<%d> expected<%d> array<%+v>", key, v, i, expected[i], value)
			return
		}
	}
}
