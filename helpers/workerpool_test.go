package helpers_test

// Disabling test due to require go 1.24
// import (
// 	"testing"
// 	"testing/synctest"
// 	"time"

// 	"github.com/qlik-oss/gopherciser/helpers"
// )

// func TestWorkerPool(t *testing.T) {
// 	synctest.Test(t, func(t *testing.T) {
// 		total := 10
// 		concurrency := 2
// 		pool, err := helpers.NewWorkerPool(concurrency, total)
// 		if err != nil {
// 			t.Fatal(err)
// 		}
// 		for i := range total {
// 			if err := pool.AddTask(func() error {
// 				time.Sleep(time.Duration(i) * time.Second)
// 				return nil
// 			}); err != nil {
// 				t.Error(err)
// 			}
// 		}
// 		totalResult := 0
// 		for result := range pool.Results() {
// 			totalResult++
// 			if result != nil {
// 				t.Error(result)
// 			}
// 		}
// 		if totalResult != total {
// 			t.Errorf("total results<%d> expected<%d>", totalResult, total)
// 		}
// 	})

// }
