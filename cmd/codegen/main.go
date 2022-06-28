/*
Copyright 2022 The Crossplane Authors.

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

package main

import (
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/crossplane/crossplane-runtime/pkg/logging"
	"github.com/pkg/errors"
	"golang.org/x/tools/go/packages"
	"gopkg.in/alecthomas/kingpin.v2"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	"github.com/crossplane-contrib/provider-aws/pkg/codegen/comments"
	"github.com/crossplane-contrib/provider-aws/pkg/codegen/generate"
	"github.com/crossplane-contrib/provider-aws/pkg/codegen/match"
	"github.com/crossplane-contrib/provider-aws/pkg/codegen/method"
)

const (
	// LoadMode used to load all packages.
	LoadMode = packages.NeedName | packages.NeedFiles | packages.NeedImports | packages.NeedDeps | packages.NeedTypes | packages.NeedSyntax

	// DisableMarker used to disable generation of managed resource methods for
	// a type that otherwise appears to be a managed resource that is missing a
	// subnet of its methods.
	DisableMarker = "provider-aws:generate:methods"
)

func main() {
	var (
		app = kingpin.New(filepath.Base(os.Args[0]), "Generates provider-aws API type methods.").DefaultEnvars()

		methodsets      = app.Command("generate-methodsets", "Generate provider-aws method sets.")
		headerFile      = methodsets.Flag("header-file", "The contents of this file will be added to the top of all generated files.").ExistingFile()
		filenameManaged = methodsets.Flag("filename-managed", "The filename of generated tagged resource files.").Default("zz_generated.tagged.go").String()
		pattern         = methodsets.Arg("packages", "Package(s) for which to generate methods, for example github.com/crossplane/crossplane/apis/...").String()
	)
	kingpin.MustParse(app.Parse(os.Args[1:]))

	pkgs, err := packages.Load(&packages.Config{Mode: LoadMode}, *pattern)
	kingpin.FatalIfError(err, "cannot load packages %s", *pattern)

	zl := zap.New(zap.UseDevMode(true))
	log := logging.NewLogrLogger(zl.WithName("provider-aws"))

	header := ""
	if *headerFile != "" {
		h, err := ioutil.ReadFile(*headerFile)
		kingpin.FatalIfError(err, "cannot read header file %s", *headerFile)
		header = string(h)
	}

	for _, p := range pkgs {
		for _, err := range p.Errors {
			kingpin.FatalIfError(err, "error loading packages using pattern %s", *pattern)
		}
		kingpin.FatalIfError(GenerateTagged(*filenameManaged, header, p, log), "cannot write tagged resource method set for package %s", p.PkgPath)
	}
}

// GenerateTagged generates the apis.Tagged method set.
func GenerateTagged(filename, header string, p *packages.Package, log logging.Logger) error {
	receiver := "mg"

	methods := method.Set{
		"AddTag": method.NewAddTag(receiver, log),
	}

	err := generate.WriteMethods(p, methods, filepath.Join(filepath.Dir(p.GoFiles[0]), filename),
		generate.WithHeaders(header),
		generate.WithMatcher(match.AllOf(
			match.Managed(),
			match.DoesNotHaveMarker(comments.In(p), DisableMarker, "false")),
		),
	)

	return errors.Wrap(err, "cannot write tagged resource methods")
}
