package window

import (
	"fmt"
	"os"
	"time"

	"golang.org/x/crypto/ssh/terminal"
)

// oscSize uses extended iTerm OSC command to get the cell size together with
// the TIOCGWINSZ IOCTL syscall to compute the window size.
func oscSize() (width, height int, err error) {
	cols, lines, err := terminal.GetSize(0)
	if err != nil {
		return
	}
	s, err := terminal.MakeRaw(0)
	if err != nil {
		return
	}
	defer terminal.Restore(0, s)
	fmt.Print("\033]1337;ReportCellSize\033\\")
	fileSetReadDeadline(os.Stdin, time.Now().Add(time.Second))
	defer fileSetReadDeadline(os.Stdin, time.Time{})
	var cellWidth, cellHeight float64
	_, err = fmt.Fscanf(os.Stdin, "\033]1337;ReportCellSize=%f;%f\033\\", &cellHeight, &cellWidth)
	return cols * int(cellWidth), lines * int(cellHeight), err
}
