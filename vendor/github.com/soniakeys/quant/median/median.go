// Copyright 2013 Sonia Keys.
// Licensed under MIT license.  See "license" file in this source tree.

// Median implements basic median cut color quantization.
package median

import (
	"container/heap"
	"image"
	"image/color"
	"image/draw"
	"math"
	"sort"

	"github.com/soniakeys/quant"
	"github.com/soniakeys/quant/internal"
)

// Quantizer methods implement median cut color quantization.
//
// The value is the target number of colors.
// Methods do not require pointer receivers, simply construct Quantizer
// objects with a type conversion.
//
// The type satisfies both quant.Quantizer and draw.Quantizer interfaces.
type Quantizer int

var _ quant.Quantizer = Quantizer(0)
var _ draw.Quantizer = Quantizer(0)

// Paletted performs color quantization and returns a paletted image.
//
// Returned is an image.Paletted with no more than q colors. Note though
// that image.Paletted is limited to 256 colors.
func (q Quantizer) Paletted(img image.Image) *image.Paletted {
	n := int(q)
	if n > 256 {
		n = 256
	}
	qz := newQuantizer(img, n)
	if n > 1 {
		qz.cluster() // cluster pixels by color
	}
	return qz.paletted() // generate paletted image from clusters
}

// Palette performs color quantization and returns a quant.Palette object.
//
// Returned is a palette with no more than q colors. Q may be > 256.
func (q Quantizer) Palette(img image.Image) quant.Palette {
	qz := newQuantizer(img, int(q))
	if q > 1 {
		qz.cluster() // cluster pixels by color
	}
	return qz.t
}

// Quantize performs color quantization and returns a color.Palette.
//
// Following the behavior documented with the draw.Quantizer interface,
// "Quantize appends up to cap(p) - len(p) colors to p and returns the
// updated palette...."  This method does not limit the number of colors
// to 256.  Cap(p) or the quantity cap(p) - len(p) may be > 256.
// Also for this method the value of the Quantizer object is ignored.
func (Quantizer) Quantize(p color.Palette, m image.Image) color.Palette {
	n := cap(p) - len(p)
	qz := newQuantizer(m, n)
	if n > 1 {
		qz.cluster() // cluster pixels by color
	}
	return p[:len(p)+copy(p[len(p):cap(p)], qz.t.ColorPalette())]
}

type quantizer struct {
	img image.Image       // original image
	cs  []cluster         // len(cs) is the desired number of colors
	ch  chValues          // buffer for computing median
	t   quant.TreePalette // root

	pxRGBA func(x, y int) (r, g, b, a uint32) // function to get original image RGBA color values
}

type point struct{ x, y int32 }
type chValues []uint16
type queue []*cluster

type cluster struct {
	px       []point // list of points in the cluster
	widestCh int     // rgb const identifying axis with widest value range
	// limits of this cluster
	minR, maxR uint32
	minG, maxG uint32
	minB, maxB uint32
	// true if corresponding value above represents a bound or hull of the
	// represented color space
	bMinR, bMaxR bool
	bMinG, bMaxG bool
	bMinB, bMaxB bool
	node         *quant.Node // palette node representing this cluster
}

// indentifiers for RGB channels, or dimensions or axes of RGB color space
const (
	rgbR = iota
	rgbG
	rgbB
)

