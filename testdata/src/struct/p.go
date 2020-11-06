// Copyright 2020 Orijtech, Inc. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package p

import "net/http/httptest"

type s struct{}

type s1 struct {
	i int
}

type s2 struct {
	i int
	j int
}

type s3 struct { // want `struct has size 24 \(size class 32\), could be 16 \(size class 16\), rearrange to struct{y uint64; x uint32; z uint32} for optimal size`
	x uint32
	y uint64
	z uint32
}

type s4 struct { // want `struct has size 40 \(size class 48\), could be 24 \(size class 32\), rearrange to struct{_ \[0\]func\(\); i1 int; i2 int; a3 \[3\]bool; b bool} for optimal size`
	b  bool
	i1 int
	i2 int
	a3 [3]bool
	_  [0]func()
}

// should be good, the struct has size 32, can be rearrange to have size 24, but runtime allocator
// allocate the same size class 32.
type s5 struct {
	x uint32
	y uint64
	z *s
	t uint32
}

type s6 struct { // should be good, see #16
	bytep *uint8
	mask  uint8
	index uintptr
}

type s7 struct { // want `struct has size 40 \(size class 48\), could be 32 \(size class 32\), rearrange to struct{y uint64; t \*httptest.Server; w uint64; x uint32; z uint32} for optimal size`
	x uint32
	y uint64
	t *httptest.Server
	z uint32
	w uint64
}

type s8 struct { // want `struct has size 40 \(size class 48\), could be 32 \(size class 32\), rearrange to struct{y uint64; t \*s; w uint64; x uint32; z uint32} for optimal size`
	x uint32
	y uint64
	t *s
	z uint32
	w uint64
}
