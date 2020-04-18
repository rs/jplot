// +build !go1.10

package term

import (
	"os"
	"time"
)

func fileSetReadDeadline(f *os.File, t time.Time) error {
	return nil
}
