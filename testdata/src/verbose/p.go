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

type s struct{} // want `struct has size 0 \(size class 0\)`

type s1 struct { // want `struct has size 1 \(size class 8\)`
	b bool
}

type s3 struct { // want `struct has size 32 \(size class 32\), could be 24 \(size class 32\), optimal fields order:\nstruct {\n\ty uint64\n\tz \*s\n\tx uint32\n\tt uint32\n}`
	x uint32
	y uint64
	z *s
	t uint32
}
