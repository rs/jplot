package window

import (
	"golang.org/x/crypto/ssh/terminal"
)

var sizeFunc = oscSize

func init() {
	if !terminal.IsTerminal(0) {
		// If stdin in not the terminal (because data is piped), fallback to
		// osascript hack.
		osaInit()
		sizeFunc = osaSize
	}
}

// Size returns the size of the current iTerm window.
func Size() (width, height int, err error) {
	return sizeFunc()
}
