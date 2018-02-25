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
	"math"
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
	marker  bool
}

func main() {
	url := flag.String("url", "", "URL to fetch every second. Read JSON objects from stdin if not specified.")
	steps := flag.Int("steps", 100, "Number of values to plot.")
	width := flag.Int("width", 2500, "Canvas width")
	height := flag.Int("height", 1400, "Canvas height")
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
			var isMarker bool
			n := strings.Split(name, ":")
			for len(n) > 1 {
				switch n[0] {
				case "counter":
					isCounter = true
				case "marker":
					isMarker = true
				default:
					log.Fatalf("Invalid field option: %s", n[0])
				}
				n = n[1:]
			}
			name = n[0]
			if strings.HasPrefix(name, "counter:") {
				isCounter = true
				name = name[8:]
			}
			gs.fields = append(gs.fields, field{
				id:      fmt.Sprintf("%d.%d.%s", i, j, name),
				name:    name,
				counter: isCounter,
				marker:  isMarker,
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
		markers := []chart.GridLine{}
		for _, f := range gs.fields {
			vals := dp.Get(f.id)
			if f.marker {
				for i, v := range vals {
					if v > 0 {
						markers = append(markers, chart.GridLine{Value: float64(i)})
					}
				}
				continue
			}
			series = append(series, chart.ContinuousSeries{
				Name:    fmt.Sprintf("%s: %s", f.name, siValueFormater(vals[len(vals)-1])),
				YValues: vals,
			})
		}
		graphs = append(graphs, graph(series, markers, gs.width, gs.height, gs.dpi))
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
func graph(series []chart.Series, markers []chart.GridLine, width, height int, dpi float64) chart.Chart {
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
			max := &chart.MaxSeries{
				Style: chart.Style{
					Show:            true,
					StrokeColor:     c,
					StrokeDashArray: []float64{5.0, 5.0},
				},
				InnerSeries: s,
			}
			last := chart.LastValueAnnotation(s, siValueFormater)
			last.Style.FillColor = c
			last.Style.FontColor = textColor(c)
			series = append(series, max, last)
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
			Style:          chart.StyleShow(),
			ValueFormatter: siValueFormater,
		},
		Series: series,
	}
	if len(markers) > 0 {
		graph.Background.Padding.Bottom = 0 // compensate transparent tick space
		graph.XAxis = chart.XAxis{
			Style: chart.StyleShow(),
			TickStyle: chart.Style{
				StrokeColor: chart.ColorTransparent,
			},
			TickPosition: 10, // hide text with non-existing position
			GridMajorStyle: chart.Style{
				Show:            true,
				StrokeColor:     chart.ColorAlternateGray,
				StrokeWidth:     3.0,
				StrokeDashArray: []float64{5.0, 5.0},
			},
			GridLines: markers,
		}
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

func textColor(bg drawing.Color) drawing.Color {
	var L float64
	for c, f := range map[uint8]float64{bg.R: 0.2126, bg.G: 0.7152, bg.B: 0.0722} {
		c := float64(c) / 255.0
		if c <= 0.03928 {
			c = c / 12.92
		} else {
			c = math.Pow(((c + 0.055) / 1.055), 2.4)
		}
		L += c * f
	}
	if L > 0.179 {
		return chart.ColorBlack
	}
	return chart.ColorWhite
}

func siValueFormater(v interface{}) string {
	value, prefix := humanize.ComputeSI(v.(float64))
	value = float64(int(value*100)) / 100
	return humanize.Ftoa(value) + " " + prefix
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
