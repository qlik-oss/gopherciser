package precisiontime

import "time"

// Tick Current clock tick via
func Tick() (int64, error) {
	// https://go-review.googlesource.com/c/go/+/67332 so it will predominantly work in newer OS versions
	var now = time.Now().UnixNano()
	return now, nil
}
