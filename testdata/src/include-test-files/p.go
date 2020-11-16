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

type s struct { // want `struct has size 24 \(size class 32\), could be 16 \(size class 16\), you'll save 50.00% if you rearrange it to:\nstruct {\n\ty uint64\n\tx uint32\n\tz uint32\n}`
	x uint32
	y uint64
	z uint32
}
