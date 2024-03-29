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

// Code generated by ack-generate. DO NOT EDIT.

package grant

import (
	"github.com/aws/aws-sdk-go/aws/awserr"
	svcsdk "github.com/aws/aws-sdk-go/service/kms"

	svcapitypes "github.com/crossplane-contrib/provider-aws/apis/kms/v1alpha1"
)

// NOTE(muvaf): We return pointers in case the function needs to start with an
// empty object, hence need to return a new pointer.

// GenerateListGrantsInput returns input for read
// operation.
func GenerateListGrantsInput(cr *svcapitypes.Grant) *svcsdk.ListGrantsInput {
	res := &svcsdk.ListGrantsInput{}

	if cr.Status.AtProvider.GrantID != nil {
		res.SetGrantId(*cr.Status.AtProvider.GrantID)
	}
	if cr.Spec.ForProvider.GranteePrincipal != nil {
		res.SetGranteePrincipal(*cr.Spec.ForProvider.GranteePrincipal)
	}

	return res
}

// GenerateGrant returns the current state in the form of *svcapitypes.Grant.
func GenerateGrant(resp *svcsdk.ListGrantsResponse) *svcapitypes.Grant {
	cr := &svcapitypes.Grant{}

	found := false
	for _, elem := range resp.Grants {
		if elem.Constraints != nil {
			f0 := &svcapitypes.GrantConstraints{}
			if elem.Constraints.EncryptionContextEquals != nil {
				f0f0 := map[string]*string{}
				for f0f0key, f0f0valiter := range elem.Constraints.EncryptionContextEquals {
					var f0f0val string
					f0f0val = *f0f0valiter
					f0f0[f0f0key] = &f0f0val
				}
				f0.EncryptionContextEquals = f0f0
			}
			if elem.Constraints.EncryptionContextSubset != nil {
				f0f1 := map[string]*string{}
				for f0f1key, f0f1valiter := range elem.Constraints.EncryptionContextSubset {
					var f0f1val string
					f0f1val = *f0f1valiter
					f0f1[f0f1key] = &f0f1val
				}
				f0.EncryptionContextSubset = f0f1
			}
			cr.Spec.ForProvider.Constraints = f0
		} else {
			cr.Spec.ForProvider.Constraints = nil
		}
		if elem.GrantId != nil {
			cr.Status.AtProvider.GrantID = elem.GrantId
		} else {
			cr.Status.AtProvider.GrantID = nil
		}
		if elem.GranteePrincipal != nil {
			cr.Spec.ForProvider.GranteePrincipal = elem.GranteePrincipal
		} else {
			cr.Spec.ForProvider.GranteePrincipal = nil
		}
		if elem.Name != nil {
			cr.Spec.ForProvider.Name = elem.Name
		} else {
			cr.Spec.ForProvider.Name = nil
		}
		if elem.Operations != nil {
			f7 := []*string{}
			for _, f7iter := range elem.Operations {
				var f7elem string
				f7elem = *f7iter
				f7 = append(f7, &f7elem)
			}
			cr.Spec.ForProvider.Operations = f7
		} else {
			cr.Spec.ForProvider.Operations = nil
		}
		if elem.RetiringPrincipal != nil {
			cr.Spec.ForProvider.RetiringPrincipal = elem.RetiringPrincipal
		} else {
			cr.Spec.ForProvider.RetiringPrincipal = nil
		}
		found = true
		break
	}
	if !found {
		return cr
	}

	return cr
}

// GenerateCreateGrantInput returns a create input.
func GenerateCreateGrantInput(cr *svcapitypes.Grant) *svcsdk.CreateGrantInput {
	res := &svcsdk.CreateGrantInput{}

	if cr.Spec.ForProvider.Constraints != nil {
		f0 := &svcsdk.GrantConstraints{}
		if cr.Spec.ForProvider.Constraints.EncryptionContextEquals != nil {
			f0f0 := map[string]*string{}
			for f0f0key, f0f0valiter := range cr.Spec.ForProvider.Constraints.EncryptionContextEquals {
				var f0f0val string
				f0f0val = *f0f0valiter
				f0f0[f0f0key] = &f0f0val
			}
			f0.SetEncryptionContextEquals(f0f0)
		}
		if cr.Spec.ForProvider.Constraints.EncryptionContextSubset != nil {
			f0f1 := map[string]*string{}
			for f0f1key, f0f1valiter := range cr.Spec.ForProvider.Constraints.EncryptionContextSubset {
				var f0f1val string
				f0f1val = *f0f1valiter
				f0f1[f0f1key] = &f0f1val
			}
			f0.SetEncryptionContextSubset(f0f1)
		}
		res.SetConstraints(f0)
	}
	if cr.Spec.ForProvider.GrantTokens != nil {
		f1 := []*string{}
		for _, f1iter := range cr.Spec.ForProvider.GrantTokens {
			var f1elem string
			f1elem = *f1iter
			f1 = append(f1, &f1elem)
		}
		res.SetGrantTokens(f1)
	}
	if cr.Spec.ForProvider.GranteePrincipal != nil {
		res.SetGranteePrincipal(*cr.Spec.ForProvider.GranteePrincipal)
	}
	if cr.Spec.ForProvider.Name != nil {
		res.SetName(*cr.Spec.ForProvider.Name)
	}
	if cr.Spec.ForProvider.Operations != nil {
		f4 := []*string{}
		for _, f4iter := range cr.Spec.ForProvider.Operations {
			var f4elem string
			f4elem = *f4iter
			f4 = append(f4, &f4elem)
		}
		res.SetOperations(f4)
	}
	if cr.Spec.ForProvider.RetiringPrincipal != nil {
		res.SetRetiringPrincipal(*cr.Spec.ForProvider.RetiringPrincipal)
	}

	return res
}

// IsNotFound returns whether the given error is of type NotFound or not.
func IsNotFound(err error) bool {
	awsErr, ok := err.(awserr.Error)
	return ok && awsErr.Code() == "UNKNOWN"
}
