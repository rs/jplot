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
var termWidth, termHeight int

func HasGraphicsSupport() bool {
	return os.Getenv("TERM_PROGRAM") == "iTerm.app" || sixelEnabled
}

// ClearScrollback clears iTerm2 scrollback.
func ClearScrollback() {
	if !sixelEnabled {
		print(ecsi + "1337;ClearScrollback" + st)
	}
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
		fmt.Fprint(os.Stdout, "\033[14t")
		fileSetReadDeadline(os.Stdout, time.Now().Add(time.Second))
		defer fileSetReadDeadline(os.Stdout, time.Time{})
		fmt.Fscanf(os.Stdout, "\033[4;%d;%dt", &termHeight, &termWidth)
	}
}

// Size gathers sizing information of the current session's controling terminal.
func Size() (size TermSize, err error) {
	size.Col, size.Row, err = terminal.GetSize(1)
	if err != nil {
		return
	}
	cellSizeOnce.Do(initCellSize)
	if termWidth > 0 && termHeight > 0 {
		size.Width = int(termWidth/(size.Col-1)) * (size.Col - 1)
		size.Height = int(termHeight/(size.Row-1)) * (size.Row - 1)
		return
	} else if cellWidth+cellHeight == 0 {
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
