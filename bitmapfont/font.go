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

package bitmapfont

import (
	"bytes"
	"compress/gzip"
	"io/ioutil"

	"golang.org/x/image/font"
	"golang.org/x/image/math/fixed"

	"zombie/bitmapfont/bitmap"
)

func init() {
	s, err := gzip.NewReader(bytes.NewReader(bitmapfontZpix))
	if err != nil {
		panic(err)
	}

	//noinspection GoUnhandledErrorResult
	defer s.Close()

	bits, err := ioutil.ReadAll(s)
	if err != nil {
		panic(err)
	}

	Zpix = bitmap.NewFace(bitmap.NewBinaryImage(bits, 12*256, 12*256), fixed.I(0), fixed.I(12))
}

var (
	Zpix font.Face
)
