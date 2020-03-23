package statistics

import "testing"

func TestErrWarn(t *testing.T) {
	ew := ErrWarn{}

	ew.IncWarn()
	w := ew.Warnings()
	if w != 1 {
		t.Errorf("incorrect warning count<%d>, expected<1>", w)
	}

	w = ew.TotWarnings()
	if w != 1 {
		t.Errorf("incorrect total warning count<%d>, expected<1>", w)
	}
}
