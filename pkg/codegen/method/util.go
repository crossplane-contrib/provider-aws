/*
Copyright 2019 The Crossplane Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// Package method contains methods that may be generated for a Go type.
package method

import (
	"go/types"
)

func lookupPath(n types.Type, segments ...string) types.Type {
	current := n
	for _, segment := range segments {
		st := isStruct(current)
		if st == nil {
			return nil
		}
		current = lookupFieldByName(st, segment)
		if current == nil {
			return nil
		}
	}
	return current
}

func lookupFieldByName(st *types.Struct, name string) types.Type {
	for i := 0; i < st.NumFields(); i++ {
		f := st.Field(i)
		if f.Name() == name {
			return f.Type()
		}
	}
	return nil
}

func isStruct(t types.Type) *types.Struct {
	if n, ok := t.(*types.Named); ok {
		t = n.Underlying()
	}
	if st, ok := t.(*types.Struct); ok {
		return st
	}
	return nil
}

func isString(t types.Type) bool {
	if t == nil {
		return false
	}
	if b, ok := t.(*types.Basic); ok {
		return b.Info() == types.IsString
	}
	return false
}

func isStringPtr(t types.Type) bool {
	if p, ok := t.(*types.Pointer); ok {
		if b, ok := p.Elem().(*types.Basic); ok {
			return b.Info() == types.IsString
		}
	}
	return false
}

func isNamedPtr(t types.Type) *types.Named {
	if p, ok := t.(*types.Pointer); ok {
		if n, ok := p.Elem().(*types.Named); ok {
			return n
		}
	}
	return nil
}

func isNamedStruct(t types.Type) (*types.Named, *types.Struct) {
	if n, ok := t.(*types.Named); ok {
		if st, ok := n.Underlying().(*types.Struct); ok {
			return n, st
		}
	}
	return nil, nil
}

func isNamedStructPtr(t types.Type) (*types.Named, *types.Struct) {
	if n := isNamedPtr(t); n != nil {
		return isNamedStruct(n)
	}
	return nil, nil
}
