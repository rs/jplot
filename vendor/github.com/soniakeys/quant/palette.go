// Copyright 2013 Sonia Keys.
// Licensed under MIT license.  See "license" file in this source tree.

package quant

import (
	"image"
	"image/color"
)

// Palette is a palette of color.Colors, much like color.Palette of the
// standard library.
//
// It is defined as an interface here to allow more general implementations,
// presumably ones that maintain some data structure to achieve performance
// advantages over linear search.
type Palette interface {
	Len() int
	IndexNear(color.Color) int
	ColorNear(color.Color) color.Color
	ColorPalette() color.Palette
}

var _ Palette = LinearPalette{}
var _ Palette = TreePalette{}

// LinearPalette implements the Palette interface with color.Palette
// and has no optimizations.
type LinearPalette struct {
	color.Palette
}

// IndexNear returns the palette index of the nearest palette color.
//
// It simply wraps color.Palette.Index.
func (p LinearPalette) IndexNear(c color.Color) int {
	return p.Palette.Index(c)
}

// Color near returns the nearest palette color.
//
// It simply wraps color.Palette.Convert.
func (p LinearPalette) ColorNear(c color.Color) color.Color {
	return p.Palette.Convert(c)
}

// ColorPalette satisfies interface Palette.
//
// It simply returns the internal color.Palette.
func (p LinearPalette) ColorPalette() color.Palette {
	return p.Palette
}

func (p LinearPalette) Len() int { return len(p.Palette) }

// TreePalette implements the Palette interface with a binary tree.
//
// XNear methods run in O(log n) time for palette size.
//
// Fields are exported for access by quantizer packages.  Typical use of
// TreePalette should be through methods.
type TreePalette struct {
	Leaves int
	Root   *Node
}

func (p TreePalette) Len() int { return p.Leaves }

// Node is a TreePalette node.  It is exported for access by quantizer
// packages and otherwise can be ignored for typical use of this package.
type Node struct {
	Type int
	// for TLeaf
	Index int
	Color color.RGBA64
	// for TSplit
	Split     uint32
	Low, High *Node
}

const (
	TLeaf = iota
	TSplitR
	TSplitG
	TSplitB
)

// IndexNear returns the index of the nearest palette color.
func (t TreePalette) IndexNear(c color.Color) (i int) {
	if t.Root == nil {
		return -1
	}
	t.Search(c, func(leaf *Node) { i = leaf.Index })
	return
}

// ColorNear returns the nearest palette color.
func (t TreePalette) ColorNear(c color.Color) (p color.Color) {
	if t.Root == nil {
		return color.RGBA64{0x7fff, 0x7fff, 0x7fff, 0xfff}
	}
	t.Search(c, func(leaf *Node) { p = leaf.Color })
	return
}

// Search searches for the given color and calls f for the node representing
// the nearest color.
func (t TreePalette) Search(c color.Color, f func(leaf *Node)) {
	r, g, b, _ := c.RGBA()
	var lt bool
	var s func(*Node)
	s = func(n *Node) {
		switch n.Type {
		case TLeaf:
			f(n)
			return
		case TSplitR:
			lt = r < n.Split
		case TSplitG:
			lt = g < n.Split
		case TSplitB:
			lt = b < n.Split
		}
		if lt {
			s(n.Low)
		} else {
			s(n.High)
		}
	}
	s(t.Root)
}

// ColorPalette returns a color.Palette corresponding to the TreePalette.
func (t TreePalette) ColorPalette() color.Palette {
	if t.Root == nil {
		return nil
	}
	p := make(color.Palette, 0, t.Leaves)
	t.Walk(func(leaf *Node, i int) {
		p = append(p, leaf.Color)
	})
	return p
}

// Walk walks the TreePalette calling f for each color.
func (t TreePalette) Walk(f func(leaf *Node, i int)) {
	i := 0
	var w func(*Node)
	w = func(n *Node) {
		if n.Type == TLeaf {
			f(n, i)
			i++
			return
		}
		w(n.Low)
		w(n.High)
	}
	w(t.Root)
}

func Paletted(p Palette, img image.Image) *image.Paletted {
	if p.Len() > 256 {
		return nil
	}
	b := img.Bounds()
	pi := image.NewPaletted(b, p.ColorPalette())
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			pi.SetColorIndex(x, y, uint8(p.IndexNear(img.At(x, y))))
		}
	}
	return pi
}
