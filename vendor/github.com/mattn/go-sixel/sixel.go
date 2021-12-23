package sixel

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"io"
	"os"
	"strings"

	"github.com/soniakeys/quant/median"
)

// Encoder encode image to sixel format
type Encoder struct {
	w      io.Writer
	Dither bool
	Width  int
	Height int
}

// NewEncoder return new instance of Encoder
func NewEncoder(w io.Writer) *Encoder {
	return &Encoder{w: w}
}

const (
	specialChNr = byte(0x6d)
	specialChCr = byte(0x64)
)

// Encode do encoding
func (e *Encoder) Encode(img image.Image) error {
	nc := 255 // (>= 2, 8bit, index 0 is reserved for transparent key color)
	width, height := img.Bounds().Dx(), img.Bounds().Dy()
	if width == 0 || height == 0 {
		return nil
	}
	if e.Width > 0 {
		width = e.Width
	}
	if e.Height > 0 {
		height = e.Height
	}

	// make adaptive palette using median cut alogrithm
	q := median.Quantizer(nc - 1)
	paletted := q.Paletted(img)
	if e.Dither {
		// copy source image to new image with applying floyd-stenberg dithering
		draw.FloydSteinberg.Draw(paletted, img.Bounds(), img, image.ZP)
	} else {
		draw.Draw(paletted, img.Bounds(), img, image.ZP, draw.Over)
	}
	// use on-memory output buffer for improving the performance
	var w io.Writer
	if _, ok := e.w.(*os.File); ok {
		w = bytes.NewBuffer(make([]byte, 0, 1024*32))
	} else {
		w = e.w
	}
	// DECSIXEL Introducer(\033P0;0;8q) + DECGRA ("1;1): Set Raster Attributes
	w.Write([]byte{0x1b, 0x50, 0x30, 0x3b, 0x30, 0x3b, 0x38, 0x71, 0x22, 0x31, 0x3b, 0x31})

	for n, v := range paletted.Palette {
		r, g, b, _ := v.RGBA()
		r = r * 100 / 0xFFFF
		g = g * 100 / 0xFFFF
		b = b * 100 / 0xFFFF
		// DECGCI (#): Graphics Color Introducer
		fmt.Fprintf(w, "#%d;2;%d;%d;%d", n+1, r, g, b)
	}

	buf := make([]byte, width*nc)
	cset := make([]bool, nc)
	ch0 := specialChNr
	for z := 0; z < (height+5)/6; z++ {
		// DECGNL (-): Graphics Next Line
		if z > 0 {
			w.Write([]byte{0x2d})
		}
		for p := 0; p < 6; p++ {
			y := z*6 + p
			for x := 0; x < width; x++ {
				_, _, _, alpha := img.At(x, y).RGBA()
				if alpha != 0 {
					idx := paletted.ColorIndexAt(x, y) + 1
					cset[idx] = false // mark as used
					buf[width*int(idx)+x] |= 1 << uint(p)
				}
			}
		}
		for n := 1; n < nc; n++ {
			if cset[n] {
				continue
			}
			cset[n] = true
			// DECGCR ($): Graphics Carriage Return
			if ch0 == specialChCr {
				w.Write([]byte{0x24})
			}
			// select color (#%d)
			if n >= 100 {
				digit1 := n / 100
				digit2 := (n - digit1*100) / 10
				digit3 := n % 10
				c1 := byte(0x30 + digit1)
				c2 := byte(0x30 + digit2)
				c3 := byte(0x30 + digit3)
				w.Write([]byte{0x23, c1, c2, c3})
			} else if n >= 10 {
				c1 := byte(0x30 + n/10)
				c2 := byte(0x30 + n%10)
				w.Write([]byte{0x23, c1, c2})
			} else {
				w.Write([]byte{0x23, byte(0x30 + n)})
			}
			cnt := 0
			for x := 0; x < width; x++ {
				// make sixel character from 6 pixels
				ch := buf[width*n+x]
				buf[width*n+x] = 0
				if ch0 < 0x40 && ch != ch0 {
					// output sixel character
					s := 63 + ch0
					for ; cnt > 255; cnt -= 255 {
						w.Write([]byte{0x21, 0x32, 0x35, 0x35, s})
					}
					if cnt == 1 {
						w.Write([]byte{s})
					} else if cnt == 2 {
						w.Write([]byte{s, s})
					} else if cnt == 3 {
						w.Write([]byte{s, s, s})
					} else if cnt >= 100 {
						digit1 := cnt / 100
						digit2 := (cnt - digit1*100) / 10
						digit3 := cnt % 10
						c1 := byte(0x30 + digit1)
						c2 := byte(0x30 + digit2)
						c3 := byte(0x30 + digit3)
						// DECGRI (!): - Graphics Repeat Introducer
						w.Write([]byte{0x21, c1, c2, c3, s})
					} else if cnt >= 10 {
						c1 := byte(0x30 + cnt/10)
						c2 := byte(0x30 + cnt%10)
						// DECGRI (!): - Graphics Repeat Introducer
						w.Write([]byte{0x21, c1, c2, s})
					} else if cnt > 0 {
						// DECGRI (!): - Graphics Repeat Introducer
						w.Write([]byte{0x21, byte(0x30 + cnt), s})
					}
					cnt = 0
				}
				ch0 = ch
				cnt++
			}
			if ch0 != 0 {
				// output sixel character
				s := 63 + ch0
				for ; cnt > 255; cnt -= 255 {
					w.Write([]byte{0x21, 0x32, 0x35, 0x35, s})
				}
				if cnt == 1 {
					w.Write([]byte{s})
				} else if cnt == 2 {
					w.Write([]byte{s, s})
				} else if cnt == 3 {
					w.Write([]byte{s, s, s})
				} else if cnt >= 100 {
					digit1 := cnt / 100
					digit2 := (cnt - digit1*100) / 10
					digit3 := cnt % 10
					c1 := byte(0x30 + digit1)
					c2 := byte(0x30 + digit2)
					c3 := byte(0x30 + digit3)
					// DECGRI (!): - Graphics Repeat Introducer
					w.Write([]byte{0x21, c1, c2, c3, s})
				} else if cnt >= 10 {
					c1 := byte(0x30 + cnt/10)
					c2 := byte(0x30 + cnt%10)
					// DECGRI (!): - Graphics Repeat Introducer
					w.Write([]byte{0x21, c1, c2, s})
				} else if cnt > 0 {
					// DECGRI (!): - Graphics Repeat Introducer
					w.Write([]byte{0x21, byte(0x30 + cnt), s})
				}
			}
			ch0 = specialChCr
		}
	}
	// string terminator(ST)
	w.Write([]byte{0x1b, 0x5c})

	// copy to given buffer
	if _, ok := e.w.(*os.File); ok {
		w.(*bytes.Buffer).WriteTo(e.w)
	}

	return nil
}

