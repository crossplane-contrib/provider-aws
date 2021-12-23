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

package volume

import (
	"github.com/aws/aws-sdk-go/aws/awserr"
	svcsdk "github.com/aws/aws-sdk-go/service/ec2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	svcapitypes "github.com/crossplane/provider-aws/apis/ec2/v1alpha1"
)

// NOTE(muvaf): We return pointers in case the function needs to start with an
// empty object, hence need to return a new pointer.

// GenerateDescribeVolumesInput returns input for read
// operation.
func GenerateDescribeVolumesInput(cr *svcapitypes.Volume) *svcsdk.DescribeVolumesInput {
	res := &svcsdk.DescribeVolumesInput{}

	if cr.Status.AtProvider.VolumeID != nil {
		f4 := []*string{}
		f4 = append(f4, cr.Status.AtProvider.VolumeID)
		res.SetVolumeIds(f4)
	}

	return res
}

// GenerateVolume returns the current state in the form of *svcapitypes.Volume.
func GenerateVolume(resp *svcsdk.DescribeVolumesOutput) *svcapitypes.Volume {
	cr := &svcapitypes.Volume{}

	found := false
	for _, elem := range resp.Volumes {
		if elem.Attachments != nil {
			f0 := []*svcapitypes.VolumeAttachment{}
			for _, f0iter := range elem.Attachments {
				f0elem := &svcapitypes.VolumeAttachment{}
				if f0iter.AttachTime != nil {
					f0elem.AttachTime = &metav1.Time{*f0iter.AttachTime}
				}
				if f0iter.DeleteOnTermination != nil {
					f0elem.DeleteOnTermination = f0iter.DeleteOnTermination
				}
				if f0iter.Device != nil {
					f0elem.Device = f0iter.Device
				}
				if f0iter.InstanceId != nil {
					f0elem.InstanceID = f0iter.InstanceId
				}
				if f0iter.State != nil {
					f0elem.State = f0iter.State
				}
				if f0iter.VolumeId != nil {
					f0elem.VolumeID = f0iter.VolumeId
				}
				f0 = append(f0, f0elem)
			}
			cr.Status.AtProvider.Attachments = f0
		} else {
			cr.Status.AtProvider.Attachments = nil
		}
		if elem.AvailabilityZone != nil {
			cr.Spec.ForProvider.AvailabilityZone = elem.AvailabilityZone
		} else {
			cr.Spec.ForProvider.AvailabilityZone = nil
		}
		if elem.CreateTime != nil {
			cr.Status.AtProvider.CreateTime = &metav1.Time{*elem.CreateTime}
		} else {
			cr.Status.AtProvider.CreateTime = nil
		}
		if elem.Encrypted != nil {
			cr.Spec.ForProvider.Encrypted = elem.Encrypted
		} else {
			cr.Spec.ForProvider.Encrypted = nil
		}
		if elem.FastRestored != nil {
			cr.Status.AtProvider.FastRestored = elem.FastRestored
		} else {
			cr.Status.AtProvider.FastRestored = nil
		}
		if elem.Iops != nil {
			cr.Spec.ForProvider.IOPS = elem.Iops
		} else {
			cr.Spec.ForProvider.IOPS = nil
		}
		if elem.KmsKeyId != nil {
			cr.Status.AtProvider.KMSKeyID = elem.KmsKeyId
		} else {
			cr.Status.AtProvider.KMSKeyID = nil
		}
		if elem.MultiAttachEnabled != nil {
			cr.Spec.ForProvider.MultiAttachEnabled = elem.MultiAttachEnabled
		} else {
			cr.Spec.ForProvider.MultiAttachEnabled = nil
		}
		if elem.OutpostArn != nil {
			cr.Spec.ForProvider.OutpostARN = elem.OutpostArn
		} else {
			cr.Spec.ForProvider.OutpostARN = nil
		}
		if elem.Size != nil {
			cr.Spec.ForProvider.Size = elem.Size
		} else {
			cr.Spec.ForProvider.Size = nil
		}
		if elem.SnapshotId != nil {
			cr.Spec.ForProvider.SnapshotID = elem.SnapshotId
		} else {
			cr.Spec.ForProvider.SnapshotID = nil
		}
		if elem.State != nil {
			cr.Status.AtProvider.State = elem.State
		} else {
			cr.Status.AtProvider.State = nil
		}
		if elem.Tags != nil {
			f12 := []*svcapitypes.Tag{}
			for _, f12iter := range elem.Tags {
				f12elem := &svcapitypes.Tag{}
				if f12iter.Key != nil {
					f12elem.Key = f12iter.Key
				}
				if f12iter.Value != nil {
					f12elem.Value = f12iter.Value
				}
				f12 = append(f12, f12elem)
			}
			cr.Status.AtProvider.Tags = f12
		} else {
			cr.Status.AtProvider.Tags = nil
		}
		if elem.Throughput != nil {
			cr.Spec.ForProvider.Throughput = elem.Throughput
		} else {
			cr.Spec.ForProvider.Throughput = nil
		}
		if elem.VolumeId != nil {
			cr.Status.AtProvider.VolumeID = elem.VolumeId
		} else {
			cr.Status.AtProvider.VolumeID = nil
		}
		if elem.VolumeType != nil {
			cr.Spec.ForProvider.VolumeType = elem.VolumeType
		} else {
			cr.Spec.ForProvider.VolumeType = nil
		}
		found = true
		break
	}
	if !found {
		return cr
	}

	return cr
}

