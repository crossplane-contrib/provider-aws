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

package capacityreservation

import (
	"github.com/aws/aws-sdk-go/aws/awserr"
	svcsdk "github.com/aws/aws-sdk-go/service/ec2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	svcapitypes "github.com/crossplane-contrib/provider-aws/apis/ec2/v1alpha1"
)

// NOTE(muvaf): We return pointers in case the function needs to start with an
// empty object, hence need to return a new pointer.

// GenerateDescribeCapacityReservationsInput returns input for read
// operation.
func GenerateDescribeCapacityReservationsInput(cr *svcapitypes.CapacityReservation) *svcsdk.DescribeCapacityReservationsInput {
	res := &svcsdk.DescribeCapacityReservationsInput{}

	if cr.Status.AtProvider.CapacityReservationID != nil {
		f0 := []*string{}
		f0 = append(f0, cr.Status.AtProvider.CapacityReservationID)
		res.SetCapacityReservationIds(f0)
	}

	return res
}

// GenerateCapacityReservation returns the current state in the form of *svcapitypes.CapacityReservation.
func GenerateCapacityReservation(resp *svcsdk.DescribeCapacityReservationsOutput) *svcapitypes.CapacityReservation {
	cr := &svcapitypes.CapacityReservation{}

	found := false
	for _, elem := range resp.CapacityReservations {
		if elem.AvailabilityZone != nil {
			cr.Spec.ForProvider.AvailabilityZone = elem.AvailabilityZone
		} else {
			cr.Spec.ForProvider.AvailabilityZone = nil
		}
		if elem.AvailabilityZoneId != nil {
			cr.Spec.ForProvider.AvailabilityZoneID = elem.AvailabilityZoneId
		} else {
			cr.Spec.ForProvider.AvailabilityZoneID = nil
		}
		if elem.AvailableInstanceCount != nil {
			cr.Status.AtProvider.AvailableInstanceCount = elem.AvailableInstanceCount
		} else {
			cr.Status.AtProvider.AvailableInstanceCount = nil
		}
		if elem.CapacityAllocations != nil {
			f3 := []*svcapitypes.CapacityAllocation{}
			for _, f3iter := range elem.CapacityAllocations {
				f3elem := &svcapitypes.CapacityAllocation{}
				if f3iter.AllocationType != nil {
					f3elem.AllocationType = f3iter.AllocationType
				}
				if f3iter.Count != nil {
					f3elem.Count = f3iter.Count
				}
				f3 = append(f3, f3elem)
			}
			cr.Status.AtProvider.CapacityAllocations = f3
		} else {
			cr.Status.AtProvider.CapacityAllocations = nil
		}
		if elem.CapacityReservationArn != nil {
			cr.Status.AtProvider.CapacityReservationARN = elem.CapacityReservationArn
		} else {
			cr.Status.AtProvider.CapacityReservationARN = nil
		}
		if elem.CapacityReservationFleetId != nil {
			cr.Status.AtProvider.CapacityReservationFleetID = elem.CapacityReservationFleetId
		} else {
			cr.Status.AtProvider.CapacityReservationFleetID = nil
		}
		if elem.CapacityReservationId != nil {
			cr.Status.AtProvider.CapacityReservationID = elem.CapacityReservationId
		} else {
			cr.Status.AtProvider.CapacityReservationID = nil
		}
		if elem.CreateDate != nil {
			cr.Status.AtProvider.CreateDate = &metav1.Time{*elem.CreateDate}
		} else {
			cr.Status.AtProvider.CreateDate = nil
		}
		if elem.EbsOptimized != nil {
			cr.Spec.ForProvider.EBSOptimized = elem.EbsOptimized
		} else {
			cr.Spec.ForProvider.EBSOptimized = nil
		}
		if elem.EndDate != nil {
			cr.Spec.ForProvider.EndDate = &metav1.Time{*elem.EndDate}
		} else {
			cr.Spec.ForProvider.EndDate = nil
		}
		if elem.EndDateType != nil {
			cr.Spec.ForProvider.EndDateType = elem.EndDateType
		} else {
			cr.Spec.ForProvider.EndDateType = nil
		}
		if elem.EphemeralStorage != nil {
			cr.Spec.ForProvider.EphemeralStorage = elem.EphemeralStorage
		} else {
			cr.Spec.ForProvider.EphemeralStorage = nil
		}
		if elem.InstanceMatchCriteria != nil {
			cr.Spec.ForProvider.InstanceMatchCriteria = elem.InstanceMatchCriteria
		} else {
			cr.Spec.ForProvider.InstanceMatchCriteria = nil
		}
		if elem.InstancePlatform != nil {
			cr.Spec.ForProvider.InstancePlatform = elem.InstancePlatform
		} else {
			cr.Spec.ForProvider.InstancePlatform = nil
		}
		if elem.InstanceType != nil {
			cr.Spec.ForProvider.InstanceType = elem.InstanceType
		} else {
			cr.Spec.ForProvider.InstanceType = nil
		}
		if elem.OutpostArn != nil {
			cr.Spec.ForProvider.OutpostARN = elem.OutpostArn
		} else {
			cr.Spec.ForProvider.OutpostARN = nil
		}
		if elem.OwnerId != nil {
			cr.Status.AtProvider.OwnerID = elem.OwnerId
		} else {
			cr.Status.AtProvider.OwnerID = nil
		}
		if elem.PlacementGroupArn != nil {
			cr.Spec.ForProvider.PlacementGroupARN = elem.PlacementGroupArn
		} else {
			cr.Spec.ForProvider.PlacementGroupARN = nil
		}
		if elem.StartDate != nil {
			cr.Status.AtProvider.StartDate = &metav1.Time{*elem.StartDate}
		} else {
			cr.Status.AtProvider.StartDate = nil
		}
		if elem.State != nil {
			cr.Status.AtProvider.State = elem.State
		} else {
			cr.Status.AtProvider.State = nil
		}
		if elem.Tags != nil {
			f20 := []*svcapitypes.Tag{}
			for _, f20iter := range elem.Tags {
				f20elem := &svcapitypes.Tag{}
				if f20iter.Key != nil {
					f20elem.Key = f20iter.Key
				}
				if f20iter.Value != nil {
					f20elem.Value = f20iter.Value
				}
				f20 = append(f20, f20elem)
			}
			cr.Status.AtProvider.Tags = f20
		} else {
			cr.Status.AtProvider.Tags = nil
		}
		if elem.Tenancy != nil {
			cr.Spec.ForProvider.Tenancy = elem.Tenancy
		} else {
			cr.Spec.ForProvider.Tenancy = nil
		}
		if elem.TotalInstanceCount != nil {
			cr.Status.AtProvider.TotalInstanceCount = elem.TotalInstanceCount
		} else {
			cr.Status.AtProvider.TotalInstanceCount = nil
		}
		found = true
		break
	}
	if !found {
		return cr
	}

	return cr
}

