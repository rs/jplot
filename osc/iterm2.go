package osc

import (
	"bytes"
	"encoding/base64"
	"errors"
	"fmt"
	"image/png"
	"io"
	"os"
	"sync"
	"time"

	"github.com/mattn/go-sixel"

	"golang.org/x/crypto/ssh/terminal"
)

var ecsi = "\033]"
var st = "\007"

var cellSizeOnce sync.Once
var cellWidth, cellHeight float64

var sixelEnabled = false

func init() {
	if os.Getenv("TERM") == "screen" {
		ecsi = "\033Ptmux;\033" + ecsi
		st += "\033\\"
	}
	sixelEnabled = checkSixel()
}

func checkSixel() bool {
	s, err := terminal.MakeRaw(1)
	if err != nil {
		return false
	}
	defer terminal.Restore(1, s)
	_, err = os.Stdout.Write([]byte("\x1b[c"))
	if err != nil {
		return false
	}
	defer fileSetReadDeadline(os.Stdout, time.Time{})

	var b [100]byte
	n, err := os.Stdout.Read(b[:])
	if err != nil {
		return false
	}
	if !bytes.HasPrefix(b[:n], []byte("\x1b[?63;")) {
		return false
	}
	for _, t := range bytes.Split(b[4:n], []byte(";")) {
		if len(t) == 1 && t[0] == '4' {
			return true
		}
	}
	return false
}

func HasGraphicsSupport() bool {
	return os.Getenv("TERM_PROGRAM") == "iTerm.app" || sixelEnabled
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
	if !sixelEnabled {
		fmt.Fprint(os.Stdout, ecsi+"1337;ReportCellSize"+st)
		fileSetReadDeadline(os.Stdout, time.Now().Add(time.Second))
		defer fileSetReadDeadline(os.Stdout, time.Time{})
		fmt.Fscanf(os.Stdout, "\033]1337;ReportCellSize=%f;%f\033\\", &cellHeight, &cellWidth)
	} else {
		// FIXME Way to get sizes from terminal response?
		cellWidth, cellHeight = 8, 8
	}
}

// Size gathers sizing information of the current session's controling terminal.
func Size() (size TermSize, err error) {
	size.Col, size.Row, err = terminal.GetSize(1)
	if err != nil {
		return
	}
	cellSizeOnce.Do(initCellSize)
	if cellWidth+cellHeight == 0 {
		err = errors.New("cannot get terminal cell size")
	}
	size.Width, size.Height = size.Col*int(cellWidth), size.Row*int(cellHeight)
	return
}

// Rows returns the number of rows for the controling terminal.
func Rows() (rows int, err error) {
	_, rows, err = terminal.GetSize(1)
	return
}

func NewImageWriter() io.WriteCloser {
	if !sixelEnabled {
		return &imageWriter{}
	}
	return &sixelWriter{}
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

type sixelWriter struct {
	once sync.Once
	enc  *sixel.Encoder
	buf  *bytes.Buffer
}

func (w *sixelWriter) init() {
	w.buf = &bytes.Buffer{}
	w.enc = sixel.NewEncoder(os.Stdout)
}

// Write writes the PNG image data into the imageWriter buffer.
func (w *sixelWriter) Write(p []byte) (n int, err error) {
	w.once.Do(w.init)
	return w.buf.Write(p)
}

// Close flushes the image to the terminal and close the writer.
func (w *sixelWriter) Close() error {
	w.once.Do(w.init)
	img, err := png.Decode(w.buf)
	if err != nil {
		return err
	}
	return w.enc.Encode(img)
}
