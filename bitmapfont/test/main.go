// go run ./main.go
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
	"flag"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"os"
	"strings"

	"github.com/pkg/browser"
	"golang.org/x/image/font"
	"golang.org/x/image/math/fixed"

	"zombie/bitmapfont"
)

func run() error {
	width := 1050

	text := `最像素BDF字库 12x12像素字体

英：All human beings are born free and equal in dignity and rights.They are endowed with reason and conscience and should act towards one another in a spirit of brotherhood.
简：人人生而自由,在尊严和权利上一律平等.他们赋有理性和良心,并应以兄弟关系的精神相对待.
繁：人生而自由;在尊嚴及權利上均各平等.人各賦有理性良知,誠應和睦相處,情同手足.
日：すべての人間は,生れながらにして自由であり,かつ,尊厳と権利とについて平等である.人間は,理性と良心とを授けられており,互いに同胞の精神をもって行動しなければならない.
`

	const (
		offsetX = 8
		offsetY = 8
	)

	var (
		dotY        int
		glyphHeight int
	)

	dotY = 12
	glyphHeight = 12

	height := glyphHeight*len(strings.Split(strings.TrimSpace(text), "\n")) + offsetX*2

	dst := image.NewRGBA(image.Rect(0, 0, width, height))
	draw.Draw(dst, dst.Bounds(), image.NewUniform(color.White), image.Point{}, draw.Src)

	var f font.Face

	f = bitmapfont.Zpix

	d := font.Drawer{
		Dst:  dst,
		Src:  image.NewUniform(color.Black),
		Face: f,
		Dot:  fixed.P(offsetX, dotY+offsetY),
	}

	for _, l := range strings.Split(text, "\n") {
		d.DrawString(l)
		d.Dot.X = fixed.I(offsetX)
		d.Dot.Y += f.Metrics().Height
	}

	path := "example.png"
	fout, err := os.Create(path)
	if err != nil {
		return err
	}
	//noinspection GoUnhandledErrorResult
	defer fout.Close()

	if err := png.Encode(fout, d.Dst); err != nil {
		return err
	}

	if err := browser.OpenFile(path); err != nil {
		return err
	}

	return nil
}

func main() {
	flag.Parse()
	if err := run(); err != nil {
		panic(err)
	}
}