// GenerateCreateCapacityReservationInput returns a create input.
func GenerateCreateCapacityReservationInput(cr *svcapitypes.CapacityReservation) *svcsdk.CreateCapacityReservationInput {
	res := &svcsdk.CreateCapacityReservationInput{}

	if cr.Spec.ForProvider.AvailabilityZone != nil {
		res.SetAvailabilityZone(*cr.Spec.ForProvider.AvailabilityZone)
	}
	if cr.Spec.ForProvider.AvailabilityZoneID != nil {
		res.SetAvailabilityZoneId(*cr.Spec.ForProvider.AvailabilityZoneID)
	}
	if cr.Spec.ForProvider.EBSOptimized != nil {
		res.SetEbsOptimized(*cr.Spec.ForProvider.EBSOptimized)
	}
	if cr.Spec.ForProvider.EndDate != nil {
		res.SetEndDate(cr.Spec.ForProvider.EndDate.Time)
	}
	if cr.Spec.ForProvider.EndDateType != nil {
		res.SetEndDateType(*cr.Spec.ForProvider.EndDateType)
	}
	if cr.Spec.ForProvider.EphemeralStorage != nil {
		res.SetEphemeralStorage(*cr.Spec.ForProvider.EphemeralStorage)
	}
	if cr.Spec.ForProvider.InstanceCount != nil {
		res.SetInstanceCount(*cr.Spec.ForProvider.InstanceCount)
	}
	if cr.Spec.ForProvider.InstanceMatchCriteria != nil {
		res.SetInstanceMatchCriteria(*cr.Spec.ForProvider.InstanceMatchCriteria)
	}
	if cr.Spec.ForProvider.InstancePlatform != nil {
		res.SetInstancePlatform(*cr.Spec.ForProvider.InstancePlatform)
	}
	if cr.Spec.ForProvider.InstanceType != nil {
		res.SetInstanceType(*cr.Spec.ForProvider.InstanceType)
	}
	if cr.Spec.ForProvider.OutpostARN != nil {
		res.SetOutpostArn(*cr.Spec.ForProvider.OutpostARN)
	}
	if cr.Spec.ForProvider.PlacementGroupARN != nil {
		res.SetPlacementGroupArn(*cr.Spec.ForProvider.PlacementGroupARN)
	}
	if cr.Spec.ForProvider.TagSpecifications != nil {
		f12 := []*svcsdk.TagSpecification{}
		for _, f12iter := range cr.Spec.ForProvider.TagSpecifications {
			f12elem := &svcsdk.TagSpecification{}
			if f12iter.ResourceType != nil {
				f12elem.SetResourceType(*f12iter.ResourceType)
			}
			if f12iter.Tags != nil {
				f12elemf1 := []*svcsdk.Tag{}
				for _, f12elemf1iter := range f12iter.Tags {
					f12elemf1elem := &svcsdk.Tag{}
					if f12elemf1iter.Key != nil {
						f12elemf1elem.SetKey(*f12elemf1iter.Key)
					}
					if f12elemf1iter.Value != nil {
						f12elemf1elem.SetValue(*f12elemf1iter.Value)
					}
					f12elemf1 = append(f12elemf1, f12elemf1elem)
				}
				f12elem.SetTags(f12elemf1)
			}
			f12 = append(f12, f12elem)
		}
		res.SetTagSpecifications(f12)
	}
	if cr.Spec.ForProvider.Tenancy != nil {
		res.SetTenancy(*cr.Spec.ForProvider.Tenancy)
	}

	return res
}

