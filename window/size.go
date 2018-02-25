package window

import (
	"bytes"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
)

// Gets the size of the current front
var script = `
tell application "iTerm"
	--set w to front window
	--set w to window id 1887
	set w to %s
	set b to bounds of w
	set width to ((item 3 of b) - (item 1 of b)) as text
	set height to ((item 4 of b) - (item 2 of b)) as text
	do shell script "echo " & (id of w) & " " & width & " " & height
end tell`

var winID string

// Size returns current window height and width.
func Size() (int, int, error) {
	win := "front window"
	if winID != "" {
		win = "window id " + winID
	}
	cmd := exec.Command("osascript", "-e", fmt.Sprintf(script, win))
	cmd.Stdin = strings.NewReader("some input")
	var out bytes.Buffer
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		return 0, 0, fmt.Errorf("Cannot execute osascript: %v", err)
	}
	rv := strings.Split(strings.TrimSpace(out.String()), " ")
	if len(rv) != 3 {
		return 0, 0, fmt.Errorf("invalid output: %s", out.String())
	}
	winID = rv[0]
	width, _ := strconv.Atoi(rv[1])
	height, _ := strconv.Atoi(rv[2])
	return width, height, nil
}
