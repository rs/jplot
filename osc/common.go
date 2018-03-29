package osc

import (
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
