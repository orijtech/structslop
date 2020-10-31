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

package structslop

import (
	"go/types"
)

type sizes struct {
	stdSizes types.Sizes
	maxAlign int64
}

func (s *sizes) Offsetsof(fields []*types.Var) []int64 {
	offsets := make([]int64, len(fields))
	var o int64
	for i, f := range fields {
		a := s.Alignof(f.Type())
		o = align(o, a)
		offsets[i] = o
		o += s.Sizeof(f.Type())
	}
	return offsets
}

func (s *sizes) Sizeof(T types.Type) int64 {
	switch t := T.Underlying().(type) {
	case *types.Array:
		return t.Len() * s.Sizeof(t.Elem())
	case *types.Struct:
		nf := t.NumFields()
		if nf == 0 {
			return 0
		}
		o := int64(0)
		max := int64(1)
		for i := 0; i < nf; i++ {
			ft := t.Field(i).Type()
			a, sz := s.Alignof(ft), s.Sizeof(ft)
			if a > max {
				max = a
			}
			if i == nf-1 && sz == 0 && o != 0 {
				sz = 1
			}
			o = align(o, a) + sz
		}
		return align(o, max)
	}
	return s.stdSizes.Sizeof(T)
}

func (s *sizes) Alignof(T types.Type) int64 {
	switch t := T.Underlying().(type) {
	case *types.Array:
		return s.Alignof(t.Elem())
	case *types.Struct:
		max := int64(1)
		for i, nf := 0, t.NumFields(); i < nf; i++ {
			if a := s.Alignof(t.Field(i).Type()); a > max {
				max = a
			}
		}
		return max
	case *types.Slice, *types.Interface, *types.Basic:
		return s.stdSizes.Alignof(T)
	}

	// All other types.
	a := s.Sizeof(T)
	if a < 1 {
		return 1
	}
	// complex{64,128} are aligned like [2]float{32,64}.
	if isComplex(T) {
		a /= 2
	}
	if a > s.maxAlign {
		return s.maxAlign
	}
	return a
}

// align returns the smallest x >= subject such that x % target == 0.
func align(subject, target int64) int64 {
	x := subject + target - 1
	return x - x%target
}

func isComplex(typ types.Type) bool {
	t, ok := typ.Underlying().(*types.Basic)
	return ok && t.Info()&types.IsComplex != 0
}
