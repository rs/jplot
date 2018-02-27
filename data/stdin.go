package data

import (
	"bufio"
	"os"

	"github.com/elgs/gojq"
)

type stdin struct {
	scan *bufio.Scanner
}

// FromStdin reads data from stdin as one JSON object per line.
func FromStdin(size int) *Points {
	return &Points{
		Size:   size,
		Source: stdin{bufio.NewScanner(os.Stdin)},
	}
}

func (s stdin) Get() (*gojq.JQ, error) {
	if s.scan.Scan() {
		return gojq.NewStringQuery(s.scan.Text())
	}
	return nil, s.scan.Err()
}

func (s stdin) Close() error {
	return nil
}
