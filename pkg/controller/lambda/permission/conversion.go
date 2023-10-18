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

package permission

import (
	svcsdk "github.com/aws/aws-sdk-go-v2/service/lambda"
	"github.com/crossplane/crossplane-runtime/pkg/meta"

	svcapitypes "github.com/crossplane-contrib/provider-aws/apis/lambda/manualv1alpha1"
	"github.com/crossplane-contrib/provider-aws/pkg/utils/pointer"
)

func generatePermission(policyDocument *policyDocument, sid string) *svcapitypes.Permission {
	cr := &svcapitypes.Permission{}

	policyStatement := policyDocument.StatementByID(sid)
	if policyStatement != nil {
		cr.Spec.ForProvider = svcapitypes.PermissionParameters{
			Action:           policyStatement.Action,
			Principal:        policyStatement.GetPrincipal(),
			EventSourceToken: policyStatement.GetEventSourceToken(),
			PrincipalOrgID:   policyStatement.GetPrincipalOrgID(),
			SourceAccount:    policyStatement.GetSourceAccount(),
			SourceArn:        policyStatement.GetSourceARN(),
		}
	}
	return cr
}

func generateAddPermissionInput(cr *svcapitypes.Permission) *svcsdk.AddPermissionInput {
	return &svcsdk.AddPermissionInput{
		Action:           &cr.Spec.ForProvider.Action,
		FunctionName:     cr.Spec.ForProvider.FunctionName,
		Principal:        &cr.Spec.ForProvider.Principal,
		StatementId:      pointer.ToOrNilIfZeroValue(meta.GetExternalName(cr)),
		EventSourceToken: cr.Spec.ForProvider.EventSourceToken,
		PrincipalOrgID:   cr.Spec.ForProvider.PrincipalOrgID,
		SourceArn:        cr.Spec.ForProvider.SourceArn,
		SourceAccount:    cr.Spec.ForProvider.SourceAccount,
	}
}

func generateRemovePermissionInput(cr *svcapitypes.Permission) *svcsdk.RemovePermissionInput {
	return &svcsdk.RemovePermissionInput{
		FunctionName: cr.Spec.ForProvider.FunctionName,
		StatementId:  pointer.ToOrNilIfZeroValue(meta.GetExternalName(cr)),
	}
}
