/*
Copyright 2020 The Crossplane Authors.

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

// NOTE(negz): Please do not add methods to this package. See below issue.
// https://github.com/crossplaneio/crossplane-runtime/issues/1

package utils

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/crossplaneio/crossplane-runtime/pkg/meta"

	awsv1alpha3 "github.com/crossplaneio/stack-aws/apis/v1alpha3"
	awsclients "github.com/crossplaneio/stack-aws/pkg/clients"
)

// RetrieveAwsConfigFromProvider retrieves the aws config from the given aws provider reference
func RetrieveAwsConfigFromProvider(ctx context.Context, client client.Reader, providerRef *corev1.ObjectReference) (*aws.Config, error) {
	p := &awsv1alpha3.Provider{}
	n := meta.NamespacedNameOf(providerRef)
	if err := client.Get(ctx, n, p); err != nil {
		return nil, errors.Wrapf(err, "cannot get provider %s", n)
	}

	secret := &corev1.Secret{}
	n = types.NamespacedName{Namespace: p.Spec.CredentialsSecretRef.Namespace, Name: p.Spec.CredentialsSecretRef.Name}
	err := client.Get(ctx, n, secret)
	if err != nil {
		return nil, errors.Wrapf(err, "cannot get provider secret %s", n)
	}

	cfg, err := awsclients.LoadConfig(secret.Data[p.Spec.CredentialsSecretRef.Key], awsclients.DefaultSection, p.Spec.Region)

	return cfg, errors.Wrap(err, "cannot create new AWS configuration")
}
