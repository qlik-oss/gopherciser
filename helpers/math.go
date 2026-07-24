package helpers

import "cmp"

// Min returns the least of a and b
func Min[T cmp.Ordered](a, b T) T {
	if a < b {
		return a
	}
	return b
}

// Max returns the greatest of a and b
func Max[T cmp.Ordered](a, b T) T {
	if a > b {
		return a
	}
	return b
}
