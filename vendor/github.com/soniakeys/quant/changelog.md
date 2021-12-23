## v0.2 2013 Nov 14

* Ditherer added.  Sierra 24A, a simpler kernel than Floyd-Steinberg.
Provided through the new draw.Drawer interface which is new to Go 1.2.
* New draw.Quantizer interface supported, which is also new to Go 1.2.
* Tests added.  These work on whatever .png files are found in the source
directory.  The tests do nothing if no suitable .png files are present.

## v0.1 2013 Sep 20

This started as a little toy program to implement color quantization.  But then
I looked to see what else was published in Go along these lines and didn't find
much.  To add something of interest, I implemented a second algorithm and added
an interface.  That's about all that is in v0.1 here.  It's far from general
utility.  It needs things like dithering, subsampling, optimization for common
image types, and test code.
