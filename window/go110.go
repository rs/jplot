// +build go1.10

package window

import (
	"os"
	"time"
)

func fileSetReadDeadline(f *os.File, t time.Time) error {
	return f.SetReadDeadline(t)
}
