package helpers

import "math"

const (
	DefaultEpsilon = 0.000000001
)

var (
	FloatMinNormal = math.Float64frombits(0x00800000)
)

func NearlyEqual(a, b float64) bool {
	return NearlyEqualEpsilon(a, b, DefaultEpsilon)
}

// NearlyEqual float points covering some edge cases as per https://floating-point-gui.de/errors/comparison/#look-out-for-edge-cases
func NearlyEqualEpsilon(a, b float64, epsilon float64) bool {
	absA := math.Abs(a)
	absB := math.Abs(b)
	diff := math.Abs(a - b)

	if a == b { // shortcut, handles infinities
		return true
	} else if a == 0 || b == 0 || absA+absB < FloatMinNormal {
		// a or b is zero or both are extremely close to it
		// relative error is less meaningful here
		return diff < (epsilon * FloatMinNormal)
	} else { // use relative error
		return diff/math.Min((absA+absB), math.MaxFloat64) < epsilon
	}
}
