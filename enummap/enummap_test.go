package enummap

import (
	"fmt"
	"testing"
)

type (
	myEnum int
)

const (
	zero myEnum = iota
	one
	two
)

func TestIllegalKey(t *testing.T) {
	_, err := NewEnumMap(map[string]int{
		"legal":   0,
		"ILLEGAL": 1,
	})
	if err == nil {
		t.Error("expected NewEnumMap to throw error on illegal key")
	}
}

func TestEnumMap(t *testing.T) {
	asInt := map[string]int{
		"zero": int(zero),
		"one":  int(one),
		"two":  int(two),
	}
	em, err := NewEnumMap(asInt)
	if err != nil {
		t.Fatal(err)
	}

	if err = assertMap(em); err != nil {
		t.Error(err)
	}
}

func TestEnumMapAdd(t *testing.T) {
	em := New()
	myMap := map[string]myEnum{
		"zero": zero,
		"one":  one,
		"two":  two,
	}

	for k, v := range myMap {
		if err := em.Add(k, int(v)); err != nil {
			t.Fatal(err)
		}
	}

	if err := assertMap(em); err != nil {
		t.Error(err)
	}
}

func TestUnMarshal(t *testing.T) {
	arg := []byte("1")

	em, err := NewEnumMap(map[string]int{
		"zero": int(zero),
		"one":  int(one),
		"two":  int(two),
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("enummap<%+v>", em)

	i, err := em.UnMarshal(arg)
	if err != nil {
		t.Fatal(err)
	}
	if i != 1 {
		t.Errorf("unexpected<%d> integer, expected<1>", i)
	}

	arg = []byte(`"two"`)
	i, err = em.UnMarshal(arg)
	if err != nil {
		t.Fatal(err)
	}
	if i != 2 {
		t.Errorf("unexpected integer<%d>, expected<2>", i)
	}
}

func assertMap(em *EnumMap) error {
	if i, err := em.Int("one"); err != nil {
		return err
	} else if myEnum(i) != one {
		return fmt.Errorf("one != 1")
	}

	if s, err := em.String(int(two)); err != nil {
		return err
	} else if s != "two" {
		return fmt.Errorf("2 != two")
	}

	if _, err := em.Int("three"); err == nil {
		return fmt.Errorf("expected StringKeyNotFound error")
	} else if err != StringKeyNotFoundError("three") {
		return fmt.Errorf("expected StringKeyNotFound error")
	}

	if _, err := em.String(3); err == nil {
		return fmt.Errorf("expected IntKeyNotFound error")
	} else if err != IntKeyNotFoundError(3) {
		return fmt.Errorf("expected IntKeyNotFound error")
	}

	return nil
}
