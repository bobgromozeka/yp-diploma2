package helpers

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPadLeft(t *testing.T) {
	type testCase struct {
		name   string
		str    string
		pad    rune
		length int
		want   string
	}

	testCases := []testCase{
		{
			name:   "Success pad",
			str:    "str",
			pad:    'q',
			length: 10,
			want:   "qqqqqqqstr",
		},
		{
			name:   "pad length less then string",
			str:    "strstr",
			pad:    'q',
			length: 5,
			want:   "strstr",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			resStr := PadLeft(tc.str, tc.pad, tc.length)

			assert.Equal(t, tc.want, resStr)
		})
	}
}
