package osc

import (
	"bytes"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"os"
	"sync"
	"time"

	"golang.org/x/crypto/ssh/terminal"
)

var ecsi = "\033]"
var st = "\007"

var cellSizeOnce sync.Once
var cellWidth, cellHeight float64

func init() {
	if os.Getenv("TERM") == "screen" {
		ecsi = "\033Ptmux;\033" + ecsi
		st += "\033\\"
	}
}

// ClearScrollback clears iTerm2 scrollback.
func ClearScrollback() {
	print(ecsi + "1337;ClearScrollback" + st)
}

// TermSize contains sizing information of the terminal.
type TermSize struct {
	Row    int
	Col    int
	Width  int
	Height int
}

func initCellSize() {
	s, err := terminal.MakeRaw(1)
	if err != nil {
		return
	}
	defer terminal.Restore(1, s)
	fmt.Fprint(os.Stdout, ecsi+"1337;ReportCellSize"+st)
	fileSetReadDeadline(os.Stdout, time.Now().Add(time.Second))
	defer fileSetReadDeadline(os.Stdout, time.Time{})
	fmt.Fscanf(os.Stdout, "\033]1337;ReportCellSize=%f;%f\033\\", &cellHeight, &cellWidth)
}

// Size gathers sizing information of the current session's controling terminal.
func Size() (size TermSize, err error) {
	size.Col, size.Row, err = terminal.GetSize(1)
	if err != nil {
		return
	}
	cellSizeOnce.Do(initCellSize)
	if cellWidth+cellHeight == 0 {
		err = errors.New("cannot get iTerm2 cell size")
	}
	size.Width, size.Height = size.Col*int(cellWidth), size.Row*int(cellHeight)
	return
}

// Rows returns the number of rows for the controling terminal.
func Rows() (rows int, err error) {
	_, rows, err = terminal.GetSize(1)
	return
}

// ImageWriter is a writer that write into iTerm2 terminal the PNG data written
// to it.
type ImageWriter struct {
	Name   string
	Width  int
	Height int

	once   sync.Once
	b66enc io.WriteCloser
	buf    *bytes.Buffer
}

func (w *ImageWriter) init() {
	w.buf = &bytes.Buffer{}
	w.b66enc = base64.NewEncoder(base64.StdEncoding, w.buf)
}

// Write writes the PNG image data into the ImageWriter buffer.
func (w *ImageWriter) Write(p []byte) (n int, err error) {
	w.once.Do(w.init)
	return w.b66enc.Write(p)
}

// Close flushes the image to the terminal and close the writer.
func (w *ImageWriter) Close() error {
	w.once.Do(w.init)
	fmt.Printf("%s1337;File=preserveAspectRatio=1;width=%dpx;height=%dpx;inline=1:%s%s", ecsi, w.Width, w.Height, w.buf.Bytes(), st)
	return w.b66enc.Close()
}
