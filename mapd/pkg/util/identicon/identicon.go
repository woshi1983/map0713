// Copyright (c) 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package identicon

import (
	"bytes"
	"hash"
	"image"
	"image/color"
	"image/png"
	"io"

	siphash "github.com/dchest/siphash"
	httputil "gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/util/httputil"
	randutil "gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/util/randutil"
)

// StandardPNGRenderer is the stardard PNG identicon renderer.
var StandardPNGRenderer = New7x7Renderer(&png.Encoder{}, randutil.BytesN(16))

// ImageEncoder is the encoder to be used for rendering an image.
type ImageEncoder interface {
	Encode(w io.Writer, m image.Image) error
}

const (
	xborder = 35
	yborder = 35
	maxX    = 420
	maxY    = 420
)

// Renderer allows rendering of data as an identicon image.
type Renderer struct {
	enc    ImageEncoder
	key    []byte
	sqSize int
	rows   int
	cols   int
	h      hash.Hash64
}

// Render renders an identicon from p.
func (icon *Renderer) Render(p []byte) []byte {
	icon.h.Reset()
	icon.h.Write(p)
	h := icon.h.Sum64()

	nrgba := color.NRGBA{
		R: uint8(h),
		G: uint8(h >> 8),
		B: uint8(h >> 16),
		A: 0xff,
	}
	h >>= 24

	img := image.NewPaletted(image.Rect(0, 0, maxX, maxY), color.Palette{color.NRGBA{0xf0, 0xf0, 0xf0, 0xff}, nrgba})
	sqx := 0
	sqy := 0

	pixels := make([]byte, icon.sqSize)
	for i := 0; i < icon.sqSize; i++ {
		pixels[i] = 1
	}

	for i := 0; i < icon.rows*(icon.cols+1)/2; i++ {
		if h&1 == 1 {
			for i := 0; i < icon.sqSize; i++ {
				x := xborder + sqx*icon.sqSize
				y := yborder + sqy*icon.sqSize + i
				offs := img.PixOffset(x, y)
				copy(img.Pix[offs:], pixels)

				x = xborder + (icon.cols-1-sqx)*icon.sqSize
				offs = img.PixOffset(x, y)
				copy(img.Pix[offs:], pixels)
			}
		}

		h >>= 1
		sqy++
		if sqy == icon.rows {
			sqy = 0
			sqx++
		}
	}
	var buf bytes.Buffer
	icon.enc.Encode(&buf, img)
	return buf.Bytes()
}

// New5x5Renderer creates a new 5-by-5 identicon renderer using 'key' as the hash salt
func New5x5Renderer(encoder ImageEncoder, key []byte) *Renderer {
	return &Renderer{
		enc:    encoder,
		sqSize: 70,
		rows:   5,
		cols:   5,
		h:      siphash.New(key),
	}
}

// New7x7Renderer creates a new 7-by-7 identicon renderer using 'key' as the hash salt
func New7x7Renderer(encoder ImageEncoder, key []byte) *Renderer {
	return &Renderer{
		enc:    encoder,
		sqSize: 50,
		rows:   7,
		cols:   7,
		h:      siphash.New(key),
	}
}

// NewPNG generates a new 7-by-7 identicon with StandardPNGRenderer and the given data.
func NewPNG(data []byte) []byte {
	return StandardPNGRenderer.Render(data)
}

// NewDataURL generates a new 7-by-7 identicon in a DataURL formation.
func NewDataURL(data []byte) *httputil.DataURL {
	return &httputil.DataURL{MediaType: "image/png", Base64: true, RawData: NewPNG(data)}
}
