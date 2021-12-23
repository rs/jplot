// Copyright 2013 Sonia Keys.
// Licensed under MIT license.  See "license" file in this source tree.

package quant

import (
	"image"
	"image/color"
	"image/draw"
	"math"
)

// Sierra24A satisfies draw.Drawer
type Sierra24A struct{}

var _ draw.Drawer = Sierra24A{}

// Draw performs error diffusion dithering.
//
// This method satisfies the draw.Drawer interface, implementing a dithering
// filter attributed to Frankie Sierra.  It uses the kernel
//
//	  X 2
//	1 1
func (d Sierra24A) Draw(dst draw.Image, r image.Rectangle, src image.Image, sp image.Point) {
	pd, ok := dst.(*image.Paletted)
	if !ok {
		// dither211 currently requires a palette
		draw.Draw(dst, r, src, sp, draw.Src)
		return
	}
	// intersect r with both dst and src bounds, fix up sp.
	ir := r.Intersect(pd.Bounds()).
		Intersect(src.Bounds().Add(r.Min.Sub(sp)))
	if ir.Empty() {
		return // no work to do.
	}
	sp = ir.Min.Sub(r.Min)
	// get subimage of src
	sr := ir.Add(sp)
	if !sr.Eq(src.Bounds()) {
		s, ok := src.(interface {
			SubImage(image.Rectangle) image.Image
		})
		if !ok {
			// dither211 currently works on whole images
			draw.Draw(dst, r, src, sp, draw.Src)
			return
		}
		src = s.SubImage(sr)
	}
	// dither211 currently returns a new image, or nil if dithering not
	// possible.
	if s := dither211(src, pd.Palette); s != nil {
		src = s
	}
	// this avoids any problem of src dst overlap but it would usually
	// work to render directly into dst.  todo.
	draw.Draw(dst, r, src, image.Point{}, draw.Src)
}

// signed color type, no alpha.  signed to represent color deltas as well as
// color values 0-ffff as with colorRGBA64
type sRGB struct{ r, g, b int32 }
type sPalette []sRGB

func (p sPalette) index(c sRGB) int {
	// still the awful linear search
	i, min := 0, int64(math.MaxInt64)
	for j, pc := range p {
		d := int64(c.r) - int64(pc.r)
		s := d * d
		d = int64(c.g) - int64(pc.g)
		s += d * d
		d = int64(c.b) - int64(pc.b)
		s += d * d
		if s < min {
			min = s
			i = j
		}
	}
	return i
}

// currently this is strictly a helper function for Dither211.Draw, so
// not generalized to use Palette from this package.
func dither211(i0 image.Image, cp color.Palette) *image.Paletted {
	if len(cp) > 256 {
		// representation limit of image.Paletted.  a little sketchy to return
		// nil, but unworkable results are always better than wrong results.
		return nil
	}
	b := i0.Bounds()
	pi := image.NewPaletted(b, cp)
	if b.Empty() {
		return pi // no work to do
	}
	sp := make(sPalette, len(cp))
	for i, c := range cp {
		r, g, b, _ := c.RGBA()
		sp[i] = sRGB{int32(r), int32(g), int32(b)}
	}
	// afc is adjustd full color.  e, rt, dn hold diffused errors.
	var afc, e, rt sRGB
	dn := make([]sRGB, b.Dx()+1)
	for y := b.Min.Y; y < b.Max.Y; y++ {
		rt = dn[0]
		dn[0] = sRGB{}
		for x := b.Min.X; x < b.Max.X; x++ {
			// full color from original image
			r0, g0, b0, _ := i0.At(x, y).RGBA()
			// adjusted full color = original color + diffused error
			afc.r = int32(r0) + rt.r>>2
			afc.g = int32(g0) + rt.g>>2
			afc.b = int32(b0) + rt.b>>2
			// clipping or clamping is usually explained as necessary
			// to avoid integer overflow but with palettes that do not
			// represent the full color space of the image, it is needed
			// to keep areas of excess color from saturating at palette
			// limits and bleeding into neighboring areas.
			if afc.r < 0 {
				afc.r = 0
			} else if afc.r > 0xffff {
				afc.r = 0xffff
			}
			if afc.g < 0 {
				afc.g = 0
			} else if afc.g > 0xffff {
				afc.g = 0xffff
			}
			if afc.b < 0 {
				afc.b = 0
			} else if afc.b > 0xffff {
				afc.b = 0xffff
			}
			// nearest palette entry
			i := sp.index(afc)
			// set pixel in destination image
			pi.SetColorIndex(x, y, uint8(i))
			// error to be diffused = full color - palette color.
			pc := sp[i]
			e.r = afc.r - pc.r
			e.g = afc.g - pc.g
			e.b = afc.b - pc.b
			// half of error*4 goes right
			dx := x - b.Min.X + 1
			rt.r = dn[dx].r + e.r*2
			rt.g = dn[dx].g + e.g*2
			rt.b = dn[dx].b + e.b*2
			// the other half goes down
			dn[dx] = e
			dn[dx-1].r += e.r
			dn[dx-1].g += e.g
			dn[dx-1].b += e.b
		}
	}
	return pi
}
