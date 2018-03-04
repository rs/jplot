// +build !go1.10

package osc

import (
	"os"
	"time"
)

func fileSetReadDeadline(f *os.File, t time.Time) error {
	return nil
}
