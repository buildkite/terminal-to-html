//go:build !unix

package rusage

import (
	"fmt"
	"runtime"
)

// Stats reports a "not implemented" error.
func Stats() (*Resources, error) {
	return nil, fmt.Errorf("not implemented for %s", runtime.GOOS)
}
