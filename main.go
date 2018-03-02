package main // import "github.com/rs/jplot"

import (
	"bytes"
	"encoding/base64"
	"flag"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/rs/jplot/data"
	"github.com/rs/jplot/graph"
	"github.com/rs/jplot/window"
)

const titleBarSize = 25

func main() {
	url := flag.String("url", "", "URL to fetch every second. Read JSON objects from stdin if not specified.")
	interval := flag.Duration("interval", time.Second, "When url is provided, defines the interval between fetches."+
		" Note that counter fields are computed based on this interval.")
	steps := flag.Int("steps", 100, "Number of values to plot.")
	flag.Parse()

	specs, err := data.ParseSpec(flag.Args())
	if err != nil {
		log.Fatalf("Cannot parse spec: %v", err)
	}
	var dp *data.Points
	if *url != "" {
		dp = data.FromHTTP(*url, *interval, *steps)
	} else {
		dp = data.FromStdin(*steps)
	}
	defer dp.Close()
	dash := graph.Dash{
		Specs: specs,
		Data:  dp,
	}

	wg := &sync.WaitGroup{}
	wg.Add(1)
	defer wg.Wait()
	exit := make(chan struct{})
	defer close(exit)
	go func() {
		defer wg.Done()
		clear()
		t := time.NewTicker(time.Second)
		defer t.Stop()
		i := 0
		for {
			width, height, err := window.Size()
			if err != nil {
				log.Fatalf("Cannot get window size: %v", err)
			}
			select {
			case <-t.C:
				i++
				if i%120 == 0 {
					// Clear scrollback to avoid iTerm from eating all the memory.
					cleanup()
				}
				render(dash, width, height-titleBarSize)
			case <-exit:
				render(dash, width, height-titleBarSize)
				return
			}
		}
	}()

	if err := dp.Run(specs); err != nil {
		log.Fatalf("Data source error: %v", err)
	}
}

func render(dash graph.Dash, width, height int) {
	var b bytes.Buffer
	enc := base64.NewEncoder(base64.StdEncoding, &b)
	defer enc.Close()
	dash.Render(enc, width, height)
	reset()
	// Use iTerm2 image display feature.
	fmt.Printf("\033]1337;File=preserveAspectRatio=1;inline=1:%s\007", b.Bytes())
}

func clear() {
	print("\033\133\110\033\133\062\112") // clear screen
	print("\033]1337;CursorShape=1\007")  // set cursor to vertical bar
}

func reset() {
	print("\033\133\061\073\061\110") // move cursor to 0x0
}

func cleanup() {
	print("\033]1337;ClearScrollback\007")
}