// Decoder decode sixel format into image
type Decoder struct {
	r io.Reader
}

// NewDecoder return new instance of Decoder
func NewDecoder(r io.Reader) *Decoder {
	return &Decoder{r}
}

// Decode do decoding from image
func (e *Decoder) Decode(img *image.Image) error {
	buf := bufio.NewReader(e.r)
	c, err := buf.ReadByte()
	if err != nil {
		if err == io.EOF {
			err = nil
		}
		return err
	}
	if c != '\x1B' {
		return errors.New("Invalid format")
	}
	c, err = buf.ReadByte()
	if err != nil {
		return err
	}
	switch c {
	case 'P':
		s, err := buf.ReadString('q')
		if err != nil {
			return err
		}
		s = s[:len(s)-1]
		tok := strings.Split(s, ";")
		if len(tok) != 3 {
			return errors.New("invalid format: illegal header tokens")
		}
	default:
		return errors.New("Invalid format: illegal header")
	}
	c, err = buf.ReadByte()
	if err != nil {
		return err
	}
	if c == '"' {
		s, err := buf.ReadString('#')
		if err != nil {
			return err
		}
		tok := strings.Split(s, ";")
		if len(tok) != 2 {
			return errors.New("invalid format: illegal size tokens")
		}
		err = buf.UnreadByte()
		if err != nil {
			return err
		}
	} else {
		err = buf.UnreadByte()
		if err != nil {
			return err
		}
	}

	colors := map[uint]color.Color{}
	dx, dy := 0, 0
	dw, dh, w, h := 0, 0, 200, 200
	pimg := image.NewNRGBA(image.Rect(0, 0, w, h))
	var tmp *image.NRGBA
data:
	for {
		c, err = buf.ReadByte()
		if err != nil {
			if err == io.EOF {
				err = nil
			}
			return err
		}
		if c == '\r' || c == '\n' || c == '\b' {
			continue
		}
		switch {
		case c == '\x1b':
			c, err = buf.ReadByte()
			if err != nil {
				return err
			}
			if c == '\\' {
				break data
			}
		case c == '$':
			dx = 0
		case c == '?':
			pimg.SetNRGBA(dx, dy, color.NRGBA{0, 0, 0, 0})
			dx++
			if dx >= dw {
				dw = dx
			}
		case c == '!':
			err = buf.UnreadByte()
			if err != nil {
				return err
			}
			var nc, c uint
			n, err := fmt.Fscanf(buf, "!%d%c", &nc, &c)
			if err != nil {
				return err
			}
			if n != 2 {
				return errors.New("invalid format: illegal data tokens")
			}
			if c == '?' {
			}
		case c == '-':
			dy++
			if dy >= dh {
				dh = dy
			}
		case c == '#':
			err = buf.UnreadByte()
			if err != nil {
				return err
			}
			var nc, ci uint
			var r, g, b uint
			var c byte
			n, err := fmt.Fscanf(buf, "#%d%c", &nc, &c)
			if err != nil {
				return err
			}
			if n != 2 {
				return errors.New("invalid format: illegal data tokens")
			}
			if c == ';' {
				n, err := fmt.Fscanf(buf, "%d;%d;%d;%d", &ci, &r, &g, &b)
				if err != nil {
					return err
				}
				if n != 4 {
					return errors.New("invalid format: illegal data tokens")
				}
				colors[uint(nc)] = color.NRGBA{uint8(r * 0xFF / 100), uint8(g * 0xFF / 100), uint8(b * 0xFF / 100), 0XFF}
			} else {
				err = buf.UnreadByte()
				if err != nil {
					return err
				}
				if int(nc) < len(colors) {
					pimg.Set(dx, dy, colors[nc])
				}
				dx++
				if dx >= dw {
					dw = dx
				}
			}
		default:
			if c >= '?' && c <= '~' {
				break
			}
			return errors.New("invalid format: illegal data tokens")
		}
		if dw > w || dh > h {
			if dw > w {
				w *= 2
			}
			if dh > h {
				h *= 2
			}
			tmp = image.NewNRGBA(image.Rect(0, 0, w, h))
			draw.Draw(tmp, pimg.Bounds(), pimg, image.Point{0, 0}, draw.Src)
			pimg = tmp
		}
	}
	rect := image.Rect(0, 0, dw, dh)
	tmp = image.NewNRGBA(rect)
	draw.Draw(tmp, rect, pimg, image.Point{0, 0}, draw.Src)
	*img = tmp
	return nil
}
