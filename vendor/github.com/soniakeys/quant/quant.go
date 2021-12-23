// Copyright 2013 Sonia Keys.
// Licensed under MIT license.  See "license" file in this source tree.

// Quant provides an interface for image color quantizers.
package quant

import "image"

// Quantizer defines a color quantizer for images.
type Quantizer interface {
	// Paletted quantizes an image and returns a paletted image.
	Paletted(image.Image) *image.Paletted
	// Palette quantizes an image and returns a Palette.  Note the return
	// type is the Palette interface of this package and not image.Palette.
	Palette(image.Image) Palette
}
