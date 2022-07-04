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
	"fmt"
	"go/token"
	"go/types"
	"sort"

	"github.com/crossplane/crossplane-runtime/pkg/logging"
	"github.com/dave/jennifer/jen"

	"github.com/crossplane-contrib/provider-aws/pkg/codegen/fields"
)

// New is a function that adds a method on the supplied object in the
// supplied file.
type New func(f *jen.File, o types.Object)

// A Set is a map of method names to the New functions that produce
// them.
type Set map[string]New

// Write the method Set for the supplied Object to the supplied file. Methods
// are filtered by the supplied Filter.
func (s Set) Write(f *jen.File, o types.Object, mf Filter) {
	names := make([]string, 0, len(s))
	for name := range s {
		names = append(names, name)
	}
	sort.Strings(names)

	for _, name := range names {
		if mf(o, name) {
			continue
		}
		s[name](f, o)
	}
}

// A Filter is a function that determines whether a method should be written for
// the supplied object. It returns true if the method should be filtered.
type Filter func(o types.Object, methodName string) bool

// DefinedOutside returns a MethodFilter that returns true if the supplied
// object has a method with the supplied name that is not defined in the
// supplied filename. The object's filename is determined using the supplied
// FileSet.
func DefinedOutside(fs *token.FileSet, filename string) Filter {
	return func(o types.Object, name string) bool {
		s := types.NewMethodSet(types.NewPointer(o.Type()))
		for i := 0; i < s.Len(); i++ {
			mo := s.At(i).Obj()
			if mo.Name() != name {
				continue
			}
			if fs.Position(mo.Pos()).Filename != filename {
				return true
			}
		}
		return false
	}
}

const (
	localFieldKey         = "key"
	localFieldNewValue    = "value"
	localFieldNewTag      = "newTag"
	localFieldOldValue    = "oldValue"
	localFieldEntryExists = "exists"
	localFieldTagIterator = "ta"
	localFieldUpdated     = "updated"

	localFieldIndexA = "i"
	localFieldIndexB = "j"

	localFieldTagA = "ta"
	localFieldTagB = "tb"
)