func newQuantizer(img image.Image, nq int) *quantizer {
	if nq < 1 {
		return &quantizer{img: img, pxRGBA: internal.PxRGBAfunc(img)}
	}
	b := img.Bounds()
	npx := (b.Max.X - b.Min.X) * (b.Max.Y - b.Min.Y)
	qz := &quantizer{
		img:    img,
		ch:     make(chValues, npx),
		cs:     make([]cluster, nq),
		pxRGBA: internal.PxRGBAfunc(img),
	}
	// Populate initial cluster with all pixels from image.
	c := &qz.cs[0]
	px := make([]point, npx)
	c.px = px
	c.node = &quant.Node{}
	qz.t.Root = c.node
	c.minR = math.MaxUint32
	c.minG = math.MaxUint32
	c.minB = math.MaxUint32
	c.bMinR = true
	c.bMinG = true
	c.bMinB = true
	c.bMaxR = true
	c.bMaxG = true
	c.bMaxB = true
	i := 0
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			px[i].x = int32(x)
			px[i].y = int32(y)
			r, g, b, _ := qz.pxRGBA(x, y)
			if r < c.minR {
				c.minR = r
			}
			if r > c.maxR {
				c.maxR = r
			}
			if g < c.minG {
				c.minG = g
			}
			if g > c.maxG {
				c.maxG = g
			}
			if b < c.minB {
				c.minB = b
			}
			if b > c.maxB {
				c.maxB = b
			}
			i++
		}
	}
	return qz
}

// Cluster by repeatedly splitting clusters.
// Use a heap as priority queue for picking clusters to split.
// The rule is to spilt the cluster with the most pixels.
// Terminate when the desired number of clusters has been populated
// or when clusters cannot be further split.
func (qz *quantizer) cluster() {
	pq := new(queue)
	// Initial cluster.  populated at this point, but not analyzed.
	c := &qz.cs[0]
	var m uint32
	i := 1
	for {
		// Only enqueue clusters that can be split.
		if qz.setWidestChannel(c) {
			heap.Push(pq, c)
		}
		// If no clusters have any color variation, mark the end of the
		// cluster list and quit early.
		if len(*pq) == 0 {
			qz.cs = qz.cs[:i]
			break
		}
		s := heap.Pop(pq).(*cluster) // get cluster to split
		m = qz.medianCut(s)
		c = &qz.cs[i] // set c to new cluster
		i++
		qz.split(s, c, m) // split s into c and s at value m
		// Normal exit is when all clusters are populated.
		if i == len(qz.cs) {
			break
		}
		if qz.setWidestChannel(s) {
			heap.Push(pq, s) // return s to queue
		}
	}
	// set TreePalette total and indexes
	qz.t.Leaves = i
	qz.t.Walk(func(leaf *quant.Node, i int) { leaf.Index = i })
	// compute palette colors
	for i := range qz.cs {
		px := qz.cs[i].px
		// Average values in cluster to get palette color.
		var rsum, gsum, bsum int64
		for _, p := range px {
			r, g, b, _ := qz.pxRGBA(int(p.x), int(p.y))
			rsum += int64(r)
			gsum += int64(g)
			bsum += int64(b)
		}
		n64 := int64(len(px))
		qz.cs[i].node.Color = color.RGBA64{
			uint16(rsum / n64),
			uint16(gsum / n64),
			uint16(bsum / n64),
			0xffff,
		}
	}
}

func (q *quantizer) setWidestChannel(c *cluster) bool {
	// Find extents of color values in each dimension.
	// (limits in cluster are not good enough here, we want extents as
	// represented by pixels.)
	var maxR, maxG, maxB uint32
	minR := uint32(math.MaxUint32)
	minG := uint32(math.MaxUint32)
	minB := uint32(math.MaxUint32)
	for _, p := range c.px {
		r, g, b, _ := q.pxRGBA(int(p.x), int(p.y))
		if r < minR {
			minR = r
		}
		if r > maxR {
			maxR = r
		}
		if g < minG {
			minG = g
		}
		if g > maxG {
			maxG = g
		}
		if b < minB {
			minB = b
		}
		if b > maxB {
			maxB = b
		}
	}
	// See which color dimension had the widest range.
	c.widestCh = rgbG
	min := minG
	max := maxG
	if maxR-minR > max-min {
		c.widestCh = rgbR
		min = minR
		max = maxR
	}
	if maxB-minB > max-min {
		c.widestCh = rgbB
		min = minB
		max = maxB
	}
	return max > min
}

