//go:build generate
// +build generate

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

//go:generate go run -tags generate github.com/golang/mock/mockgen --build_flags=--mod=mod -copyright_file ../../../hack/boilerplate.go.txt -package ec2iface -destination ./ec2iface/zz_ec2_api.go github.com/aws/aws-sdk-go/service/ec2/ec2iface EC2API
//go:generate go run -tags generate github.com/golang/mock/mockgen --build_flags=--mod=mod -copyright_file ../../../hack/boilerplate.go.txt -package eksiface  -destination ./eksiface/zz_eks_api.go github.com/aws/aws-sdk-go/service/eks/eksiface EKSAPI
//go:generate go run -tags generate github.com/golang/mock/mockgen --build_flags=--mod=mod -copyright_file ../../../hack/boilerplate.go.txt -package kmsiface -destination ./kmsiface/zz_kms_api.go github.com/aws/aws-sdk-go/service/kms/kmsiface KMSAPI
//go:generate go run -tags generate github.com/golang/mock/mockgen --build_flags=--mod=mod -copyright_file ../../../hack/boilerplate.go.txt -package transferiface -destination ./transferiface/zz_transfer_api.go github.com/aws/aws-sdk-go/service/transfer/transferiface TransferAPI
//go:generate go run -tags generate github.com/golang/mock/mockgen --build_flags=--mod=mod -copyright_file ../../../hack/boilerplate.go.txt -package kube -destination ./kube/zz_client.go sigs.k8s.io/controller-runtime/pkg/client Client
//go:generate go run -tags generate github.com/golang/mock/mockgen --build_flags=--mod=mod -copyright_file ../../../hack/boilerplate.go.txt -package cognitoidentityprovider -destination ./cognitoidentityprovider/zz_resolver_service.go github.com/crossplane-contrib/provider-aws/pkg/clients/cognitoidentityprovider ResolverService

package mock

import (
	// Workaround to vendor mockgen (https://github.com/golang/mock/issues/415#issuecomment-602547154)
	_ "github.com/golang/mock/mockgen" //nolint:typecheck
)