// NewAddTag creates a new AddTag function generator.
func NewAddTag(receiver string, log logging.Logger) New {
	return func(f *jen.File, o types.Object) {
		n, ok := o.Type().(*types.Named)
		if !ok {
			return
		}
		tagObj := lookupPath(n, fields.NameSpec, fields.NameSpecForProvider, fields.NameTags)
		if tagObj == nil {
			log.Debug(fmt.Sprintf("No field Spec.ForProvider.Tags found for managed resource %s in package %s", n.Obj().Name(), n.Obj().Pkg().Path()))
			return
		}

		funcHeader := func() *jen.Statement {
			f.Commentf("AddTag adds a tag to this %s. If it already exists, it will be overwritten.", o.Name())
			f.Comment("It returns true if the tag has been added/changed. Otherwise false.")
			return f.
				Func().
				Params(jen.Id(receiver).Op("*").Id(o.Name())).
				Id("AddTag").
				Params(
					jen.Id(localFieldKey).String(),
					jen.Id(localFieldNewValue).String(),
				).
				Bool()
		}

		// The following tag definitions are covered by this generator:
		//
		// Assuming tags are defined at Spec.ForProvider.Tags.
		//
		// - map[string]string
		// - map[string]*string
		// - []Tag  {string, string}
		// - []*Tag {string, string}
		// - []Tag  {*string, *string}
		// - []*Tag {*string, *string}
		// - []Tag  (with string key and *string value)
		// - []*Tag  (with string key and *string value)
		//
		// NOTE: There are resources (i.e. S3 buckets) that have a custom
		// tagging implementation. In this case it is necessary to define the
		// AddTag method manually.

		tagFieldAccessor := func() *jen.Statement {
			return jen.Id(receiver).Dot(fields.NameSpec).Dot(fields.NameSpecForProvider).Dot(fields.NameTags)
		}

		if ref := isTagMap(tagObj); ref != nil {
			funcHeader().Block(
				jen.If(tagFieldAccessor().Op("==").Nil()).Block(
					tagFieldAccessor().Op("=").Add(ref.newMap),
					jen.Return(jen.True()),
				),
				jen.List(jen.Id(localFieldOldValue), jen.Id(localFieldEntryExists)).Op(":=").Add(tagFieldAccessor()).Index(jen.Id(localFieldKey)),
				jen.If(jen.Op("!").Id(localFieldEntryExists).Op("||").Add(ref.oldValueRef).Op("!=").Id(localFieldNewValue)).Block(
					tagFieldAccessor().Index(jen.Id(localFieldKey)).Op("=").Add(ref.assignValue),
					jen.Return(jen.True()),
				),
				jen.Return(jen.False()),
			)
		} else if ref := isTagSlice(tagObj); ref != nil {
			funcHeader().Block(
				jen.Id(localFieldNewTag).Op(":=").Add(ref.newTag),
				jen.Id(localFieldUpdated).Op(":=").False(),
				jen.For(jen.List(jen.Id("i"), jen.Id(localFieldTagIterator)).Op(":=").Range().Add(tagFieldAccessor())).Block(
					jen.If(ref.checkEqualKey).Block(
						jen.If(ref.checkEqualValue).Block(
							jen.Return().False(),
						),
						tagFieldAccessor().Index(jen.Id("i")).Op("=").Id(localFieldNewTag),
						jen.Id(localFieldUpdated).Op("=").True(),
						jen.Break(),
					),
				),
				jen.If(jen.Op("!").Id(localFieldUpdated)).Block(
					tagFieldAccessor().Op("=").Id("append").Call(
						tagFieldAccessor(),
						jen.Id(localFieldNewTag),
					),
				),
				jen.Qual("sort", "Slice").Call(tagFieldAccessor(), jen.Func().Params(jen.Id(localFieldIndexA).Int(), jen.Id(localFieldIndexB).Int()).Bool().Block(
					jen.Id(localFieldTagA).Op(":=").Add(tagFieldAccessor()).Index(jen.Id(localFieldIndexA)),
					jen.Id(localFieldTagB).Op(":=").Add(tagFieldAccessor()).Index(jen.Id(localFieldIndexB)),
					ref.compareTagStruct(jen.Id(localFieldTagA), jen.Id(localFieldTagB)),
					ref.compareTagKeys(jen.Id(localFieldTagA), jen.Id(localFieldTagB)),
				)),
				jen.Return().True(),
			)
		}
	}
}

const (
	packageK8sUtilPointer = "k8s.io/utils/pointer"
)

type tagMapReference struct {
	newMap      *jen.Statement
	oldValueRef *jen.Statement
	assignValue *jen.Statement
}

func isTagMap(t types.Type) *tagMapReference {
	m, ok := t.(*types.Map)
	if !ok {
		return nil
	}
	ref := &tagMapReference{}
	if !isString(m.Key()) {
		return nil
	}
	switch {
	case isString(m.Elem()):
		ref.oldValueRef = jen.Id(localFieldOldValue)
		ref.assignValue = jen.Id(localFieldNewValue)
		ref.newMap = jen.Map(jen.String()).String().Values(jen.Dict{
			jen.Id(localFieldKey): jen.Id(localFieldNewValue),
		})
	case isStringPtr(m.Elem()):
		ref.oldValueRef = jen.Qual(packageK8sUtilPointer, "StringDeref").Call(jen.Id(localFieldOldValue), jen.Lit(""))
		ref.assignValue = jen.Op("&").Id(localFieldNewValue)
		ref.newMap = jen.Map(jen.String()).Op("*").String().Values(jen.Dict{
			jen.Id(localFieldKey): jen.Op("&").Id(localFieldNewValue),
		})
	default:
		return nil
	}
	return ref
}

