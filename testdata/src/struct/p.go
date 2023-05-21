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

type s3 struct { // want `struct has size 24 \(size class 24\), could be 16 \(size class 16\), you'll save 33.33% \(8 bytes\) if you rearrange it to:\nstruct {\n\ty uint64\n\tx uint32\n\tz uint32\n}`
	x uint32
	y uint64
	z uint32
}

type s4 struct { // want `struct has size 40 \(size class 48\), could be 24 \(size class 24\), you'll save 50.00% \(24 bytes\) if you rearrange it to:\nstruct {\n\t_  \[0\]func\(\)\n\ti1 int\n\ti2 int\n\ta3 \[3\]bool\n\tb  bool\n}`
	b  bool
	i1 int
	i2 int
	a3 [3]bool
	_  [0]func()
}

// should be good, the struct has size 32, can be rearranged to have size 24, but runtime allocator
// allocate the same size class 32.
type s5 struct { // want `struct has size 32 \(size class 32\), could be 24 \(size class 24\), you'll save 25.00% \(8 bytes\) if you rearrange it to:\nstruct {\n\ty uint64\n\tz \*s\n\tx uint32\n\tt uint32\n}`
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

type s7 struct { // want `struct has size 40 \(size class 48\), could be 32 \(size class 32\), you'll save 33.33% \(16 bytes\) if you rearrange it to:\nstruct {\n\ty uint64\n\tt \*httptest.Server\n\tw uint64\n\tx uint32\n\tz uint32\n}`
	x uint32
	y uint64
	t *httptest.Server
	z uint32
	w uint64
}

type s8 struct { // want `struct has size 40 \(size class 48\), could be 32 \(size class 32\), you'll save 33.33% \(16 bytes\) if you rearrange it to:\nstruct {\n\ty uint64\n\tt \*s\n\tw uint64\n\tx uint32\n\tz uint32\n}`
	x uint32
	y uint64
	t *s
	z uint32
	w uint64
}

// Struct which combines multiple fields of the same type, see issue #41.
type s9 struct { // want `struct has size 40 \(size class 48\), could be 24 \(size class 24\), you'll save 50.00% \(24 bytes\) if you rearrange it to:\nstruct {\n\t_  \[0\]func\(\)\n\ti1 int\n\ti2 int\n\ta3 \[3\]bool\n\tb  bool\n}`
	b      bool
	i1, i2 int
	a3     [3]bool
	_      [0]func()
}

// Preserve comments.
type s10 struct { // want `struct has size 40 \(size class 48\), could be 24 \(size class 24\), you'll save 50.00% \(24 bytes\) if you rearrange it to:\nstruct {\n\t_  \[0\]func\(\)\n\ti1 int\n\ti2 int\n\ta3 \[3\]bool\n\tb  bool\n}`
	b      bool    // b is bool
	i1, i2 int     // i1, i2 are int
	a3     [3]bool // a3 is array of bool
	_      [0]func()
}
