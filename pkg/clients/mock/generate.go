//go:build generate

/*
Copyright 2021 The Crossplane Authors.

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

//go:generate go run -tags generate github.com/golang/mock/mockgen -package kube -destination ./kube/mock.go sigs.k8s.io/controller-runtime/pkg/client Client

//go:generate go run -tags generate github.com/golang/mock/mockgen -package cognitoidentityprovider -destination ./cognitoidentityprovider/mock.go github.com/crossplane-contrib/provider-aws/pkg/clients/cognitoidentityprovider ResolverService

package mock

import (
	// Workaround to vendor mockgen (https://github.com/golang/mock/issues/415#issuecomment-602547154)
	_ "github.com/golang/mock/mockgen" //nolint:typecheck
)