type tagSliceReference struct {
	newTag           *jen.Statement
	checkEqualKey    *jen.Statement
	checkEqualValue  *jen.Statement
	compareTagStruct func(tagA, tagB *jen.Statement) *jen.Statement
	compareTagKeys   func(tagA, tagB *jen.Statement) *jen.Statement
}

func isTagSlice(t types.Type) *tagSliceReference {
	s, ok := t.(*types.Slice)
	if !ok {
		return nil
	}

	ref := &tagSliceReference{}

	var named *types.Named
	var elem *types.Struct
	var newTag *jen.Statement
	var baseCheck *jen.Statement

	if named, elem = isNamedStruct(s.Elem()); elem != nil {
		newTag = jen.Empty()
		baseCheck = jen.Empty()
		ref.compareTagStruct = func(tagA, tagB *jen.Statement) *jen.Statement {
			return jen.Empty()
		}
	} else if named, elem = isNamedStructPtr(s.Elem()); elem != nil {
		newTag = jen.Op("&")
		baseCheck = jen.Id(localFieldTagIterator).Op("!=").Nil().Op("&&")

		ref.compareTagStruct = func(tagA, tagB *jen.Statement) *jen.Statement {
			return jen.If(tagA.Op("==").Nil()).Block(
				jen.Return(jen.True()),
			).Else().If(tagB.Op("==").Nil()).Block(
				jen.Return(jen.False()),
			)
		}
	} else {
		return nil
	}

	key := lookupFieldByName(elem, fields.NameTagKey)
	value := lookupFieldByName(elem, fields.NameTagValue)

	var keyValue *jen.Statement
	switch {
	case isString(key):
		keyValue = jen.Id(localFieldKey)
		ref.checkEqualKey = baseCheck.
			Id(localFieldTagIterator).Dot(fields.NameTagKey).
			Op("==").
			Id(localFieldKey)

		ref.compareTagKeys = func(tagA, tagB *jen.Statement) *jen.Statement {
			return jen.Return(tagA.Dot(fields.NameTagKey).Op("<").Add(tagB).Dot(fields.NameTagKey))
		}
	case isStringPtr(key):
		keyValue = jen.Op("&").Id(localFieldKey)
		ref.checkEqualKey = baseCheck.
			Qual(packageK8sUtilPointer, "StringDeref").Call(jen.Id(localFieldTagIterator).Dot(fields.NameTagKey), jen.Lit("")).
			Op("==").
			Id(localFieldKey)
		ref.compareTagKeys = func(tagA, tagB *jen.Statement) *jen.Statement {
			return jen.Return(jen.
				Qual(packageK8sUtilPointer, "StringDeref").Call(tagA.Clone().Dot(fields.NameTagKey), jen.Lit("")).
				Op("<").
				Qual(packageK8sUtilPointer, "StringDeref").Call(tagB.Clone().Dot(fields.NameTagKey), jen.Lit("")),
			)
		}
	default:
		return nil
	}

	var valueValue *jen.Statement
	switch {
	case isString(value):
		valueValue = jen.Id(localFieldNewValue)
		ref.checkEqualValue = jen.Id(localFieldTagIterator).Dot(fields.NameTagValue).Op("==").Id(localFieldNewValue)
	case isStringPtr(value):
		valueValue = jen.Op("&").Id(localFieldNewValue)
		ref.checkEqualValue = jen.
			Qual(packageK8sUtilPointer, "StringDeref").Call(jen.Id(localFieldTagIterator).Dot(fields.NameTagValue), jen.Lit("")).
			Op("==").
			Id(localFieldNewValue)
	default:
		return nil
	}

	ref.newTag = newTag.Qual(named.Obj().Pkg().Path(), named.Obj().Name()).Values(jen.Dict{
		jen.Id(fields.NameTagKey):   keyValue,
		jen.Id(fields.NameTagValue): valueValue,
	})
	return ref
}
