// Generate bitmapfontZpix.go and sample.png
// 生成bitmapfontZpix.go并附带一个字体库预览图
// go build && ./gen
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

package main

import (
	"compress/gzip"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"io"
	"io/ioutil"
	"os"
)

// getGlyph 获取字符图形
func getGlyph(r rune) (Glyph, bool) {
	g, ok := getGlyphFromBDF(r)
	if ok {
		return g, true
	}

	return Glyph{}, false
}

// addGlyphs 绘制所有字符图形
func addGlyphs(img draw.Image) {
	gw, gh := 12, 12
	for j := 0; j < 0x100; j++ {
		for i := 0; i < 0x100; i++ {
			r := rune(i + j*0x100)
			g, ok := getGlyph(r)
			if !ok {
				continue
			}

			dstX := i*gw + g.X
			dstY := j*gh + ((gh - g.Height) - 2 - g.Y)
			dstR := image.Rect(dstX, dstY, dstX+g.Width, dstY+g.Height)
			p := g.Bounds().Min
			draw.Draw(img, dstR, &g, p, draw.Over)
		}
	}
}

// write 写入字体byte到bitmapfontZpix.go
func write(w io.Writer, r io.Reader) error {
	bs, err := ioutil.ReadAll(r)
	if err != nil {
		return err
	}

	if _, err := fmt.Fprintln(w, "// DO NOT EDIT."); err != nil {
		return err
	}
	if _, err := fmt.Fprintln(w, ""); err != nil {
		return err
	}
	if _, err := fmt.Fprintln(w, "package bitmapfont"); err != nil {
		return err
	}
	if _, err := fmt.Fprintln(w, ""); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, "var %s = []byte(%q)\n", "bitmapfontZpix", string(bs)); err != nil {
		return err
	}
	return nil
}

func main() {
	img := image.NewAlpha(image.Rect(0, 0, 12*256, 12*256))
	addGlyphs(img)

	b := img.Bounds()
	w, h := b.Dx(), b.Dy()
	as := make([]byte, w*h/8)
	for j := 0; j < h; j++ {
		for i := 0; i < w; i++ {
			a := img.At(i, j).(color.Alpha).A
			idx := w*j + i
			if a != 0 {
				as[idx/8] |= 1 << uint(7-idx%8)
			}
		}
	}

	file, _ := os.Create("../sample.png")
	_ = png.Encode(file, img)
	_ = file.Close()

	fout, err := os.Create("zpix")
	if err != nil {
		panic(err)
	}

	cw, err := gzip.NewWriterLevel(fout, gzip.BestCompression)
	if err != nil {
		panic(err)
	}

	if _, err := cw.Write(as); err != nil {
		panic(err)
	}

	_ = cw.Close()
	_ = fout.Close()

	var out io.Writer
	o, err := os.Create("../bitmapfontZpix.go")
	if err != nil {
		panic(err)
	}
	out = o

	var in io.Reader
	i, err := os.Open("zpix")
	if err != nil {
		panic(err)
	}
	in = i

	if err := write(out, in); err != nil {
		panic(err)
	}

	_ = o.Close()
	_ = i.Close()
	_ = os.Remove("zpix")
}
