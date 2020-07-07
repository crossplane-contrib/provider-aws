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

package hostedzone

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/awserr"
	"github.com/aws/aws-sdk-go-v2/service/route53"

	"github.com/crossplane/provider-aws/apis/route53/v1alpha1"
	awsclients "github.com/crossplane/provider-aws/pkg/clients"
)

// Client defines Route53 Client operations
type Client interface {
	CreateHostedZoneRequest(input *route53.CreateHostedZoneInput) route53.CreateHostedZoneRequest
	DeleteHostedZoneRequest(input *route53.DeleteHostedZoneInput) route53.DeleteHostedZoneRequest
	GetHostedZoneRequest(input *route53.GetHostedZoneInput) route53.GetHostedZoneRequest
	UpdateHostedZoneCommentRequest(input *route53.UpdateHostedZoneCommentInput) route53.UpdateHostedZoneCommentRequest
}

// NewClient creates new AWS Client with provided AWS Configurations/Credentials
func NewClient(config *aws.Config) Client {
	return route53.New(*config)
}

// IsErrorNoSuchHostedZone returns true if the error code indicates that the requested Zone was not found
func IsErrorNoSuchHostedZone(err error) bool {
	if zoneErr, ok := err.(awserr.Error); ok && zoneErr.Code() == route53.ErrCodeNoSuchHostedZone {
		return true
	}
	return false
}

// IsUpToDate check whether the comment in Spec and Response are same or not
func IsUpToDate(cfg *v1alpha1.HostedZoneConfig, res *route53.HostedZoneConfig) bool {
	if cfg == nil && aws.StringValue(res.Comment) != "" {
		return false
	}
	if cfg != nil && !(aws.StringValue(cfg.Comment) == aws.StringValue(res.Comment)) {
		return false
	}
	return true
}

// LateInitialize fills the empty fields in *v1alpha1.HostedZoneParameters with
// the values seen in route53.HostedZone.
func LateInitialize(in *v1alpha1.HostedZoneParameters, hz *route53.GetHostedZoneResponse) {
	in.DelegationSetID = awsclients.LateInitializeStringPtr(in.DelegationSetID, hz.DelegationSet.Id)
	if in.HostedZoneConfig != nil {
		in.HostedZoneConfig = &v1alpha1.HostedZoneConfig{
			Comment:     in.HostedZoneConfig.Comment,
			PrivateZone: awsclients.LateInitializeBoolPtr(in.HostedZoneConfig.PrivateZone, hz.HostedZone.Config.PrivateZone),
		}
	}
}

// GenerateCreateHostedZoneInput returns a route53 CreateHostedZoneInput using which a route53
// Hosted Zone can be created.
func GenerateCreateHostedZoneInput(cr *v1alpha1.HostedZone) *route53.CreateHostedZoneInput {
	reqInput := &route53.CreateHostedZoneInput{
		CallerReference: aws.String(string(cr.ObjectMeta.UID)),
		Name:            cr.Spec.ForProvider.Name,
		DelegationSetId: cr.Spec.ForProvider.DelegationSetID,
	}

	if cr.Spec.ForProvider.HostedZoneConfig != nil {
		reqInput.HostedZoneConfig = &route53.HostedZoneConfig{
			PrivateZone: cr.Spec.ForProvider.HostedZoneConfig.PrivateZone,
			Comment:     cr.Spec.ForProvider.HostedZoneConfig.Comment,
		}

		if *cr.Spec.ForProvider.HostedZoneConfig.PrivateZone {

			if cr.Spec.ForProvider.VPCId != nil {
				reqInput.VPC = &route53.VPC{
					VPCId: cr.Spec.ForProvider.VPCId,
				}
			}

			if cr.Spec.ForProvider.VPCRegion != nil {
				reqInput.VPC.VPCRegion = route53.VPCRegion(aws.StringValue(cr.Spec.ForProvider.VPCRegion))
			}
		}
	}

	return reqInput
}

// GenerateObservation generates and returns v1alpha1.HostedZoneObservation which can be used as the status of the runtime object
func GenerateObservation(op *route53.GetHostedZoneResponse) v1alpha1.HostedZoneObservation {

	o := v1alpha1.HostedZoneObservation{}
	if op.DelegationSet != nil {
		n := make([]string, len(op.DelegationSet.NameServers))
		copy(n, op.DelegationSet.NameServers)
		o.DelegationSet = v1alpha1.DelegationSet{
			CallerReference: aws.StringValue(op.DelegationSet.CallerReference),
			ID:              aws.StringValue(op.DelegationSet.Id),
			NameServers:     n,
		}
	}
	if op.HostedZone != nil {
		o.HostedZone = v1alpha1.HostedZoneResponse{
			CallerReference:        aws.StringValue(op.HostedZone.CallerReference),
			ID:                     aws.StringValue(op.HostedZone.Id),
			ResourceRecordSetCount: aws.Int64Value(op.HostedZone.ResourceRecordSetCount),
		}

		if op.HostedZone.LinkedService != nil {
			o.HostedZone.LinkedService = v1alpha1.LinkedService{
				Description:      aws.StringValue(op.HostedZone.LinkedService.Description),
				ServicePrincipal: aws.StringValue(op.HostedZone.LinkedService.ServicePrincipal),
			}
		}
	}
	return o
}

// GenerateUpdateHostedZoneCommentInput returns a route53 UpdateHostedZoneCommentInput using which a route53
// Hosted Zone comment can be updated.
func GenerateUpdateHostedZoneCommentInput(cfg *v1alpha1.HostedZoneConfig, id *string) *route53.UpdateHostedZoneCommentInput {
	var c *string
	c = aws.String("")
	if cfg != nil && cfg.Comment != nil {
		c = cfg.Comment
	}
	return &route53.UpdateHostedZoneCommentInput{
		Comment: c,
		Id:      id,
	}
}
