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

type s3 struct { // want "struct{.+} has size 20, could be 16, rearrange to struct{y uint64; x uint32; z uint32} for optimal size"
	x uint32
	y uint64
	z uint32
}

type s4 struct { // want `struct{.+} has size 32, could be 20, rearrange to struct{_ \[0\]func\(\); i1 int; i2 int; a3 \[3\]bool; b bool} for optimal size`
	b  bool
	i1 int
	i2 int
	a3 [3]bool
	_  [0]func()
}

type s5 struct { // want `struct{.+} has size 24, could be 20, rearrange to struct{y uint64; z \*httptest.Server; x uint32} for optimal size`
	x uint32
	y uint64
	z *httptest.Server
}
