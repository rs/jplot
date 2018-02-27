package window

import (
	"bytes"
	"fmt"
	"log"
	"os/exec"
	"strconv"
	"strings"
)

// Gets current front ID.
var getID = `
tell application "iTerm"
	set w to front window
	set b to bounds of w
	do shell script "echo " & (id of w)
end tell`

// Gets the size of the current front window.
var getSize = `
tell application "iTerm"
	set w to window id %s
	set b to bounds of w
	set width to ((item 3 of b) - (item 1 of b)) as text
	set height to ((item 4 of b) - (item 2 of b)) as text
	do shell script "echo " & " " & width & " " & height
end tell`

var winID string

func init() {
	var err error
	winID, err = runScript(getID)
	if err != nil {
		log.Fatalf("Cannot get current window ID: %v", err)
	}
}

func runScript(script string) (string, error) {
	cmd := exec.Command("osascript", "-e", script)
	var out bytes.Buffer
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("osascript error: %v", err)
	}
	return strings.TrimSpace(out.String()), nil
}

// Size returns current window height and width.
func Size() (int, int, error) {
	out, err := runScript(fmt.Sprintf(getSize, winID))
	if err != nil {
		return 0, 0, err
	}
	rv := strings.Split(out, " ")
	if len(rv) != 2 {
		return 0, 0, fmt.Errorf("invalid output: %s", out)
	}
	width, _ := strconv.Atoi(rv[0])
	height, _ := strconv.Atoi(rv[1])
	return width, height, nil
}
