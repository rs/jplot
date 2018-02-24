package main // import "github.com/rs/jplot"

import (
	"bufio"
	"bytes"
	"encoding/base64"
	"flag"
	"fmt"
	"image"
	"image/draw"
	"image/png"
	"log"
	"os"
	"strings"

	humanize "github.com/dustin/go-humanize"
	"github.com/elgs/gojq"
	"github.com/wcharczuk/go-chart"
	"github.com/wcharczuk/go-chart/drawing"
	"github.com/wcharczuk/go-chart/seq"
)

func main() {
	steps := flag.Int("steps", 100, "Number of values to plot.")
	width := flag.Int("width", 2500, "Canvas width")
	height := flag.Int("height", 1300, "Canvas height")
	dpi := flag.Float64("dpi", 220, "Canvas definition")
	flag.Parse()
	fields := flag.Args()

	dp := &dataPoints{Steps: *steps}

	clear()
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		jq, err := gojq.NewStringQuery(scanner.Text())
		if err != nil {
			log.Fatalf("Input error: %v", err)
		}
		graphs := make([]chart.Chart, 0, len(fields))
		for i, field := range fields {
			series := []chart.Series{}
			for j, f := range strings.Split(field, "+") {
				var isCounter bool
				name := f
				if strings.HasPrefix(name, "counter:") {
					isCounter = true
					name = name[8:]
				}
				v, err := jq.Query(name)
				if err != nil {
					log.Fatalf("Cannot get %s: %v", name, err)
				}
				n, ok := v.(float64)
				if !ok {
					log.Fatalf("Invalid type %s: %T", name, v)
				}
				vals := dp.push(fmt.Sprintf("%d.%d.%s", i, j, name), n, isCounter)
				series = append(series, chart.ContinuousSeries{
					Name:    fmt.Sprintf("%s: %s", name, humanize.SI(vals[len(vals)-1], "")),
					YValues: vals,
				})
			}
			graphs = append(graphs, graph(series, *width, *height/len(fields), *dpi))
		}
		printGraphs(graphs, *width, *height)
	}
}

type dataPoints struct {
	Steps  int
	points map[string][]float64
	last   map[string]float64
}

func (dp *dataPoints) push(name string, value float64, counter bool) []float64 {
	if dp.points == nil {
		dp.points = make(map[string][]float64, 1)
		dp.last = make(map[string]float64)
	}
	d, found := dp.points[name]
	if !found {
		d = make([]float64, dp.Steps)
	}
	if counter {
		var diff float64
		if last := dp.last[name]; last > 0 {
			diff = value - last
		}
		dp.last[name] = value
		value = diff
	}
	d = append(append(make([]float64, 0, dp.Steps), d[1:]...), value)
	dp.points[name] = d
	return d
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
func printGraphs(graphs []chart.Chart, width, height int) {
	reset()
	canvas := image.NewRGBA(image.Rectangle{image.Point{0, 0}, image.Point{width, height}})
	graphHeight := height / len(graphs)
	for i, graph := range graphs {
		iw := &chart.ImageWriter{}
		graph.Render(chart.PNG, iw)
		img, _ := iw.Image()
		r := image.Rectangle{image.Point{0, graphHeight * i}, image.Point{width, graphHeight * (i + 1)}}
		draw.Draw(canvas, r, img, image.Point{0, 0}, draw.Src)
	}
	var b bytes.Buffer
	enc := base64.NewEncoder(base64.StdEncoding, &b)
	defer enc.Close()
	png.Encode(enc, canvas)
	fmt.Printf("\033]1337;File=preserveAspectRatio=1;inline=1:%s\007", b.Bytes())
}
