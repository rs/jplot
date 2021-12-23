// Copyright 2013 Sonia Keys.
// Licensed under MIT license.  See "license" file in this source tree.

package internal

import "image"

// PxRGBAfunc returns function to get RGBA color values at (x, y) coordinates of
// image img. Returned function works the same as img.At(x, y).RGBA() but
// implements special cases for certain image types to use type-specific methods
// bypassing color.Color interface which escapes to the heap.
func PxRGBAfunc(img image.Image) func(x, y int) (r, g, b, a uint32) {
	switch img0 := img.(type) {
	case *image.RGBA:
		return func(x, y int) (r, g, b, a uint32) { return img0.RGBAAt(x, y).RGBA() }
	case *image.NRGBA:
		return func(x, y int) (r, g, b, a uint32) { return img0.NRGBAAt(x, y).RGBA() }
	case *image.YCbCr:
		return func(x, y int) (r, g, b, a uint32) { return img0.YCbCrAt(x, y).RGBA() }
	}
	return func(x, y int) (r, g, b, a uint32) { return img.At(x, y).RGBA() }
}