// GenerateCreateVolumeInput returns a create input.
func GenerateCreateVolumeInput(cr *svcapitypes.Volume) *svcsdk.CreateVolumeInput {
	res := &svcsdk.CreateVolumeInput{}

	if cr.Spec.ForProvider.AvailabilityZone != nil {
		res.SetAvailabilityZone(*cr.Spec.ForProvider.AvailabilityZone)
	}
	if cr.Spec.ForProvider.Encrypted != nil {
		res.SetEncrypted(*cr.Spec.ForProvider.Encrypted)
	}
	if cr.Spec.ForProvider.IOPS != nil {
		res.SetIops(*cr.Spec.ForProvider.IOPS)
	}
	if cr.Spec.ForProvider.MultiAttachEnabled != nil {
		res.SetMultiAttachEnabled(*cr.Spec.ForProvider.MultiAttachEnabled)
	}
	if cr.Spec.ForProvider.OutpostARN != nil {
		res.SetOutpostArn(*cr.Spec.ForProvider.OutpostARN)
	}
	if cr.Spec.ForProvider.Size != nil {
		res.SetSize(*cr.Spec.ForProvider.Size)
	}
	if cr.Spec.ForProvider.SnapshotID != nil {
		res.SetSnapshotId(*cr.Spec.ForProvider.SnapshotID)
	}
	if cr.Spec.ForProvider.TagSpecifications != nil {
		f7 := []*svcsdk.TagSpecification{}
		for _, f7iter := range cr.Spec.ForProvider.TagSpecifications {
			f7elem := &svcsdk.TagSpecification{}
			if f7iter.ResourceType != nil {
				f7elem.SetResourceType(*f7iter.ResourceType)
			}
			if f7iter.Tags != nil {
				f7elemf1 := []*svcsdk.Tag{}
				for _, f7elemf1iter := range f7iter.Tags {
					f7elemf1elem := &svcsdk.Tag{}
					if f7elemf1iter.Key != nil {
						f7elemf1elem.SetKey(*f7elemf1iter.Key)
					}
					if f7elemf1iter.Value != nil {
						f7elemf1elem.SetValue(*f7elemf1iter.Value)
					}
					f7elemf1 = append(f7elemf1, f7elemf1elem)
				}
				f7elem.SetTags(f7elemf1)
			}
			f7 = append(f7, f7elem)
		}
		res.SetTagSpecifications(f7)
	}
	if cr.Spec.ForProvider.Throughput != nil {
		res.SetThroughput(*cr.Spec.ForProvider.Throughput)
	}
	if cr.Spec.ForProvider.VolumeType != nil {
		res.SetVolumeType(*cr.Spec.ForProvider.VolumeType)
	}

	return res
}

// GenerateModifyVolumeInput returns an update input.
func GenerateModifyVolumeInput(cr *svcapitypes.Volume) *svcsdk.ModifyVolumeInput {
	res := &svcsdk.ModifyVolumeInput{}

	if cr.Spec.ForProvider.IOPS != nil {
		res.SetIops(*cr.Spec.ForProvider.IOPS)
	}
	if cr.Spec.ForProvider.MultiAttachEnabled != nil {
		res.SetMultiAttachEnabled(*cr.Spec.ForProvider.MultiAttachEnabled)
	}
	if cr.Spec.ForProvider.Size != nil {
		res.SetSize(*cr.Spec.ForProvider.Size)
	}
	if cr.Spec.ForProvider.Throughput != nil {
		res.SetThroughput(*cr.Spec.ForProvider.Throughput)
	}
	if cr.Status.AtProvider.VolumeID != nil {
		res.SetVolumeId(*cr.Status.AtProvider.VolumeID)
	}
	if cr.Spec.ForProvider.VolumeType != nil {
		res.SetVolumeType(*cr.Spec.ForProvider.VolumeType)
	}

	return res
}

// GenerateDeleteVolumeInput returns a deletion input.
func GenerateDeleteVolumeInput(cr *svcapitypes.Volume) *svcsdk.DeleteVolumeInput {
	res := &svcsdk.DeleteVolumeInput{}

	if cr.Status.AtProvider.VolumeID != nil {
		res.SetVolumeId(*cr.Status.AtProvider.VolumeID)
	}

	return res
}

// IsNotFound returns whether the given error is of type NotFound or not.
func IsNotFound(err error) bool {
	awsErr, ok := err.(awserr.Error)
	return ok && awsErr.Code() == "InvalidVolume.NotFound"
}
