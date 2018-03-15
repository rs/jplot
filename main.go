package main // import "github.com/rs/jplot"

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/monochromegane/terminal"
	"github.com/rs/jplot/data"
	"github.com/rs/jplot/graph"
	"github.com/rs/jplot/osc"
)

func main() {
	flag.Usage = func() {
		out := os.Stderr
		fmt.Fprintln(out, "Usage: jplot [OPTIONS] FIELD_SPEC [FIELD_SPEC...]:")
		fmt.Fprintln(out, "")
		fmt.Fprintln(out, "OPTIONS:")
		flag.PrintDefaults()
		fmt.Fprintln(out, "")
		fmt.Fprintln(out, "FIELD_SPEC: [<option>[,<option>...]:]path")
		fmt.Fprintln(out, "  option:")
		fmt.Fprintln(out, "    - counter: Computes the difference with the last value. The value must increase monotonically.")
		fmt.Fprintln(out, "    - marker: When the value is none-zero, a vertical line is drawn.")
		fmt.Fprintln(out, "  path:")
		fmt.Fprintln(out, "    JSON field path (eg: field.sub-field).")
	}
	url := flag.String("url", "", "URL to fetch every second. Read JSON objects from stdin if not specified.")
	interval := flag.Duration("interval", time.Second, "When url is provided, defines the interval between fetches."+
		" Note that counter fields are computed based on this interval.")
	steps := flag.Int("steps", 100, "Number of values to plot.")
	rows := flag.Int("rows", 0, "Limits the height of the graph output.")
	flag.Parse()

	if os.Getenv("TERM_PROGRAM") != "iTerm.app" {
		fatal("iTerm2 required")
	}
	if os.Getenv("TERM") == "screen" {
		fatal("screen and tmux not supported")
	}

	if len(flag.Args()) == 0 {
		flag.Usage()
		os.Exit(1)
	}

	specs, err := data.ParseSpec(flag.Args())
	if err != nil {
		fatal("Cannot parse spec: ", err)
	}
	var dp *data.Points
	if *url != "" {
		dp = data.FromHTTP(*url, *interval, *steps)
	} else if !terminal.IsTerminal(os.Stdin) {
		dp = data.FromStdin(*steps)
	} else {
		fatal("neither --url nor stdin is provided")
	}
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
		t := time.NewTicker(time.Second)
		defer t.Stop()
		c := make(chan os.Signal, 2)
		signal.Notify(c, os.Interrupt, syscall.SIGTERM)
		i := 0
		for {
			select {
			case <-t.C:
				if i == 0 {
					prepare(*rows)
					defer cleanup(*rows)
				}
				i++
				if i%120 == 0 {
					// Clear scrollback to avoid iTerm from eating all the memory.
					osc.ClearScrollback()
				}
				osc.CursorSavePosition()
				render(dash, *rows)
				osc.CursorRestorePosition()
			case <-exit:
				if i == 0 {
					render(dash, *rows)
				}
				return
			case <-c:
				dp.Close()
				signal.Stop(c)
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

func prepare(rows int) {
	osc.HideCursor()
	if rows == 0 {
		var err error
		if rows, err = osc.Rows(); err != nil {
			fatal("Cannot get window size: ", err)
		}
	}
	print(strings.Repeat("\n", rows))
	osc.CursorMove(osc.Up, rows)
}

func cleanup(rows int) {
	osc.ShowCursor()
	if rows == 0 {
		rows, _ = osc.Rows()
	}
	osc.CursorMove(osc.Down, rows)
	print("\n")
}

func render(dash graph.Dash, rows int) {
	size, err := osc.Size()
	if err != nil {
		fatal("Cannot get window size: ", err)
	}
	width, height := size.Width, size.Height
	if rows > 0 {
		height = size.Height / size.Row * rows
	} else {
		rows = size.Row
	}
	// Use iTerm2 image display feature.
	term := &osc.ImageWriter{}
	defer term.Close()
	if err := dash.Render(term, width, height); err != nil {
		fatal(fmt.Sprintf("cannot render graph: %v", err.Error()))
	}
}
