package helpers_test

import (
	"testing"

	"github.com/qlik-oss/gopherciser/helpers"
)

func TestWorkerPool(t *testing.T) {
	total := 10
	concurrency := 2
	pool, err := helpers.NewWorkerPool(concurrency, total)
	if err != nil {
		t.Fatal(err)
	}
	for range total {
		if err := pool.AddTask(func() error {
			return nil
		}); err != nil {
			t.Error(err)
		}
	}
	totalResult := 0
	for result := range pool.Results() {
		totalResult++
		if result != nil {
			t.Error(result)
		}
	}
	if totalResult != total {
		t.Errorf("total results<%d> expected<%d>", totalResult, total)
	}
}
