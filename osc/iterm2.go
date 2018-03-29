package osc

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"io"
	"os"
	"sync"
)

func init() {
	if os.Getenv("TERM") == "screen" {
		ecsi = "\033Ptmux;\033" + ecsi
		st += "\033\\"
	}
}

// imageWriter is a writer that write into iTerm2 terminal the PNG data written
// to it.
type imageWriter struct {
	Name string

	once   sync.Once
	b66enc io.WriteCloser
	buf    *bytes.Buffer
}

func (w *imageWriter) init() {
	w.buf = &bytes.Buffer{}
	w.b66enc = base64.NewEncoder(base64.StdEncoding, w.buf)
}

// Write writes the PNG image data into the imageWriter buffer.
func (w *imageWriter) Write(p []byte) (n int, err error) {
	w.once.Do(w.init)
	return w.b66enc.Write(p)
}

// Close flushes the image to the terminal and close the writer.
func (w *imageWriter) Close() error {
	w.once.Do(w.init)
	fmt.Printf("%s1337;File=preserveAspectRatio=1;inline=1:%s%s", ecsi, w.buf.Bytes(), st)
	return w.b66enc.Close()
}
