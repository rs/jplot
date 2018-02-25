package main // import "github.com/rs/jplot"

import (
	"bytes"
	"encoding/base64"
	"flag"
	"fmt"
	"image"
	"image/draw"
	"image/png"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/rs/jplot/data"
	"github.com/rs/jplot/source"

	humanize "github.com/dustin/go-humanize"
	"github.com/wcharczuk/go-chart"
	"github.com/wcharczuk/go-chart/drawing"
	"github.com/wcharczuk/go-chart/seq"
)

type graphSpec struct {
	width, height int
	dpi           float64
	fields        []field
}

type field struct {
	id      string
	name    string
	counter bool
}

func main() {
	url := flag.String("url", "", "URL to fetch every second. Read JSON objects from stdin if not specified.")
	steps := flag.Int("steps", 100, "Number of values to plot.")
	width := flag.Int("width", 2500, "Canvas width")
	height := flag.Int("height", 1300, "Canvas height")
	dpi := flag.Float64("dpi", 220, "Canvas definition")
	flag.Parse()

	specs := parseSpec(flag.Args(), *width, *height, *dpi)

	dp := &data.Points{Size: *steps}
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
		for {
			select {
			case <-t.C:
				render(specs, dp)
			case <-exit:
				render(specs, dp)
				return
			}
		}
	}()

	var s source.Getter = source.NewStdin()
	if *url != "" {
		s = source.NewHTTP(*url, time.Second)
	}
	defer s.Close()
	for {
		jq, err := s.Get()
		if err != nil {
			log.Fatalf("Input error: %v", err)
		}
		if jq == nil {
			break
		}
		for _, gs := range specs {
			for _, f := range gs.fields {
				v, err := jq.Query(f.name)
				if err != nil {
					log.Fatalf("Cannot get %s: %v", f.name, err)
				}
				n, ok := v.(float64)
				if !ok {
					log.Fatalf("Invalid type %s: %T", f.name, v)
				}
				dp.Push(f.id, n, f.counter)
			}
		}
	}
}

func parseSpec(args []string, width, height int, dpi float64) []graphSpec {
	specs := make([]graphSpec, 0, len(args))
	for i, v := range flag.Args() {
		gs := graphSpec{
			width:  width,
			height: height / len(args),
			dpi:    dpi,
		}
		for j, name := range strings.Split(v, "+") {
			var isCounter bool
			if strings.HasPrefix(name, "counter:") {
				isCounter = true
				name = name[8:]
			}
			gs.fields = append(gs.fields, field{
				id:      fmt.Sprintf("%d.%d.%s", i, j, name),
				name:    name,
				counter: isCounter,
			})
		}
		specs = append(specs, gs)
	}
	return specs
}

func render(specs []graphSpec, dp *data.Points) {
	graphs := make([]chart.Chart, 0, len(specs))
	for _, gs := range specs {
		series := []chart.Series{}
		for _, f := range gs.fields {
			vals := dp.Get(f.id)
			series = append(series, chart.ContinuousSeries{
				Name:    fmt.Sprintf("%s: %s", f.name, humanize.SI(vals[len(vals)-1], "")),
				YValues: vals,
			})
		}
		graphs = append(graphs, graph(series, gs.width, gs.height, gs.dpi))
	}
	printGraphs(graphs)
}

func init() {
	chart.DefaultBackgroundColor = chart.ColorTransparent
	chart.DefaultCanvasColor = chart.ColorTransparent
	chart.DefaultTextColor = drawing.Color{R: 180, G: 180, B: 180, A: 255}
	chart.DefaultAxisColor = drawing.Color{R: 180, G: 180, B: 180, A: 255}
	chart.DefaultAnnotationFillColor = chart.ColorBlack.WithAlpha(200)
}

func clear() {
	print("\033\133\110\033\133\062\112") // clear screen
	print("\033]1337;CursorShape=1\007")  // set cursor to vertical bar
}

func reset() {
	print("\033\133\061\073\061\110") // move cursor to 0x0
}

// graph generate a line graph with series.
func graph(series []chart.Series, width, height int, dpi float64) chart.Chart {
	for i, s := range series {
		if s, ok := s.(chart.ContinuousSeries); ok {
			s.XValues = seq.Range(0, float64(len(s.YValues)-1))
			c := chart.GetAlternateColor(i + 4)
			s.Style = chart.Style{
				Show:        true,
				StrokeWidth: 2,
				StrokeColor: c,
				FillColor:   c.WithAlpha(20),
			}
			series[i] = s
		}
	}
	graph := chart.Chart{
		Width:  width,
		Height: height,
		DPI:    dpi,
		Background: chart.Style{
			Padding: chart.NewBox(20, 0, 0, 20),
		},
		YAxis: chart.YAxis{
			Style: chart.StyleShow(),
			ValueFormatter: func(v interface{}) string {
				return humanize.SI(v.(float64), "")
			},
		},
		Series: series,
	}
	graph.Elements = []chart.Renderable{
		chart.Legend(&graph, chart.Style{
			FillColor:   drawing.Color{A: 100},
			FontColor:   chart.ColorWhite,
			StrokeColor: chart.ColorTransparent,
		}),
	}
	return graph
}

// printGraphs generates a single PNG with graphs stacked and print it to iTerm2.
func printGraphs(graphs []chart.Chart) {
	var width, height int
	for _, graph := range graphs {
		if graph.Width > width {
			width = graph.Width
		}
		height += graph.Height
	}
	reset()
	canvas := image.NewRGBA(image.Rectangle{image.Point{0, 0}, image.Point{width, height}})
	var top int
	for _, graph := range graphs {
		iw := &chart.ImageWriter{}
		graph.Render(chart.PNG, iw)
		img, _ := iw.Image()
		r := image.Rectangle{image.Point{0, top}, image.Point{width, top + graph.Height}}
		top += graph.Height
		draw.Draw(canvas, r, img, image.Point{0, 0}, draw.Src)
	}
	var b bytes.Buffer
	enc := base64.NewEncoder(base64.StdEncoding, &b)
	defer enc.Close()
	png.Encode(enc, canvas)
	fmt.Printf("\033]1337;File=preserveAspectRatio=1;inline=1:%s\007", b.Bytes())
}
