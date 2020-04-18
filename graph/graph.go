package graph

import (
	"fmt"
	"math"

	humanize "github.com/dustin/go-humanize"
	"github.com/rs/jplot/data"
	chart "github.com/wcharczuk/go-chart"
	"github.com/wcharczuk/go-chart/drawing"
)

func init() {
	chart.DefaultBackgroundColor = chart.ColorTransparent
	chart.DefaultCanvasColor = chart.ColorTransparent
	chart.DefaultTextColor = drawing.Color{R: 180, G: 180, B: 180, A: 255}
	chart.DefaultAxisColor = drawing.Color{R: 180, G: 180, B: 180, A: 255}
	chart.DefaultAnnotationFillColor = chart.ColorBlack.WithAlpha(200)
}

// New generate a line graph with series.
func New(spec data.Spec, dp *data.Points, width, height int) chart.Chart {
	series := []chart.Series{}
	markers := []chart.GridLine{}
	for _, f := range spec.Fields {
		vals := dp.Get(f.ID)
		if f.IsMarker {
			for i, v := range vals {
				if v > 0 {
					markers = append(markers, chart.GridLine{Value: float64(i)})
				}
			}
			continue
		}
		series = append(series, chart.ContinuousSeries{
			Name:    fmt.Sprintf("%s: %s", f.Name, siValueFormater(vals[len(vals)-1])),
			YValues: vals,
		})
	}
	return newChart(series, markers, width, height)
}

func newChart(series []chart.Series, markers []chart.GridLine, width, height int) chart.Chart {
	var min, max float64 = math.MaxFloat64, -math.MaxFloat64
	for i, s := range series {
		if s, ok := s.(chart.ContinuousSeries); ok {
			min, max = minMax(s.YValues, min, max)
			s.XValues = chart.LinearRange(0, float64(len(s.YValues)-1))
			c := chart.GetAlternateColor(i + 4)
			s.Style = chart.Style{
				Hidden:      false,
				StrokeWidth: 2,
				StrokeColor: c,
				FillColor:   c.WithAlpha(20),
				FontSize:    9,
			}
			series[i] = s
			last := chart.LastValueAnnotationSeries(s, siValueFormater)
			last.Style.FillColor = c
			last.Style.FontColor = textColor(c)
			last.Style.FontSize = 9
			last.Style.Padding = chart.NewBox(2, 2, 2, 2)
			series = append(series, last)
		}
	}
	graph := chart.Chart{
		Width:  width,
		Height: height,
		Background: chart.Style{
			Padding: chart.NewBox(5, 0, 0, 5),
		},
		YAxis: chart.YAxis{
			Style:          chart.Shown(),
			ValueFormatter: siValueFormater,
		},
		Series: series,
	}
	if min == max {
		// By default, go-chart will fail to render a flat line as the range will be NaN.
		// Define a manual range in such case.
		// See https://github.com/wcharczuk/go-chart/issues/31
		graph.YAxis.Range = &chart.ContinuousRange{
			Min: min - 0.05,
			Max: max + 0.05,
		}
	}
	if len(markers) > 0 {
		graph.Background.Padding.Bottom = 0 // compensate transparent tick space
		graph.XAxis = chart.XAxis{
			Style: chart.Shown(),
			TickStyle: chart.Style{
				StrokeColor: chart.ColorTransparent,
			},
			TickPosition: 10, // hide text with non-existing position
			GridMajorStyle: chart.Style{
				Hidden:          false,
				StrokeColor:     chart.ColorAlternateGray.WithAlpha(100),
				StrokeWidth:     2.0,
				StrokeDashArray: []float64{2.0, 2.0},
			},
			GridLines: markers,
		}
	}
	graph.Elements = []chart.Renderable{
		legend(&graph, chart.Style{
			FillColor:   drawing.Color{A: 100},
			FontColor:   chart.ColorWhite,
			StrokeColor: chart.ColorTransparent,
		}),
	}
	return graph
}

func minMax(values []float64, curMin, curMax float64) (min, max float64) {
	min, max = curMin, curMax
	for _, value := range values {
		if value < min {
			min = value
		}
		if value > max {
			max = value
		}
	}
	return
}

func siValueFormater(v interface{}) string {
	value, prefix := humanize.ComputeSI(v.(float64))
	value = float64(int(value*100)) / 100
	return humanize.Ftoa(value) + " " + prefix
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