// GenerateModifyCapacityReservationInput returns an update input.
func GenerateModifyCapacityReservationInput(cr *svcapitypes.CapacityReservation) *svcsdk.ModifyCapacityReservationInput {
	res := &svcsdk.ModifyCapacityReservationInput{}

	if cr.Status.AtProvider.CapacityReservationID != nil {
		res.SetCapacityReservationId(*cr.Status.AtProvider.CapacityReservationID)
	}
	if cr.Spec.ForProvider.EndDate != nil {
		res.SetEndDate(cr.Spec.ForProvider.EndDate.Time)
	}
	if cr.Spec.ForProvider.EndDateType != nil {
		res.SetEndDateType(*cr.Spec.ForProvider.EndDateType)
	}
	if cr.Spec.ForProvider.InstanceCount != nil {
		res.SetInstanceCount(*cr.Spec.ForProvider.InstanceCount)
	}

	return res
}

// GenerateDeleteCapacityReservationInput returns a deletion input.
func GenerateCancelCapacityReservationInput(cr *svcapitypes.CapacityReservation) *svcsdk.CancelCapacityReservationInput {
	res := &svcsdk.CancelCapacityReservationInput{}

	if cr.Status.AtProvider.CapacityReservationID != nil {
		res.SetCapacityReservationId(*cr.Status.AtProvider.CapacityReservationID)
	}

	return res
}

// IsNotFound returns whether the given error is of type NotFound or not.
func IsNotFound(err error) bool {
	awsErr, ok := err.(awserr.Error)
	return ok && awsErr.Code() == "UNKNOWN"
}
