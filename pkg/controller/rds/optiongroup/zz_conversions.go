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

package optiongroup

import (
	"github.com/aws/aws-sdk-go/aws/awserr"
	svcsdk "github.com/aws/aws-sdk-go/service/rds"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	svcapitypes "github.com/crossplane-contrib/provider-aws/apis/rds/v1alpha1"
)

// NOTE(muvaf): We return pointers in case the function needs to start with an
// empty object, hence need to return a new pointer.

// GenerateDescribeOptionGroupsInput returns input for read
// operation.
func GenerateDescribeOptionGroupsInput(cr *svcapitypes.OptionGroup) *svcsdk.DescribeOptionGroupsInput {
	res := &svcsdk.DescribeOptionGroupsInput{}

	if cr.Spec.ForProvider.EngineName != nil {
		res.SetEngineName(*cr.Spec.ForProvider.EngineName)
	}
	if cr.Spec.ForProvider.MajorEngineVersion != nil {
		res.SetMajorEngineVersion(*cr.Spec.ForProvider.MajorEngineVersion)
	}
	if cr.Status.AtProvider.OptionGroupName != nil {
		res.SetOptionGroupName(*cr.Status.AtProvider.OptionGroupName)
	}

	return res
}

// GenerateOptionGroup returns the current state in the form of *svcapitypes.OptionGroup.
func GenerateOptionGroup(resp *svcsdk.DescribeOptionGroupsOutput) *svcapitypes.OptionGroup {
	cr := &svcapitypes.OptionGroup{}

	found := false
	for _, elem := range resp.OptionGroupsList {
		if elem.AllowsVpcAndNonVpcInstanceMemberships != nil {
			cr.Status.AtProvider.AllowsVPCAndNonVPCInstanceMemberships = elem.AllowsVpcAndNonVpcInstanceMemberships
		} else {
			cr.Status.AtProvider.AllowsVPCAndNonVPCInstanceMemberships = nil
		}
		if elem.CopyTimestamp != nil {
			cr.Status.AtProvider.CopyTimestamp = &metav1.Time{*elem.CopyTimestamp}
		} else {
			cr.Status.AtProvider.CopyTimestamp = nil
		}
		if elem.EngineName != nil {
			cr.Spec.ForProvider.EngineName = elem.EngineName
		} else {
			cr.Spec.ForProvider.EngineName = nil
		}
		if elem.MajorEngineVersion != nil {
			cr.Spec.ForProvider.MajorEngineVersion = elem.MajorEngineVersion
		} else {
			cr.Spec.ForProvider.MajorEngineVersion = nil
		}
		if elem.OptionGroupArn != nil {
			cr.Status.AtProvider.OptionGroupARN = elem.OptionGroupArn
		} else {
			cr.Status.AtProvider.OptionGroupARN = nil
		}
		if elem.OptionGroupDescription != nil {
			cr.Spec.ForProvider.OptionGroupDescription = elem.OptionGroupDescription
		} else {
			cr.Spec.ForProvider.OptionGroupDescription = nil
		}
		if elem.OptionGroupName != nil {
			cr.Status.AtProvider.OptionGroupName = elem.OptionGroupName
		} else {
			cr.Status.AtProvider.OptionGroupName = nil
		}
		if elem.Options != nil {
			f7 := []*svcapitypes.Option{}
			for _, f7iter := range elem.Options {
				f7elem := &svcapitypes.Option{}
				if f7iter.DBSecurityGroupMemberships != nil {
					f7elemf0 := []*svcapitypes.DBSecurityGroupMembership{}
					for _, f7elemf0iter := range f7iter.DBSecurityGroupMemberships {
						f7elemf0elem := &svcapitypes.DBSecurityGroupMembership{}
						if f7elemf0iter.DBSecurityGroupName != nil {
							f7elemf0elem.DBSecurityGroupName = f7elemf0iter.DBSecurityGroupName
						}
						if f7elemf0iter.Status != nil {
							f7elemf0elem.Status = f7elemf0iter.Status
						}
						f7elemf0 = append(f7elemf0, f7elemf0elem)
					}
					f7elem.DBSecurityGroupMemberships = f7elemf0
				}
				if f7iter.OptionDescription != nil {
					f7elem.OptionDescription = f7iter.OptionDescription
				}
				if f7iter.OptionName != nil {
					f7elem.OptionName = f7iter.OptionName
				}
				if f7iter.OptionSettings != nil {
					f7elemf3 := []*svcapitypes.OptionSetting{}
					for _, f7elemf3iter := range f7iter.OptionSettings {
						f7elemf3elem := &svcapitypes.OptionSetting{}
						if f7elemf3iter.AllowedValues != nil {
							f7elemf3elem.AllowedValues = f7elemf3iter.AllowedValues
						}
						if f7elemf3iter.ApplyType != nil {
							f7elemf3elem.ApplyType = f7elemf3iter.ApplyType
						}
						if f7elemf3iter.DataType != nil {
							f7elemf3elem.DataType = f7elemf3iter.DataType
						}
						if f7elemf3iter.DefaultValue != nil {
							f7elemf3elem.DefaultValue = f7elemf3iter.DefaultValue
						}
						if f7elemf3iter.Description != nil {
							f7elemf3elem.Description = f7elemf3iter.Description
						}
						if f7elemf3iter.IsCollection != nil {
							f7elemf3elem.IsCollection = f7elemf3iter.IsCollection
						}
						if f7elemf3iter.IsModifiable != nil {
							f7elemf3elem.IsModifiable = f7elemf3iter.IsModifiable
						}
						if f7elemf3iter.Name != nil {
							f7elemf3elem.Name = f7elemf3iter.Name
						}
						if f7elemf3iter.Value != nil {
							f7elemf3elem.Value = f7elemf3iter.Value
						}
						f7elemf3 = append(f7elemf3, f7elemf3elem)
					}
					f7elem.OptionSettings = f7elemf3
				}
				if f7iter.OptionVersion != nil {
					f7elem.OptionVersion = f7iter.OptionVersion
				}
				if f7iter.Permanent != nil {
					f7elem.Permanent = f7iter.Permanent
				}
				if f7iter.Persistent != nil {
					f7elem.Persistent = f7iter.Persistent
				}
				if f7iter.Port != nil {
					f7elem.Port = f7iter.Port
				}
				if f7iter.VpcSecurityGroupMemberships != nil {
					f7elemf8 := []*svcapitypes.VPCSecurityGroupMembership{}
					for _, f7elemf8iter := range f7iter.VpcSecurityGroupMemberships {
						f7elemf8elem := &svcapitypes.VPCSecurityGroupMembership{}
						if f7elemf8iter.Status != nil {
							f7elemf8elem.Status = f7elemf8iter.Status
						}
						if f7elemf8iter.VpcSecurityGroupId != nil {
							f7elemf8elem.VPCSecurityGroupID = f7elemf8iter.VpcSecurityGroupId
						}
						f7elemf8 = append(f7elemf8, f7elemf8elem)
					}
					f7elem.VPCSecurityGroupMemberships = f7elemf8
				}
				f7 = append(f7, f7elem)
			}
			cr.Status.AtProvider.Options = f7
		} else {
			cr.Status.AtProvider.Options = nil
		}
		if elem.SourceAccountId != nil {
			cr.Status.AtProvider.SourceAccountID = elem.SourceAccountId
		} else {
			cr.Status.AtProvider.SourceAccountID = nil
		}
		if elem.SourceOptionGroup != nil {
			cr.Status.AtProvider.SourceOptionGroup = elem.SourceOptionGroup
		} else {
			cr.Status.AtProvider.SourceOptionGroup = nil
		}
		if elem.VpcId != nil {
			cr.Status.AtProvider.VPCID = elem.VpcId
		} else {
			cr.Status.AtProvider.VPCID = nil
		}
		found = true
		break
	}
	if !found {
		return cr
	}

	return cr
}

