package helpers

import (
	"strings"
)

// PadLeft Adds specified rune to the left of string until length rule is met
func PadLeft(s string, pad rune, length int) string {
	if len(s) >= length {
		return s
	}

	return strings.Repeat(string(pad), length-len(s)) + s
}
