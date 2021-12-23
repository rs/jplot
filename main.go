package main // import "github.com/rs/jplot"

import (
	"flag"
	"fmt"
	"github.com/rs/jplot/streamer"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/monochromegane/terminal"
	"github.com/rs/jplot/data"
	"github.com/rs/jplot/graph"
	"github.com/rs/jplot/term"
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
		fmt.Fprintln(out, "    - port: launches a web-server that will be drawing images. This disables drawing in terminal")
		fmt.Fprintln(out, "    - hiw (http image width): defines web server's image width")
		fmt.Fprintln(out, "    - hih (http image height): defines web server's image width")
		fmt.Fprintln(out, "    - hirr (http image refresh rate): how often image will be refreshed (milliseconds)")
		fmt.Fprintln(out, "  path:")
		fmt.Fprintln(out, "    JSON field path (eg: field.sub-field).")
	}
	url := flag.String("url", "", "URL to fetch every second. Read JSON objects from stdin if not specified.")
	interval := flag.Duration("interval", time.Second, "When url is provided, defines the interval between fetches."+
		" Note that counter fields are computed based on this interval.")
	steps := flag.Int("steps", 100, "Number of values to plot.")
	rows := flag.Int("rows", 0, "Limits the height of the graph output.")
	port := flag.Int("port", 0, "web server port")
	hiw := flag.Int("hiw", 500, "web server image height")
	hih := flag.Int("hih", 500, "web server  image width")
	hirr := flag.Int("hirr", 1000, "web server refresh rate (milliseconds)")
	flag.Parse()

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
	go func() {
		if er := dp.Run(specs); er != nil {
			fatal("Data source error: ", er)
		}
	}()

	if *port != 0 {
		err := streamer.Start(*port, &dash, *hiw, *hih, *hirr)
		if err != nil {
			fatal("cannot start server: %s", err.Error())
		}
	}

	if !term.HasGraphicsSupport() {
		fatal("iTerm2 or DRCS Sixel graphics required")
	}
	if os.Getenv("TERM") == "screen" {
		fatal("screen and tmux not supported")
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
					term.ClearScrollback()
				}
				term.CursorSavePosition()
				render(dash, *rows)
				term.CursorRestorePosition()
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

}

func fatal(a ...interface{}) {
	fmt.Println(append([]interface{}{"jplot: "}, a...)...)
	os.Exit(1)
}

func prepare(rows int) {
	term.HideCursor()
	if rows == 0 {
		var err error
		if rows, err = term.Rows(); err != nil {
			fatal("Cannot get window size: ", err)
		}
	}
	print(strings.Repeat("\n", rows))
	term.CursorMove(term.Up, rows)
}

func cleanup(rows int) {
	term.ShowCursor()
	if rows == 0 {
		rows, _ = term.Rows()
	}
	term.CursorMove(term.Down, rows)
	print("\n")
}

func render(dash graph.Dash, rows int) {
	size, err := term.Size()
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
	tm := term.NewImageWriter(width, height)
	defer tm.Close()
	if err := dash.Render(tm, width, height); err != nil {
		fatal(fmt.Sprintf("cannot render graph: %v", err.Error()))
	}
}
