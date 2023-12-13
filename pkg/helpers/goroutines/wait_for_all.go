package goroutines

import (
	"sync"
)

type mxErrors struct {
	mx   sync.Mutex
	data []error
}

// WaitForAll waits for all specified goroutines finish. Returns slices of all occurred errors.
func WaitForAll(fns ...func() error) []error {
	waitLen := len(fns)

	var (
		wg     sync.WaitGroup
		errors mxErrors
	)

	wg.Add(waitLen)

	for _, fn := range fns {
		go func(fn func() error) {
			defer wg.Done()

			if err := fn(); err != nil {
				errors.mx.Lock()
				errors.data = append(errors.data, err)
				errors.mx.Unlock()
			}
		}(fn)
	}

	wg.Wait()

	return errors.data
}