// GenerateCreateOptionGroupInput returns a create input.
func GenerateCreateOptionGroupInput(cr *svcapitypes.OptionGroup) *svcsdk.CreateOptionGroupInput {
	res := &svcsdk.CreateOptionGroupInput{}

	if cr.Spec.ForProvider.EngineName != nil {
		res.SetEngineName(*cr.Spec.ForProvider.EngineName)
	}
	if cr.Spec.ForProvider.MajorEngineVersion != nil {
		res.SetMajorEngineVersion(*cr.Spec.ForProvider.MajorEngineVersion)
	}
	if cr.Spec.ForProvider.OptionGroupDescription != nil {
		res.SetOptionGroupDescription(*cr.Spec.ForProvider.OptionGroupDescription)
	}
	if cr.Spec.ForProvider.Tags != nil {
		f3 := []*svcsdk.Tag{}
		for _, f3iter := range cr.Spec.ForProvider.Tags {
			f3elem := &svcsdk.Tag{}
			if f3iter.Key != nil {
				f3elem.SetKey(*f3iter.Key)
			}
			if f3iter.Value != nil {
				f3elem.SetValue(*f3iter.Value)
			}
			f3 = append(f3, f3elem)
		}
		res.SetTags(f3)
	}

	return res
}

// GenerateModifyOptionGroupInput returns an update input.
func GenerateModifyOptionGroupInput(cr *svcapitypes.OptionGroup) *svcsdk.ModifyOptionGroupInput {
	res := &svcsdk.ModifyOptionGroupInput{}

	if cr.Status.AtProvider.OptionGroupName != nil {
		res.SetOptionGroupName(*cr.Status.AtProvider.OptionGroupName)
	}

	return res
}

// GenerateDeleteOptionGroupInput returns a deletion input.
func GenerateDeleteOptionGroupInput(cr *svcapitypes.OptionGroup) *svcsdk.DeleteOptionGroupInput {
	res := &svcsdk.DeleteOptionGroupInput{}

	if cr.Status.AtProvider.OptionGroupName != nil {
		res.SetOptionGroupName(*cr.Status.AtProvider.OptionGroupName)
	}

	return res
}

// IsNotFound returns whether the given error is of type NotFound or not.
func IsNotFound(err error) bool {
	awsErr, ok := err.(awserr.Error)
	return ok && awsErr.Code() == "OptionGroupNotFoundFault"
}