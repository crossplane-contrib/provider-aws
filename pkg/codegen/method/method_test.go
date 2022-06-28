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

package method

import (
	"fmt"
	"testing"

	"github.com/crossplane/crossplane-runtime/pkg/logging"
	"github.com/dave/jennifer/jen"
	"github.com/google/go-cmp/cmp"
	"golang.org/x/tools/go/packages"
	"golang.org/x/tools/go/packages/packagestest"
)

func TestNewAddTags(t *testing.T) {
	source := `
package v1alpha1

type Model1Parameters struct {
	Tags map[string]string
}
type Model1Spec struct {
	ForProvider Model1Parameters
}
type Model1 struct {
	Spec Model1Spec
}

type Model2Parameters struct {
	Tags map[string]*string
}
type Model2Spec struct {
	ForProvider Model2Parameters
}
type Model2 struct {
	Spec Model2Spec
}

type Tag struct {
	Key string
	Value string
}

type Model3Parameters struct {
	Tags []Tag
}
type Model3Spec struct {
	ForProvider Model3Parameters
}
type Model3 struct {
	Spec Model3Spec
}

type Model4Parameters struct {
	Tags []*Tag
}
type Model4Spec struct {
	ForProvider Model4Parameters
}
type Model4 struct {
	Spec Model4Spec
}

type TagPtr struct {
	Key *string
	Value *string
}
type Model5Parameters struct {
	Tags []TagPtr
}
type Model5Spec struct {
	ForProvider Model5Parameters
}
type Model5 struct {
	Spec Model5Spec
}

type Model6Parameters struct {
	Tags []*TagPtr
}
type Model6Spec struct {
	ForProvider Model6Parameters
}
type Model6 struct {
	Spec Model6Spec
}

type TagValuePtr struct {
	Key string
	Value *string
}
type Model7Parameters struct {
	Tags []TagValuePtr
}
type Model7Spec struct {
	ForProvider Model7Parameters
}
type Model7 struct {
	Spec Model7Spec
}

type Model8Parameters struct {
	Tags []*TagValuePtr
}
type Model8Spec struct {
	ForProvider Model8Parameters
}
type Model8 struct {
	Spec Model8Spec
}
`

	generated := `package v1alpha1

// AddTag adds a tag to this Model1. If it already exists, it will be overwritten.
// It returns true if the tag has been added/changed. Otherwise false.
func (t *Model1) AddTag(key string, value string) bool {
	if t.Spec.ForProvider.Tags == nil {
		t.Spec.ForProvider.Tags = map[string]string{key: value}
		return true
	}
	oldValue, exists := t.Spec.ForProvider.Tags[key]
	if !exists || oldValue != value {
		t.Spec.ForProvider.Tags[key] = value
		return true
	}
	return false
}

// AddTag adds a tag to this Model2. If it already exists, it will be overwritten.
// It returns true if the tag has been added/changed. Otherwise false.
func (t *Model2) AddTag(key string, value string) bool {
	if t.Spec.ForProvider.Tags == nil {
		t.Spec.ForProvider.Tags = map[string]*string{key: &value}
		return true
	}
	oldValue, exists := t.Spec.ForProvider.Tags[key]
	if !exists || oldValue == nil || *oldValue != value {
		t.Spec.ForProvider.Tags[key] = &value
		return true
	}
	return false
}

// AddTag adds a tag to this Model3. If it already exists, it will be overwritten.
// It returns true if the tag has been added/changed. Otherwise false.
func (t *Model3) AddTag(key string, value string) bool {
	newTag := Tag{
		Key:   key,
		Value: value,
	}
	for i, ta := range t.Spec.ForProvider.Tags {
		if ta.Key == key {
			if ta.Value == value {
				return false
			}
			t.Spec.ForProvider.Tags[i] = newTag
			return true
		}
	}
	t.Spec.ForProvider.Tags = append(t.Spec.ForProvider.Tags, newTag)
	return true
}

// AddTag adds a tag to this Model4. If it already exists, it will be overwritten.
// It returns true if the tag has been added/changed. Otherwise false.
func (t *Model4) AddTag(key string, value string) bool {
	newTag := &Tag{
		Key:   key,
		Value: value,
	}
	for i, ta := range t.Spec.ForProvider.Tags {
		if ta != nil && ta.Key == key {
			if ta.Value == value {
				return false
			}
			t.Spec.ForProvider.Tags[i] = newTag
			return true
		}
	}
	t.Spec.ForProvider.Tags = append(t.Spec.ForProvider.Tags, newTag)
	return true
}

// AddTag adds a tag to this Model5. If it already exists, it will be overwritten.
// It returns true if the tag has been added/changed. Otherwise false.
func (t *Model5) AddTag(key string, value string) bool {
	newTag := TagPtr{
		Key:   &key,
		Value: &value,
	}
	for i, ta := range t.Spec.ForProvider.Tags {
		if ta.Key != nil && *ta.Key == key {
			if ta.Value != nil && *ta.Value == value {
				return false
			}
			t.Spec.ForProvider.Tags[i] = newTag
			return true
		}
	}
	t.Spec.ForProvider.Tags = append(t.Spec.ForProvider.Tags, newTag)
	return true
}

// AddTag adds a tag to this Model6. If it already exists, it will be overwritten.
// It returns true if the tag has been added/changed. Otherwise false.
func (t *Model6) AddTag(key string, value string) bool {
	newTag := &TagPtr{
		Key:   &key,
		Value: &value,
	}
	for i, ta := range t.Spec.ForProvider.Tags {
		if ta != nil && ta.Key != nil && *ta.Key == key {
			if ta.Value != nil && *ta.Value == value {
				return false
			}
			t.Spec.ForProvider.Tags[i] = newTag
			return true
		}
	}
	t.Spec.ForProvider.Tags = append(t.Spec.ForProvider.Tags, newTag)
	return true
}

// AddTag adds a tag to this Model7. If it already exists, it will be overwritten.
// It returns true if the tag has been added/changed. Otherwise false.
func (t *Model7) AddTag(key string, value string) bool {
	newTag := TagValuePtr{
		Key:   key,
		Value: &value,
	}
	for i, ta := range t.Spec.ForProvider.Tags {
		if ta.Key == key {
			if ta.Value != nil && *ta.Value == value {
				return false
			}
			t.Spec.ForProvider.Tags[i] = newTag
			return true
		}
	}
	t.Spec.ForProvider.Tags = append(t.Spec.ForProvider.Tags, newTag)
	return true
}

// AddTag adds a tag to this Model8. If it already exists, it will be overwritten.
// It returns true if the tag has been added/changed. Otherwise false.
func (t *Model8) AddTag(key string, value string) bool {
	newTag := &TagValuePtr{
		Key:   key,
		Value: &value,
	}
	for i, ta := range t.Spec.ForProvider.Tags {
		if ta != nil && ta.Key == key {
			if ta.Value != nil && *ta.Value == value {
				return false
			}
			t.Spec.ForProvider.Tags[i] = newTag
			return true
		}
	}
	t.Spec.ForProvider.Tags = append(t.Spec.ForProvider.Tags, newTag)
	return true
}
`
	exported := packagestest.Export(t, packagestest.Modules, []packagestest.Module{{
		Name: "golang.org/fake",
		Files: map[string]interface{}{
			"v1alpha1/model.go": source,
		},
	}})
	defer exported.Cleanup()
	exported.Config.Mode = packages.NeedName | packages.NeedFiles | packages.NeedImports | packages.NeedDeps | packages.NeedTypes | packages.NeedSyntax
	pkgs, err := packages.Load(exported.Config, fmt.Sprintf("file=%s", exported.File("golang.org/fake", "v1alpha1/model.go")))
	if err != nil {
		t.Error(err)
	}
	f := jen.NewFilePath(pkgs[0].PkgPath)
	for _, n := range pkgs[0].Types.Scope().Names() {
		NewAddTag("t", logging.NewNopLogger())(f, pkgs[0].Types.Scope().Lookup(n))
	}
	if diff := cmp.Diff(generated, fmt.Sprintf("%#v", f)); diff != "" {
		t.Errorf("NewAddTag(): -want, +got\n%s", diff)
	}
}
