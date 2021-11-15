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

package dbparametergroup

import (
	"github.com/aws/aws-sdk-go/aws/awserr"
	svcsdk "github.com/aws/aws-sdk-go/service/rds"

	svcapitypes "github.com/crossplane/provider-aws/apis/rds/v1alpha1"
)

// NOTE(muvaf): We return pointers in case the function needs to start with an
// empty object, hence need to return a new pointer.

// GenerateDescribeDBParameterGroupsInput returns input for read
// operation.
func GenerateDescribeDBParameterGroupsInput(cr *svcapitypes.DBParameterGroup) *svcsdk.DescribeDBParameterGroupsInput {
	res := &svcsdk.DescribeDBParameterGroupsInput{}

	return res
}

// GenerateDBParameterGroup returns the current state in the form of *svcapitypes.DBParameterGroup.
func GenerateDBParameterGroup(resp *svcsdk.DescribeDBParameterGroupsOutput) *svcapitypes.DBParameterGroup {
	cr := &svcapitypes.DBParameterGroup{}

	found := false
	for _, elem := range resp.DBParameterGroups {
		if elem.DBParameterGroupArn != nil {
			cr.Status.AtProvider.DBParameterGroupARN = elem.DBParameterGroupArn
		} else {
			cr.Status.AtProvider.DBParameterGroupARN = nil
		}
		if elem.DBParameterGroupFamily != nil {
			cr.Spec.ForProvider.DBParameterGroupFamily = elem.DBParameterGroupFamily
		} else {
			cr.Spec.ForProvider.DBParameterGroupFamily = nil
		}
		if elem.DBParameterGroupName != nil {
			cr.Status.AtProvider.DBParameterGroupName = elem.DBParameterGroupName
		} else {
			cr.Status.AtProvider.DBParameterGroupName = nil
		}
		if elem.Description != nil {
			cr.Spec.ForProvider.Description = elem.Description
		} else {
			cr.Spec.ForProvider.Description = nil
		}
		found = true
		break
	}
	if !found {
		return cr
	}

	return cr
}

// GenerateCreateDBParameterGroupInput returns a create input.
func GenerateCreateDBParameterGroupInput(cr *svcapitypes.DBParameterGroup) *svcsdk.CreateDBParameterGroupInput {
	res := &svcsdk.CreateDBParameterGroupInput{}

	if cr.Spec.ForProvider.DBParameterGroupFamily != nil {
		res.SetDBParameterGroupFamily(*cr.Spec.ForProvider.DBParameterGroupFamily)
	}
	if cr.Spec.ForProvider.Description != nil {
		res.SetDescription(*cr.Spec.ForProvider.Description)
	}
	if cr.Spec.ForProvider.Tags != nil {
		f2 := []*svcsdk.Tag{}
		for _, f2iter := range cr.Spec.ForProvider.Tags {
			f2elem := &svcsdk.Tag{}
			if f2iter.Key != nil {
				f2elem.SetKey(*f2iter.Key)
			}
			if f2iter.Value != nil {
				f2elem.SetValue(*f2iter.Value)
			}
			f2 = append(f2, f2elem)
		}
		res.SetTags(f2)
	}

	return res
}

// GenerateModifyDBParameterGroupInput returns an update input.
func GenerateModifyDBParameterGroupInput(cr *svcapitypes.DBParameterGroup) *svcsdk.ModifyDBParameterGroupInput {
	res := &svcsdk.ModifyDBParameterGroupInput{}

	return res
}

// GenerateDeleteDBParameterGroupInput returns a deletion input.
func GenerateDeleteDBParameterGroupInput(cr *svcapitypes.DBParameterGroup) *svcsdk.DeleteDBParameterGroupInput {
	res := &svcsdk.DeleteDBParameterGroupInput{}

	return res
}

// IsNotFound returns whether the given error is of type NotFound or not.
func IsNotFound(err error) bool {
	awsErr, ok := err.(awserr.Error)
	return ok && awsErr.Code() == "DBParameterGroupNotFound"
}
