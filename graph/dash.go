package graph

import (
	"image"
	"image/draw"
	"image/png"
	"io"

	"github.com/rs/jplot/data"
	chart "github.com/wcharczuk/go-chart/v2"
)

type Dash struct {
	Specs []data.Spec
	Data  *data.Points
}

// Render generates a PNG with all graphs stacked.
func (d Dash) Render(w io.Writer, width, height int) error {
	graphs := make([]chart.Chart, 0, len(d.Specs))
	for _, spec := range d.Specs {
		graphs = append(graphs, New(spec, d.Data, width, height/len(d.Specs)))
	}
	canvas := image.NewRGBA(image.Rectangle{image.Point{0, 0}, image.Point{width, height}})
	var top int
	for _, graph := range graphs {
		iw := &chart.ImageWriter{}
		if err := graph.Render(chart.PNG, iw); err != nil {
			return err
		}
		img, _ := iw.Image()
		r := image.Rectangle{image.Point{0, top}, image.Point{width, top + graph.Height}}
		top += graph.Height
		draw.Draw(canvas, r, img, image.Point{0, 0}, draw.Src)
	}
	return png.Encode(w, canvas)
}
