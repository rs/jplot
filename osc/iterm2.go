package osc

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"io"
	"sync"
)

// ClearScrollback clears iTerm2 scrollback.
func ClearScrollback() {
	print("\033]1337;ClearScrollback\007")
}

// ImageWriter is a writer that write into iTerm2 terminal the PNG data written
// to it.
type ImageWriter struct {
	Name string

	once   sync.Once
	b66enc io.WriteCloser
	buf    *bytes.Buffer
}

func (w *ImageWriter) init() {
	w.buf = &bytes.Buffer{}
	w.b66enc = base64.NewEncoder(base64.StdEncoding, w.buf)
}

func (w *ImageWriter) Write(p []byte) (n int, err error) {
	w.once.Do(w.init)
	return w.b66enc.Write(p)
}

func (w *ImageWriter) Close() error {
	w.once.Do(w.init)
	fmt.Printf("\033]1337;File=preserveAspectRatio=1;inline=1:%s\007", w.buf.Bytes())
	return w.b66enc.Close()
}
