package graph

import (
	"image"
	"image/draw"
	"image/png"
	"io"

	"github.com/rs/jplot/data"
	chart "github.com/wcharczuk/go-chart"
)

type Dash struct {
	Specs []data.Spec
	Data  *data.Points
}

// Render generates a PNG with all graphs stacked.
func (d Dash) Render(w io.Writer, width, height int) {
	graphs := make([]chart.Chart, 0, len(d.Specs))
	for _, spec := range d.Specs {
		graphs = append(graphs, New(spec, d.Data, width, height/len(d.Specs)))
	}
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
	png.Encode(w, canvas)
}
