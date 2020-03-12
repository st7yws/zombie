// Analyze zpix.bdf
// 解析zpix.bdf字体文件
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
	"bufio"
	"fmt"
	"image"
	"image/color"
	"io"
	"strconv"
	"strings"
)

type Glyph struct {
	Encoding int

	Width  int
	Height int
	X      int
	Y      int

	Bitmap [][]byte
}

func (g *Glyph) ColorModel() color.Model {
	return color.AlphaModel
}

func (g *Glyph) Bounds() image.Rectangle {
	return image.Rect(0, 0, g.Width, g.Height)
}

func (g *Glyph) At(x, y int) color.Color {
	if x < 0 || y < 0 || x >= g.Width || y >= g.Height {
		return color.Alpha{}
	}
	bits := g.Bitmap[y][x/8]
	if (bits>>uint(7-x%8))&1 != 0 {
		return color.Alpha{A: 0xff}
	}
	return color.Alpha{}
}

func parse(f io.Reader) ([]*Glyph, error) {
	var glyphs []*Glyph
	var current *Glyph

	s := bufio.NewScanner(f)
	for s.Scan() {
		line := s.Text()
		if strings.HasPrefix(line, "STARTCHAR ") {
			if current != nil {
				panic("not reach")
			}
			current = &Glyph{}
			continue
		}
		if strings.HasPrefix(line, "ENCODING ") {
			idx, err := strconv.ParseInt(line[len("ENCODING "):], 10, 32)
			if err != nil {
				return nil, err
			}
			current.Encoding = int(idx)
			continue
		}
		if strings.HasPrefix(line, "BBX ") {
			tokens := strings.Split(line, " ")
			w, err := strconv.ParseInt(tokens[1], 10, 32)
			if err != nil {
				return nil, err
			}
			h, err := strconv.ParseInt(tokens[2], 10, 32)
			if err != nil {
				return nil, err
			}
			x, err := strconv.ParseInt(tokens[3], 10, 32)
			if err != nil {
				return nil, err
			}
			y, err := strconv.ParseInt(tokens[4], 10, 32)
			if err != nil {
				return nil, err
			}
			current.Width = int(w)
			current.Height = int(h)
			current.X = int(x)
			current.Y = int(y)
		}
		if strings.HasPrefix(line, "BITMAP") {
			current.Bitmap = [][]byte{}
			continue
		}
		if strings.HasPrefix(line, "ENDCHAR") {
			glyphs = append(glyphs, current)
			current = nil
			continue
		}
		if current == nil {
			continue
		}
		if current.Bitmap == nil {
			continue
		}
		if len(line)%2 != 0 {
			return nil, fmt.Errorf("bdf: len(line) must be even")
		}
		var bits []byte
		for ; len(line) > 0; line = line[2:] {
			b, err := strconv.ParseInt(line[:2], 16, 32)
			if err != nil {
				return nil, err
			}
			bits = append(bits, byte(b))
		}
		current.Bitmap = append(current.Bitmap, bits)
	}
	return glyphs, nil
}
