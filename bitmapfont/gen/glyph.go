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
	"os"
	"path/filepath"
	"runtime"
)

func readBDF() (map[rune]*Glyph, error) {
	_, current, _, _ := runtime.Caller(1)

	if zpix, err := os.Open(filepath.Join(filepath.Dir(current), "zpix.bdf")); err == nil {
		//noinspection GoUnhandledErrorResult
		defer zpix.Close()

		m := map[rune]*Glyph{}

		glyphs, err := parse(zpix)
		if err != nil {
			return nil, err
		}
		for _, g := range glyphs {
			m[rune(g.Encoding)] = g
		}

		return m, nil
	} else {
		return nil, err
	}
}

var glyphs map[rune]*Glyph

func init() {
	zpixGlyphs, err := readBDF()
	if err != nil {
		panic(err)
	}

	glyphs = zpixGlyphs
}

func getGlyphFromBDF(r rune) (Glyph, bool) {
	g, ok := glyphs[r]
	if !ok {
		return Glyph{}, false
	}
	return *g, true
}
