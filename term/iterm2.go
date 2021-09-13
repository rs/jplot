package term

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
type imageWriter struct {
	Name   string
	Width  int
	Height int

	once   sync.Once
	b64enc io.WriteCloser
	buf    *bytes.Buffer
}

func (w *imageWriter) init() {
	w.buf = &bytes.Buffer{}
	w.b64enc = base64.NewEncoder(base64.StdEncoding, w.buf)
}

// Write writes the PNG image data into the imageWriter buffer.
func (w *imageWriter) Write(p []byte) (n int, err error) {
	w.once.Do(w.init)
	return w.b64enc.Write(p)
}

// Close flushes the image to the terminal and close the writer.
func (w *imageWriter) Close() error {
	w.once.Do(w.init)
	fmt.Printf("%s1337;File=preserveAspectRatio=1;width=%dpx;height=%dpx;inline=1:%s%s", ecsi, w.Width, w.Height, w.buf.Bytes(), st)
	return w.b64enc.Close()
}