// Arg c must have value range > 0 in dimension c.widestDim.
// return value m is guararanteed to split cluster into two non-empty clusters
// by v < m where v is pixel value of dimension c.Widest.
func (q *quantizer) medianCut(c *cluster) uint32 {
	px := c.px
	ch := q.ch[:len(px)]
	// Copy values from appropriate color channel to buffer for
	// computing median.
	switch c.widestCh {
	case rgbR:
		for i, p := range c.px {
			r, _, _, _ := q.pxRGBA(int(p.x), int(p.y))
			ch[i] = uint16(r)
		}
	case rgbG:
		for i, p := range c.px {
			_, g, _, _ := q.pxRGBA(int(p.x), int(p.y))
			ch[i] = uint16(g)
		}
	case rgbB:
		for i, p := range c.px {
			_, _, b, _ := q.pxRGBA(int(p.x), int(p.y))
			ch[i] = uint16(b)
		}
	}
	// Find cut.
	sort.Sort(ch)
	m1 := len(ch) / 2 // median
	if ch[m1] != ch[m1-1] {
		return uint32(ch[m1])
	}
	m2 := m1
	// Dec m1 until element to left is different.
	for m1--; m1 > 0 && ch[m1] == ch[m1-1]; m1-- {
	}
	// Inc m2 until element to left is different.
	for m2++; m2 < len(ch) && ch[m2] == ch[m2-1]; m2++ {
	}
	// Return value that makes more equitable cut.
	if m1 > len(ch)-m2 {
		return uint32(ch[m1])
	}
	return uint32(ch[m2])
}

// split s into c and s at value m
func (q *quantizer) split(s, c *cluster, m uint32) {
	*c = *s // copy extent data
	px := s.px
	var v uint32
	i := 0
	last := len(px) - 1
	for i <= last {
		// Get color value in appropriate dimension.
		r, g, b, _ := q.pxRGBA(int(px[i].x), int(px[i].y))
		switch s.widestCh {
		case rgbR:
			v = r
		case rgbG:
			v = g
		case rgbB:
			v = b
		}
		// Split at m.
		if v < m {
			i++
		} else {
			px[last], px[i] = px[i], px[last]
			last--
		}
	}
	// Split the pixel list.  s keeps smaller values, c gets larger values.
	s.px = px[:i]
	c.px = px[i:]
	// Split color extent
	n := s.node
	switch s.widestCh {
	case rgbR:
		s.maxR = m
		c.minR = m
		s.bMaxR = false
		c.bMinR = false
		n.Type = quant.TSplitR
	case rgbG:
		s.maxG = m
		c.minG = m
		s.bMaxG = false
		c.bMinG = false
		n.Type = quant.TSplitG
	case rgbB:
		s.maxB = m
		c.minB = m
		s.bMaxB = false
		c.bMinB = false
		n.Type = quant.TSplitB
	}
	// Split node
	n.Split = m
	n.Low = &quant.Node{}
	n.High = &quant.Node{}
	s.node, c.node = n.Low, n.High
}

func (qz *quantizer) paletted() *image.Paletted {
	cp := qz.t.ColorPalette()
	pi := image.NewPaletted(qz.img.Bounds(), cp)
	for i := range qz.cs {
		x := uint8(qz.cs[i].node.Index)
		for _, p := range qz.cs[i].px {
			pi.SetColorIndex(int(p.x), int(p.y), x)
		}
	}
	return pi
}

// Implement sort.Interface for sort in median algorithm.
func (c chValues) Len() int           { return len(c) }
func (c chValues) Less(i, j int) bool { return c[i] < c[j] }
func (c chValues) Swap(i, j int)      { c[i], c[j] = c[j], c[i] }

// Implement heap.Interface for priority queue of clusters.
func (q queue) Len() int { return len(q) }

// Priority is number of pixels in cluster.
func (q queue) Less(i, j int) bool { return len(q[i].px) > len(q[j].px) }
func (q queue) Swap(i, j int) {
	q[i], q[j] = q[j], q[i]
}
func (pq *queue) Push(x interface{}) {
	c := x.(*cluster)
	*pq = append(*pq, c)
}
func (pq *queue) Pop() interface{} {
	q := *pq
	n := len(q) - 1
	c := q[n]
	*pq = q[:n]
	return c
}
