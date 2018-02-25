package main

import (
	"math"

	chart "github.com/wcharczuk/go-chart"
	"github.com/wcharczuk/go-chart/drawing"
)

// custom version of chart.Legend
func legend(c *chart.Chart, userDefaults ...chart.Style) chart.Renderable {
	return func(r chart.Renderer, cb chart.Box, chartDefaults chart.Style) {
		legendDefaults := chart.Style{
			FillColor:   drawing.ColorWhite,
			FontColor:   chart.DefaultTextColor,
			FontSize:    8.0,
			StrokeColor: chart.DefaultAxisColor,
			StrokeWidth: chart.DefaultAxisLineWidth,
		}

		legendPadding := chart.Box{
			Top:    5,
			Left:   5,
			Right:  5,
			Bottom: 5,
		}
		lineTextGap := 5
		lineLengthMinimum := 25

		var legendStyle chart.Style
		if len(userDefaults) > 0 {
			legendStyle = userDefaults[0].InheritFrom(chartDefaults.InheritFrom(legendDefaults))
		} else {
			legendStyle = chartDefaults.InheritFrom(legendDefaults)
		}

		var labels []string
		var lines []chart.Style
		for _, s := range c.Series {
			if s.GetStyle().IsZero() || s.GetStyle().Show {
				if _, isAnnotationSeries := s.(chart.AnnotationSeries); !isAnnotationSeries {
					labels = append(labels, s.GetName())
					lines = append(lines, s.GetStyle())
				}
			}
		}

		legend := chart.Box{
			Top:  cb.Top,
			Left: cb.Left,
			// bottom and right will be sized by the legend content
		}

		legendContent := chart.Box{
			Top:    legend.Top + legendPadding.Top,
			Left:   legend.Left + legendPadding.Left,
			Right:  legend.Left + legendPadding.Left,
			Bottom: legend.Top + legendPadding.Top,
		}

		legendStyle.GetTextOptions().WriteToRenderer(r)

		// measure
		labelCount := 0
		for x := 0; x < len(labels); x++ {
			if len(labels[x]) > 0 {
				tb := r.MeasureText(labels[x])
				legendContent.Bottom += tb.Height() + 5
				right := legendContent.Left + tb.Width() + lineTextGap + lineLengthMinimum
				legendContent.Right = int(math.Max(float64(legendContent.Right), float64(right)))
				labelCount++
			}
		}

		legend = legend.Grow(legendContent)
		legend.Right = legendContent.Right + legendPadding.Right
		legend.Bottom = legendContent.Bottom + legendPadding.Bottom

		chart.Draw.Box(r, legend, legendStyle)

		legendStyle.GetTextOptions().WriteToRenderer(r)

		ycursor := legendContent.Top
		tx := legendContent.Left
		legendCount := 0
		var label string
		for x := 0; x < len(labels); x++ {
			label = labels[x]
			if len(label) > 0 {
				if legendCount > 0 {
					ycursor += 5
				}

				tb := r.MeasureText(label)

				ty := ycursor + tb.Height()
				r.Text(label, tx, ty)

				th2 := tb.Height() >> 1

				lx := tx + tb.Width() + lineTextGap
				ly := ty - th2
				lx2 := legendContent.Right - legendPadding.Right

				r.SetStrokeColor(lines[x].GetStrokeColor())
				r.SetStrokeWidth(lines[x].GetStrokeWidth())
				r.SetStrokeDashArray(lines[x].GetStrokeDashArray())

				r.MoveTo(lx, ly)
				r.LineTo(lx2, ly)
				r.Stroke()

				ycursor += tb.Height()
				legendCount++
			}
		}
	}
}
