package source

import (
	"io"

	"github.com/elgs/gojq"
)

type Getter interface {
	io.Closer
	Get() (*gojq.JQ, error)
}
