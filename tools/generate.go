//go:build generate
// +build generate

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

// Package tools keeps tracks of the code generator dependencies that are
// only needed during development.
package tools

import (
	_ "github.com/aws-controllers-k8s/code-generator/cmd/ack-generate" //nolint:typecheck
	_ "github.com/crossplane/crossplane-tools/cmd/angryjet"            //nolint:typecheck
	_ "github.com/jmattheis/goverter/cmd/goverter"                     //nolint:typecheck
	_ "sigs.k8s.io/controller-tools/cmd/controller-gen"                //nolint:typecheck

	// Workaround to vendor mockgen (https://github.com/golang/mock/issues/415#issuecomment-602547154)
	_ "github.com/golang/mock/mockgen" //nolint:typecheck
)
