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

package loggroup

import (
	"github.com/aws/aws-sdk-go/aws/awserr"
	svcsdk "github.com/aws/aws-sdk-go/service/cloudwatchlogs"

	svcapitypes "github.com/crossplane-contrib/provider-aws/apis/cloudwatchlogs/v1alpha1"
)

// NOTE(muvaf): We return pointers in case the function needs to start with an
// empty object, hence need to return a new pointer.

// GenerateDescribeLogGroupsInput returns input for read
// operation.
func GenerateDescribeLogGroupsInput(cr *svcapitypes.LogGroup) *svcsdk.DescribeLogGroupsInput {
	res := &svcsdk.DescribeLogGroupsInput{}

	return res
}

// GenerateLogGroup returns the current state in the form of *svcapitypes.LogGroup.
func GenerateLogGroup(resp *svcsdk.DescribeLogGroupsOutput) *svcapitypes.LogGroup {
	cr := &svcapitypes.LogGroup{}

	found := false
	for _, elem := range resp.LogGroups {
		if elem.CreationTime != nil {
			cr.Status.AtProvider.CreationTime = elem.CreationTime
		} else {
			cr.Status.AtProvider.CreationTime = nil
		}
		if elem.KmsKeyId != nil {
			cr.Status.AtProvider.KMSKeyID = elem.KmsKeyId
		} else {
			cr.Status.AtProvider.KMSKeyID = nil
		}
		if elem.LogGroupName != nil {
			cr.Spec.ForProvider.LogGroupName = elem.LogGroupName
		} else {
			cr.Spec.ForProvider.LogGroupName = nil
		}
		if elem.MetricFilterCount != nil {
			cr.Status.AtProvider.MetricFilterCount = elem.MetricFilterCount
		} else {
			cr.Status.AtProvider.MetricFilterCount = nil
		}
		if elem.RetentionInDays != nil {
			cr.Status.AtProvider.RetentionInDays = elem.RetentionInDays
		} else {
			cr.Status.AtProvider.RetentionInDays = nil
		}
		if elem.StoredBytes != nil {
			cr.Status.AtProvider.StoredBytes = elem.StoredBytes
		} else {
			cr.Status.AtProvider.StoredBytes = nil
		}
		found = true
		break
	}
	if !found {
		return cr
	}

	return cr
}

// GenerateCreateLogGroupInput returns a create input.
func GenerateCreateLogGroupInput(cr *svcapitypes.LogGroup) *svcsdk.CreateLogGroupInput {
	res := &svcsdk.CreateLogGroupInput{}

	if cr.Spec.ForProvider.LogGroupName != nil {
		res.SetLogGroupName(*cr.Spec.ForProvider.LogGroupName)
	}
	if cr.Spec.ForProvider.Tags != nil {
		f1 := map[string]*string{}
		for f1key, f1valiter := range cr.Spec.ForProvider.Tags {
			var f1val string
			f1val = *f1valiter
			f1[f1key] = &f1val
		}
		res.SetTags(f1)
	}

	return res
}

// GenerateDeleteLogGroupInput returns a deletion input.
func GenerateDeleteLogGroupInput(cr *svcapitypes.LogGroup) *svcsdk.DeleteLogGroupInput {
	res := &svcsdk.DeleteLogGroupInput{}

	if cr.Spec.ForProvider.LogGroupName != nil {
		res.SetLogGroupName(*cr.Spec.ForProvider.LogGroupName)
	}

	return res
}

// IsNotFound returns whether the given error is of type NotFound or not.
func IsNotFound(err error) bool {
	awsErr, ok := err.(awserr.Error)
	return ok && awsErr.Code() == "UNKNOWN"
}
