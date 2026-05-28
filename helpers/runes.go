package helpers

import "unicode/utf8"

func TruncateString(val string, maxLength int) (string, bool) {
	if utf8.RuneCountInString(val) > maxLength {
		return string([]rune(val)[:maxLength]) + "...", true
	}
	return val, false
}
