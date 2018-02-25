package source

import (
	"bufio"
	"os"

	"github.com/elgs/gojq"
)

type Stdin struct {
	scan *bufio.Scanner
}

func NewStdin() Stdin {
	return Stdin{bufio.NewScanner(os.Stdin)}
}

func (s Stdin) Get() (*gojq.JQ, error) {
	if s.scan.Scan() {
		return gojq.NewStringQuery(s.scan.Text())
	}
	return nil, s.scan.Err()
}

func (s Stdin) Close() error {
	return nil
}
