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

// Package comments extracts and parses comments from a package.
package comments

import (
	"go/ast"
	"go/token"
	"go/types"
	"strings"

	"golang.org/x/tools/go/packages"
)

// DefaultMarkerPrefix that is commonly used by comment markers.
const DefaultMarkerPrefix = "+"

type fl struct {
	Filename string
	Line     int
}

// Comments for a particular package.
type Comments struct {
	groups map[fl]*ast.CommentGroup
	fset   *token.FileSet
}

// In returns all comments in a particular package.
func In(p *packages.Package) Comments {
	groups := map[fl]*ast.CommentGroup{}

	for _, f := range p.Syntax {
		for _, g := range f.Comments {
			p := p.Fset.Position(g.End())
			groups[fl{Filename: p.Filename, Line: p.Line}] = g
		}
	}
	return Comments{groups: groups, fset: p.Fset}
}

// For returns the comments for the supplied Object, if any.
func (c Comments) For(o types.Object) string {
	p := c.fset.Position(o.Pos())
	return c.groups[fl{Filename: p.Filename, Line: p.Line - 1}].Text()
}

// Before returns the comments before the supplied Object, if any. A comment is
// deemed to be 'before' (rather than 'for') an Object if it ends exactly one
// blank line above where the Object (including its comment, if any) begins.
func (c Comments) Before(o types.Object) string {
	p := c.fset.Position(o.Pos())
	g := c.groups[fl{Filename: p.Filename, Line: p.Line - 1}]

	if g == nil {
		// No comment group ends immediately before this object. Check for one
		// ending two lines back.
		return c.groups[fl{Filename: p.Filename, Line: p.Line - 2}].Text()
	}

	// A comment group ends immediately before this object. Check for another
	// one ending two lines back from where it starts.
	start := c.fset.Position(g.List[0].Slash)
	return c.groups[fl{Filename: start.Filename, Line: start.Line - 2}].Text()
}

// Markers are comments that begin with a special character (typically
// DefaultMarkerPrefix). Comment markers that contain '=' are considered to be
// key=value pairs, represented as one map key with a slice of multiple values.
type Markers map[string][]string

// ParseMarkers parses comment markers from the supplied comment using the
// DefaultMarkerPrefix.
func ParseMarkers(comment string) Markers {
	return ParseMarkersWithPrefix(DefaultMarkerPrefix, comment)
}

// ParseMarkersWithPrefix parses comment markers from the supplied comment. Any
// line that begins with the supplied prefix is considered a comment marker. For
// example using marker prefix '+' the following comments:
//
// +key:value1
// +key:value2
//
// Would be parsed as Markers{"key": []string{"value1", "value2"}}
func ParseMarkersWithPrefix(prefix, comment string) Markers {
	m := map[string][]string{}

	for _, line := range strings.Split(comment, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		if !strings.HasPrefix(line, prefix) {
			continue
		}
		kv := strings.SplitN(line[len(prefix):], "=", 2)
		k, v := kv[0], ""
		if len(kv) > 1 {
			v = kv[1]
		}
		m[k] = append(m[k], v)
	}

	return m
}
