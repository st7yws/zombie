// Copyright (C) 2020  Thenagi<thenagi@ruiko.net>  https://www.thenagi.com/
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

// Copyright 2018 Hajime Hoshi
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package bitmap

import (
	"image"
	"image/color"
	"zombie/bitmapfont/unicode"

	"golang.org/x/image/font"
	"golang.org/x/image/math/fixed"
)

type BinaryImage struct {
	bits   []byte
	width  int
	height int
	bounds image.Rectangle
}

func NewBinaryImage(bits []byte, width, height int) *BinaryImage {
	return &BinaryImage{
		bits:   bits,
		width:  width,
		height: height,
		bounds: image.Rect(0, 0, width, height),
	}
}

func (b *BinaryImage) At(i, j int) color.Color {
	if i < b.bounds.Min.X || j < b.bounds.Min.Y || i >= b.bounds.Max.X || j >= b.bounds.Max.Y {
		return color.Alpha{}
	}
	idx := b.width*j + i
	if (b.bits[idx/8]>>uint(7-idx%8))&1 != 0 {
		return color.Alpha{A: 0xff}
	}
	return color.Alpha{}
}

func (b *BinaryImage) ColorModel() color.Model {
	return color.AlphaModel
}

func (b *BinaryImage) Bounds() image.Rectangle {
	return b.bounds
}

func (b *BinaryImage) SubImage(r image.Rectangle) image.Image {
	bounds := r.Intersect(b.bounds)
	if bounds.Empty() {
		return &BinaryImage{}
	}
	return &BinaryImage{
		bits:   b.bits,
		width:  b.width,
		height: b.height,
		bounds: bounds,
	}
}

const (
	charXNum = 256
	charYNum = 256
)

type Face struct {
	image *BinaryImage
	dotX  fixed.Int26_6
	dotY  fixed.Int26_6
}

func NewFace(image *BinaryImage, dotX, dotY fixed.Int26_6) *Face {
	return &Face{
		image: image,
		dotX:  dotX,
		dotY:  dotY,
	}
}

func (f *Face) runeWidth(r rune) int {
	if unicode.IsEuropian(r) {
		return f.charHalfWidth()
	}
	if 0xff61 <= r && r <= 0xffdc {
		return f.charHalfWidth()
	}
	if 0xffe8 <= r && r <= 0xffee {
		return f.charHalfWidth()
	}
	return f.charFullWidth()
}

func (f *Face) charFullWidth() int {
	return f.image.Bounds().Dx() / charXNum
}

func (f *Face) charHalfWidth() int {
	return f.charFullWidth() / 2
}

func (f *Face) charHeight() int {
	return f.image.Bounds().Dy() / charYNum
}

func (f *Face) Close() error {
	return nil
}

func (f *Face) Glyph(dot fixed.Point26_6, r rune) (dr image.Rectangle, mask image.Image, maskp image.Point, advance fixed.Int26_6, ok bool) {
	if r >= 0x10000 {
		return
	}

	rw := f.runeWidth(r)
	dx := (dot.X - f.dotX).Floor()
	dy := (dot.Y - f.dotY).Floor()
	dr = image.Rect(dx, dy, dx+rw, dy+f.charHeight())

	mx := (int(r) % charXNum) * f.charFullWidth()
	my := (int(r) / charXNum) * f.charHeight()
	mask = f.image.SubImage(image.Rect(mx, my, mx+rw, my+f.charHeight()))
	maskp = image.Pt(mx, my)
	advance = fixed.I(f.runeWidth(r))
	ok = true
	return
}

func (f *Face) GlyphBounds(r rune) (bounds fixed.Rectangle26_6, advance fixed.Int26_6, ok bool) {
	if r >= 0x10000 {
		return
	}
	bounds = fixed.Rectangle26_6{
		Min: fixed.Point26_6{X: -f.dotX, Y: -f.dotY},
		Max: fixed.Point26_6{X: -f.dotX + fixed.I(f.runeWidth(r)), Y: -f.dotY + fixed.I(f.charHeight())},
	}
	advance = fixed.I(f.runeWidth(r))
	ok = true
	return
}

func (f *Face) GlyphAdvance(r rune) (advance fixed.Int26_6, ok bool) {
	if r >= 0x10000 {
		return
	}
	advance = fixed.I(f.runeWidth(r))
	ok = true
	return
}

func (f *Face) Kern(_, _ rune) fixed.Int26_6 {
	return 0
}

func (f *Face) Metrics() font.Metrics {
	return font.Metrics{
		Height:  fixed.I(f.charHeight()),
		Ascent:  f.dotY,
		Descent: fixed.I(f.charHeight()) - f.dotY,
	}
}
