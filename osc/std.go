package osc

import "fmt"

type Direction string

const (
	// Up moves the cursor up.
	Up Direction = "A"
	// Down moves the cursor down.
	Down Direction = "B"
	// Forward moves the cursor foward.
	Forward Direction = "C"
	// Backward moves the cursor backward.
	Backward Direction = "D"
	// NextLine cursor to beginning of the line next line.
	NextLine Direction = "E"
	// PreviousLine cursor to beginning of the line previous line.
	PreviousLine Direction = "F"
	// HorizontalAbsolute the cursor to the specified column.
	HorizontalAbsolute Direction = "G"
)

const csi = "\033["

// HideCursor hides cursor.
func HideCursor() {
	print(csi + "?25l")
}

// ShowCursor shows cursor.
func ShowCursor() {
	print(csi + "?25h")
}

// Clear clears the screen.
func Clear() {
	print(csi + "H\033[2J")
}

// CursorPosition moves cursor to row, col.
func CursorPosition(row, col int) {
	fmt.Printf("%s%d;%dH", csi, row, col)
}

// CursorMove moves the cursor n times in the direction d.
func CursorMove(d Direction, n uint) {
	fmt.Printf("%s%d%s", csi, n, d)
}
