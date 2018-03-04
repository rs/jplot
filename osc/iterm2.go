package osc

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"io"
	"os"
	"sync"
	"time"

	"golang.org/x/crypto/ssh/terminal"
)

// ClearScrollback clears iTerm2 scrollback.
func ClearScrollback() {
	print("\033]1337;ClearScrollback\007")
}

// TermSize contains sizing information of the terminal.
type TermSize struct {
	Row    int
	Col    int
	Width  int
	Height int
}

// Size gathers sizing information of the current session's controling terminal.
func Size() (size TermSize, err error) {
	size.Col, size.Row, err = terminal.GetSize(1)
	if err != nil {
		return
	}
	s, err := terminal.MakeRaw(1)
	if err != nil {
		return
	}
	defer terminal.Restore(1, s)
	var cellWidth, cellHeight float64
	fmt.Fprint(os.Stdout, "\033]1337;ReportCellSize\033\\")
	fileSetReadDeadline(os.Stdout, time.Now().Add(time.Second))
	defer fileSetReadDeadline(os.Stdout, time.Time{})
	_, err = fmt.Fscanf(os.Stdout, "\033]1337;ReportCellSize=%f;%f\033\\", &cellHeight, &cellWidth)
	size.Width, size.Height = size.Col*int(cellWidth), size.Row*int(cellHeight)
	return
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
