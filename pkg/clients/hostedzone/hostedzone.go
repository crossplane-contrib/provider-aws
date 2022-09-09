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
	"context"
	"errors"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/route53"
	route53types "github.com/aws/aws-sdk-go-v2/service/route53/types"

	"github.com/crossplane-contrib/provider-aws/apis/route53/v1alpha1"
	awsclients "github.com/crossplane-contrib/provider-aws/pkg/clients"
)

// IDPrefix is the prefix of the actual ID that's returned from GET call.
const IDPrefix = "/hostedzone/"

// Client defines Route53 Client operations
type Client interface {
	CreateHostedZone(ctx context.Context, input *route53.CreateHostedZoneInput, opts ...func(*route53.Options)) (*route53.CreateHostedZoneOutput, error)
	DeleteHostedZone(ctx context.Context, input *route53.DeleteHostedZoneInput, opts ...func(*route53.Options)) (*route53.DeleteHostedZoneOutput, error)
	GetHostedZone(ctx context.Context, input *route53.GetHostedZoneInput, opts ...func(*route53.Options)) (*route53.GetHostedZoneOutput, error)
	UpdateHostedZoneComment(ctx context.Context, input *route53.UpdateHostedZoneCommentInput, opts ...func(*route53.Options)) (*route53.UpdateHostedZoneCommentOutput, error)
}

// NewClient creates new RDS RDSClient with provided AWS Configurations/Credentials
func NewClient(cfg aws.Config) Client {
	return route53.NewFromConfig(cfg)
}

// IsNotFound returns true if the error code indicates that the requested Zone was not found
func IsNotFound(err error) bool {
	var nshz *route53types.NoSuchHostedZone
	return errors.As(err, &nshz)
}

// IsUpToDate check whether the comment in Spec and Response are same or not
func IsUpToDate(spec v1alpha1.HostedZoneParameters, obs route53types.HostedZone) bool {
	s := ""
	if spec.Config != nil {
		s = awsclients.StringValue(spec.Config.Comment)
	}
	o := ""
	if obs.Config != nil {
		o = awsclients.StringValue(obs.Config.Comment)
	}
	return s == o
}

// LateInitialize fills the empty fields in *v1alpha1.HostedZoneParameters with
// the values seen in route53types.HostedZone.
func LateInitialize(spec *v1alpha1.HostedZoneParameters, obs *route53.GetHostedZoneOutput) {
	if obs == nil || obs.HostedZone == nil {
		return
	}
	if obs.DelegationSet != nil {
		spec.DelegationSetID = awsclients.LateInitializeStringPtr(spec.DelegationSetID, obs.DelegationSet.Id)
	}
	if spec.Config == nil && obs.HostedZone != nil {
		spec.Config = &v1alpha1.Config{}
	}
	if spec.Config != nil && obs.HostedZone.Config != nil {
		spec.Config.Comment = awsclients.LateInitializeStringPtr(spec.Config.Comment, obs.HostedZone.Config.Comment)
		spec.Config.PrivateZone = awsclients.LateInitializeBoolPtr(spec.Config.PrivateZone, &obs.HostedZone.Config.PrivateZone)
	}
}

// GenerateCreateHostedZoneInput returns a route53 CreateHostedZoneInput using which a route53
// Hosted Zone can be created.
func GenerateCreateHostedZoneInput(cr *v1alpha1.HostedZone) *route53.CreateHostedZoneInput {
	reqInput := &route53.CreateHostedZoneInput{
		CallerReference: aws.String(cr.ObjectMeta.ResourceVersion),
		Name:            aws.String(cr.Spec.ForProvider.Name),
		DelegationSetId: cr.Spec.ForProvider.DelegationSetID,
	}
	if cr.Spec.ForProvider.Config != nil {
		reqInput.HostedZoneConfig = &route53types.HostedZoneConfig{
			PrivateZone: aws.ToBool(cr.Spec.ForProvider.Config.PrivateZone),
			Comment:     cr.Spec.ForProvider.Config.Comment,
		}
	}
	if cr.Spec.ForProvider.VPC != nil {
		reqInput.VPC = &route53types.VPC{VPCId: cr.Spec.ForProvider.VPC.VPCID, VPCRegion: route53types.VPCRegion(awsclients.StringValue(cr.Spec.ForProvider.VPC.VPCRegion))}
	}
	return reqInput
}

// GenerateObservation generates and returns v1alpha1.HostedZoneObservation which can be used as the status of the runtime object
func GenerateObservation(op *route53.GetHostedZoneOutput) v1alpha1.HostedZoneObservation {
	o := v1alpha1.HostedZoneObservation{}
	if op.DelegationSet != nil {
		n := make([]string, len(op.DelegationSet.NameServers))
		copy(n, op.DelegationSet.NameServers)
		o.DelegationSet = v1alpha1.DelegationSet{
			CallerReference: aws.ToString(op.DelegationSet.CallerReference),
			ID:              aws.ToString(op.DelegationSet.Id),
			NameServers:     n,
		}
	}
	if op.HostedZone != nil {
		o.HostedZone = v1alpha1.HostedZoneResponse{
			CallerReference:        aws.ToString(op.HostedZone.CallerReference),
			ID:                     aws.ToString(op.HostedZone.Id),
			ResourceRecordSetCount: aws.ToInt64(op.HostedZone.ResourceRecordSetCount),
		}

		if op.HostedZone.LinkedService != nil {
			o.HostedZone.LinkedService = v1alpha1.LinkedService{
				Description:      aws.ToString(op.HostedZone.LinkedService.Description),
				ServicePrincipal: aws.ToString(op.HostedZone.LinkedService.ServicePrincipal),
			}
		}
	}
	for _, vpc := range op.VPCs {
		o.VPCs = append(o.VPCs, v1alpha1.VPCObservation{VPCID: awsclients.StringValue(vpc.VPCId), VPCRegion: string(vpc.VPCRegion)})
	}
	return o
}

// GenerateUpdateHostedZoneCommentInput returns a route53 UpdateHostedZoneCommentInput using which a route53
// Hosted Zone comment can be updated.
func GenerateUpdateHostedZoneCommentInput(spec v1alpha1.HostedZoneParameters, id string) *route53.UpdateHostedZoneCommentInput {
	comment := ""
	if spec.Config != nil && spec.Config.Comment != nil {
		comment = *spec.Config.Comment
	}
	return &route53.UpdateHostedZoneCommentInput{
		Comment: &comment,
		Id:      &id,
	}
}
