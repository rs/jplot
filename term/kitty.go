package term

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"os"
	"sync"
	"time"

	"golang.org/x/crypto/ssh/terminal"
)

var kittyEnabled = false
var kittyCurrentID = 1     // Track current image ID globally
var kittyFirstImage = true // Track if this is the first image

func init() {
	if os.Getenv("TERM_PROGRAM") != "iTerm.app" {
		kittyEnabled = checkKitty()
	}
}

// checkKitty detects if the terminal supports Kitty graphics protocol
func checkKitty() bool {
	s, err := terminal.MakeRaw(1)
	if err != nil {
		return false
	}
	defer terminal.Restore(1, s)

	// Send Kitty graphics query to check if graphics are supported
	// Use a simple query that Kitty will respond to if it supports graphics
	_, err = os.Stdout.Write([]byte("\033_Gi=1,a=q,t=d,f=24\033\\"))
	if err != nil {
		return false
	}

	fileSetReadDeadline(os.Stdout, time.Now().Add(500*time.Millisecond))
	defer fileSetReadDeadline(os.Stdout, time.Time{})

	var b [200]byte
	n, err := os.Stdout.Read(b[:])
	if err != nil {
		return false
	}

	// Look for Kitty's graphics response
	// Kitty will respond with \033_Gi=1;...\033\ if it supports graphics
	response := b[:n]
	return bytes.Contains(response, []byte("_G")) && bytes.Contains(response, []byte("i=1"))
}

// kittyWriter is a writer that displays PNG images in Kitty terminal using the Kitty graphics protocol
type kittyWriter struct {
	Name   string
	Width  int
	Height int

	once sync.Once
	buf  *bytes.Buffer
}

func (w *kittyWriter) init() {
	w.buf = &bytes.Buffer{}
}

// Write writes the PNG image data into the kittyWriter buffer.
func (w *kittyWriter) Write(p []byte) (n int, err error) {
	w.once.Do(w.init)
	return w.buf.Write(p)
}

// Close flushes the image to the terminal using Kitty's graphics protocol and closes the writer.
func (w *kittyWriter) Close() error {
	w.once.Do(w.init)

	// Encode the image data as base64
	b64data := base64.StdEncoding.EncodeToString(w.buf.Bytes())

	// Calculate next image ID (alternate between 1 and 2)
	nextID := 3 - kittyCurrentID

	// Step 1: Display the new image with absolute positioning
	// a=T (transmit and display), f=100 (PNG), t=d (direct), i=nextID (image ID),
	// X=0,Y=0 (absolute position), q=2 (suppress responses)
	fmt.Printf("\033_Ga=T,f=100,t=d,i=%d,X=0,Y=0,q=2;%s\033\\", nextID, b64data)

	// Step 2: Delete the previously displayed image (skip deletion only on very first image)
	// Remove q=2 from delete to ensure it executes properly
	if !kittyFirstImage {
		fmt.Printf("\033_Ga=d,d=i,i=%d,q=2\033\\", kittyCurrentID)
	}
	kittyFirstImage = false // After first image, we always delete

	// Update current ID for next time
	kittyCurrentID = nextID

	return nil
}
