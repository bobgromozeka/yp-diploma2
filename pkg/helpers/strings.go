package helpers

import (
	"strings"
)

func PadLeft(s string, pad rune, length int) string {
	if len(s) >= length {
		return s
	}

	return strings.Repeat(string(pad), length-len(s)) + s
}
