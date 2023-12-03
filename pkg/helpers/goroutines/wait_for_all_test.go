package goroutines

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWaitForAll(t *testing.T) {
	t.Parallel()

	var result string

	err2 := fmt.Errorf("err 2")

	errs := WaitForAll(
		func() error {
			result = "res 1"
			return nil
		},
		func() error {
			return err2
		},
	)

	assert.Equal(t, "res 1", result)
	assert.Equal(t, err2, errs[0])
}
