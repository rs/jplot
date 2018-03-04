package main // import "github.com/rs/jplot"

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/rs/jplot/data"
	"github.com/rs/jplot/graph"
	"github.com/rs/jplot/osc"
)

func main() {
	url := flag.String("url", "", "URL to fetch every second. Read JSON objects from stdin if not specified.")
	interval := flag.Duration("interval", time.Second, "When url is provided, defines the interval between fetches."+
		" Note that counter fields are computed based on this interval.")
	steps := flag.Int("steps", 100, "Number of values to plot.")
	flag.Parse()

	if os.Getenv("TERM_PROGRAM") != "iTerm.app" {
		fatal("iTerm2 required")
	}
	if os.Getenv("TERM") == "screen" {
		fatal("screen and tmux not supported")
	}

	specs, err := data.ParseSpec(flag.Args())
	if err != nil {
		fatal("Cannot parse spec: ", err)
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
		osc.Clear()
		osc.HideCursor()
		defer osc.ShowCursor()
		t := time.NewTicker(time.Second)
		defer t.Stop()
		c := make(chan os.Signal, 2)
		signal.Notify(c, os.Interrupt, syscall.SIGTERM)
		i := 0
		for {
			size, err := osc.Size()
			if err != nil {
				fatal("Cannot get window size: ", err)
			}
			select {
			case <-t.C:
				i++
				if i%120 == 0 {
					// Clear scrollback to avoid iTerm from eating all the memory.
					osc.ClearScrollback()
				}
				render(dash, size.Width, size.Height)
			case <-exit:
				render(dash, size.Width, size.Height)
				return
			case <-c:
				osc.ShowCursor()
				os.Exit(0)
			}
		}
	}()

	if err := dp.Run(specs); err != nil {
		fatal("Data source error: ", err)
	}
}

func fatal(a ...interface{}) {
	fmt.Println(append([]interface{}{"jplot: "}, a...)...)
	os.Exit(1)
}

func render(dash graph.Dash, width, height int) {
	osc.CursorPosition(0, 0)
	// Use iTerm2 image display feature.
	term := &osc.ImageWriter{}
	defer term.Close()
	dash.Render(term, width, height)
}
