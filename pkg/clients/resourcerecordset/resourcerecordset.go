/*
Copyright 2019 The Crossplane Authors.

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

package resourcerecordset

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/route53"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	runtimev1alpha1 "github.com/crossplane/crossplane-runtime/apis/core/v1alpha1"

	"github.com/crossplane/provider-aws/apis/route53/v1alpha1"
	awsclients "github.com/crossplane/provider-aws/pkg/clients"
)

const (
	//RRSetNotFound is the error code that is returned if RRSet is present
	RRSetNotFound = "InvalidRRSetName.NotFound"
)

// Client defines ResourceRecordSet operations
type Client interface {
	ChangeResourceRecordSetsRequest(input *route53.ChangeResourceRecordSetsInput) route53.ChangeResourceRecordSetsRequest
	ListResourceRecordSetsRequest(input *route53.ListResourceRecordSetsInput) route53.ListResourceRecordSetsRequest
}

// RRsetNotFound will be raised when there is no ResourceRecordSet
type RRsetNotFound struct{}

// Error satisfies the Error interface for RRsetNotFound.
// We need to implement our own error for this because AWS SDK doesn't have
// a predefined error for Resource Record not found.
func (r RRsetNotFound) Error() string {
	return fmt.Sprint(RRSetNotFound)
}

// NewClient creates new AWS client with provided AWS Configuration/Credentials
func NewClient(config *aws.Config) Client {
	return route53.New(*config)
}

// GenerateChangeResourceRecordSetsInput prepares input for a ChangeResourceRecordSetsInput
func GenerateChangeResourceRecordSetsInput(p *v1alpha1.ResourceRecordSetParameters, action route53.ChangeAction) *route53.ChangeResourceRecordSetsInput { // nolint:gocyclo
	var ttl *int64

	if p.TTL == nil {
		num := int64(300)
		ttl = &num
	} else {
		ttl = p.TTL
	}

	resourceRecords := make([]route53.ResourceRecord, len(p.ResourceRecords))
	for i, r := range p.ResourceRecords {
		resourceRecords[i] = route53.ResourceRecord{
			Value: r.Value,
		}
	}

	r := &route53.ResourceRecordSet{
		Name:            p.Name,
		Type:            route53.RRType(aws.StringValue(p.Type)),
		TTL:             ttl,
		ResourceRecords: resourceRecords,
	}

	if p.AliasTarget != nil {
		r.AliasTarget = &route53.AliasTarget{
			DNSName:              awsclients.LateInitializeStringPtr(p.AliasTarget.DNSName, aws.String("")),
			EvaluateTargetHealth: awsclients.LateInitializeBoolPtr(p.AliasTarget.EvaluateTargetHealth, aws.Bool(false)),
			HostedZoneId:         awsclients.LateInitializeStringPtr(p.AliasTarget.HostedZoneID, aws.String("")),
		}
	}

	if p.GeoLocation != nil {
		r.GeoLocation = &route53.GeoLocation{
			ContinentCode:   awsclients.LateInitializeStringPtr(p.GeoLocation.ContinentCode, aws.String("")),
			CountryCode:     awsclients.LateInitializeStringPtr(p.GeoLocation.CountryCode, aws.String("")),
			SubdivisionCode: awsclients.LateInitializeStringPtr(p.GeoLocation.SubdivisionCode, aws.String("")),
		}
	}

	if p.SetIdentifier != nil {
		r.SetIdentifier = p.SetIdentifier
	}
	if p.Weight != nil {
		r.Weight = p.Weight
	}
	if p.Failover != "" {
		r.Failover = route53.ResourceRecordSetFailover(p.Failover)
	}
	if p.HealthCheckID != nil {
		r.HealthCheckId = p.HealthCheckID
	}
	if p.MultiValueAnswer != nil {
		r.MultiValueAnswer = p.MultiValueAnswer
	}
	if p.Region != "" {
		r.Region = route53.ResourceRecordSetRegion(p.Region)
	}
	if p.TrafficPolicyInstanceID != nil {
		r.TrafficPolicyInstanceId = p.TrafficPolicyInstanceID
	}

	return &route53.ChangeResourceRecordSetsInput{
		HostedZoneId: p.ZoneID,
		ChangeBatch: &route53.ChangeBatch{
			Changes: []route53.Change{
				{
					Action:            action,
					ResourceRecordSet: r,
				},
			},
		},
	}
}

// IsUpToDate checks if object is up to date
func IsUpToDate(p v1alpha1.ResourceRecordSetParameters, rrset route53.ResourceRecordSet) (bool, error) {
	// check for the root "." found in DNS entries and add if not found
	if !strings.HasSuffix(*p.Name, ".") {
		p.Name = aws.String(fmt.Sprintf("%s.", *p.Name))
	}
	patch, err := CreatePatch(&rrset, &p)
	if err != nil {
		return false, err
	}
	return cmp.Equal(&v1alpha1.ResourceRecordSetParameters{}, patch, cmpopts.IgnoreTypes(&runtimev1alpha1.Reference{}, &runtimev1alpha1.Selector{})), nil
}

// LateInitialize fills the empty fields in *v1alpha1.ResourceRecordSetParameters with
// the values seen in route53.ResourceRecordSet.
func LateInitialize(in *v1alpha1.ResourceRecordSetParameters, rrSet *route53.ResourceRecordSet) {
	if rrSet == nil {
		return
	}

	in.Name = awsclients.LateInitializeStringPtr(in.Name, rrSet.Name)

	rrType := string(rrSet.Type)
	in.Type = awsclients.LateInitializeStringPtr(in.Type, &rrType)

	in.TTL = awsclients.LateInitializeInt64Ptr(in.TTL, rrSet.TTL)
	if len(in.ResourceRecords) == 0 && len(rrSet.ResourceRecords) != 0 {
		in.ResourceRecords = make([]v1alpha1.ResourceRecord, len(rrSet.ResourceRecords))
		for i, val := range rrSet.ResourceRecords {
			in.ResourceRecords[i] = v1alpha1.ResourceRecord{
				Value: val.Value,
			}
		}
	}
}

// CreatePatch creates a *v1beta1.ResourceRecordSetParameters that has only the changed
// values between the target *v1beta1.ResourceRecordSetParameters and the current
// *route53.ResourceRecordSet
func CreatePatch(in *route53.ResourceRecordSet, target *v1alpha1.ResourceRecordSetParameters) (*v1alpha1.ResourceRecordSetParameters, error) {
	currentParams := &v1alpha1.ResourceRecordSetParameters{}
	LateInitialize(currentParams, in)

	jsonPatch, err := awsclients.CreateJSONPatch(currentParams, target)

	if err != nil {
		return nil, err
	}
	patch := &v1alpha1.ResourceRecordSetParameters{}
	if err := json.Unmarshal(jsonPatch, patch); err != nil {
		return nil, err
	}
	patch.ZoneID = patch.Name
	return patch, nil
}

// GetResourceRecordSet returns recordSet if present or err
func GetResourceRecordSet(ctx context.Context, c Client, name, zoneID, rrType, si *string) (route53.ResourceRecordSet, error) {
	// check for the root "." found in DNS entries and add if not found
	if !strings.HasSuffix(*name, ".") {
		name = aws.String(fmt.Sprintf("%s.", *name))
	}
	res, err := c.ListResourceRecordSetsRequest(&route53.ListResourceRecordSetsInput{
		HostedZoneId: zoneID,
	}).Send(ctx)
	if err != nil {
		return route53.ResourceRecordSet{}, err
	}

	for _, rrSet := range res.ResourceRecordSets {
		if aws.StringValue(rrSet.Name) == aws.StringValue(name) && string(rrSet.Type) == aws.StringValue(rrType) && aws.StringValue(rrSet.SetIdentifier) == aws.StringValue(si) {
			return rrSet, nil
		}
	}

	return route53.ResourceRecordSet{}, RRsetNotFound{}
}

// IsErrorRRsetNotFound returns true if the error code indicates that the requested Resource Record was not found
func IsErrorRRsetNotFound(err error) bool {
	if rrErr, ok := err.(RRsetNotFound); ok && rrErr.Error() == RRSetNotFound {
		return true
	}
	return false
}
