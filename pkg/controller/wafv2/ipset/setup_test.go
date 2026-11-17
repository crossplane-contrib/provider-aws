/*
Copyright 2026 The Crossplane Authors.

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

package ipset

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	svcsdk "github.com/aws/aws-sdk-go/service/wafv2"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/test"
	"github.com/google/go-cmp/cmp"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	svcapitypes "github.com/crossplane-contrib/provider-aws/apis/wafv2/v1alpha1"
	"github.com/crossplane-contrib/provider-aws/pkg/clients/wafv2/fake"
)

func TestObserve(t *testing.T) {
	ipSetName := "test-ipset"
	ipSetID := "abc-123"
	lockToken := "lock-token-1"
	scope := "REGIONAL"

	type args struct {
		cr     *svcapitypes.IPSet
		client *fake.MockWAFV2Client
	}
	type want struct {
		cr  *svcapitypes.IPSet
		obs managed.ExternalObservation
		err error
	}

	cases := map[string]struct {
		args args
		want want
	}{
		"ResourceNotFound": {
			args: args{
				cr: &svcapitypes.IPSet{
					ObjectMeta: metav1.ObjectMeta{
						Annotations: map[string]string{meta.AnnotationKeyExternalName: ipSetName},
					},
					Spec: svcapitypes.IPSetSpec{
						ForProvider: svcapitypes.IPSetParameters{
							Scope: aws.String(scope),
						},
					},
				},
				client: &fake.MockWAFV2Client{
					MockListIPSetsWithContext: func(input *svcsdk.ListIPSetsInput) (*svcsdk.ListIPSetsOutput, error) {
						return &svcsdk.ListIPSetsOutput{IPSets: []*svcsdk.IPSetSummary{}}, nil
					},
				},
			},
			want: want{
				obs: managed.ExternalObservation{ResourceExists: false},
			},
		},
		"ResourceExists_UpToDate": {
			args: args{
				cr: &svcapitypes.IPSet{
					ObjectMeta: metav1.ObjectMeta{
						Annotations: map[string]string{meta.AnnotationKeyExternalName: ipSetName},
					},
					Spec: svcapitypes.IPSetSpec{
						ForProvider: svcapitypes.IPSetParameters{
							Scope:       aws.String(scope),
							Addresses:   []*string{aws.String("10.0.0.0/8")},
							Description: aws.String("test"),
						},
					},
				},
				client: &fake.MockWAFV2Client{
					MockListIPSetsWithContext: func(input *svcsdk.ListIPSetsInput) (*svcsdk.ListIPSetsOutput, error) {
						return &svcsdk.ListIPSetsOutput{
							IPSets: []*svcsdk.IPSetSummary{
								{Name: aws.String(ipSetName), Id: aws.String(ipSetID)},
							},
						}, nil
					},
					MockGetIPSetWithContext: func(input *svcsdk.GetIPSetInput) (*svcsdk.GetIPSetOutput, error) {
						return &svcsdk.GetIPSetOutput{
							IPSet: &svcsdk.IPSet{
								Name:        aws.String(ipSetName),
								Id:          aws.String(ipSetID),
								ARN:         aws.String("arn:aws:wafv2:eu-central-1:123456789:regional/ipset/test-ipset/abc-123"),
								Addresses:   []*string{aws.String("10.0.0.0/8")},
								Description: aws.String("test"),
							},
							LockToken: aws.String(lockToken),
						}, nil
					},
				},
			},
			want: want{
				obs: managed.ExternalObservation{ResourceExists: true, ResourceUpToDate: true},
			},
		},
		"ResourceExists_NotUpToDate": {
			args: args{
				cr: &svcapitypes.IPSet{
					ObjectMeta: metav1.ObjectMeta{
						Annotations: map[string]string{meta.AnnotationKeyExternalName: ipSetName},
					},
					Spec: svcapitypes.IPSetSpec{
						ForProvider: svcapitypes.IPSetParameters{
							Scope:       aws.String(scope),
							Addresses:   []*string{aws.String("10.0.0.0/8"), aws.String("192.168.0.0/16")},
							Description: aws.String("test"),
						},
					},
				},
				client: &fake.MockWAFV2Client{
					MockListIPSetsWithContext: func(input *svcsdk.ListIPSetsInput) (*svcsdk.ListIPSetsOutput, error) {
						return &svcsdk.ListIPSetsOutput{
							IPSets: []*svcsdk.IPSetSummary{
								{Name: aws.String(ipSetName), Id: aws.String(ipSetID)},
							},
						}, nil
					},
					MockGetIPSetWithContext: func(input *svcsdk.GetIPSetInput) (*svcsdk.GetIPSetOutput, error) {
						return &svcsdk.GetIPSetOutput{
							IPSet: &svcsdk.IPSet{
								Name:        aws.String(ipSetName),
								Id:          aws.String(ipSetID),
								ARN:         aws.String("arn:aws:wafv2:eu-central-1:123456789:regional/ipset/test-ipset/abc-123"),
								Addresses:   []*string{aws.String("10.0.0.0/8")},
								Description: aws.String("test"),
							},
							LockToken: aws.String(lockToken),
						}, nil
					},
				},
			},
			want: want{
				obs: managed.ExternalObservation{ResourceExists: true, ResourceUpToDate: false},
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &external{client: tc.args.client}
			obs, err := e.Observe(context.Background(), tc.args.cr)
			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("Observe(...): -want error, +got error:\n%s", diff)
			}
			// Compare only ResourceExists and ResourceUpToDate
			if obs.ResourceExists != tc.want.obs.ResourceExists {
				t.Errorf("Observe(...): ResourceExists = %v, want %v", obs.ResourceExists, tc.want.obs.ResourceExists)
			}
			if tc.want.obs.ResourceExists && obs.ResourceUpToDate != tc.want.obs.ResourceUpToDate {
				t.Errorf("Observe(...): ResourceUpToDate = %v, want %v", obs.ResourceUpToDate, tc.want.obs.ResourceUpToDate)
			}
		})
	}
}

func TestCreate(t *testing.T) {
	ipSetName := "test-ipset"
	ipSetID := "abc-123"
	lockToken := "lock-token-1"
	scope := "REGIONAL"
	arn := "arn:aws:wafv2:eu-central-1:123456789:regional/ipset/test-ipset/abc-123"

	type want struct {
		err error
	}

	cases := map[string]struct {
		cr     *svcapitypes.IPSet
		client *fake.MockWAFV2Client
		want   want
	}{
		"SuccessfulCreate": {
			cr: &svcapitypes.IPSet{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{meta.AnnotationKeyExternalName: ipSetName},
				},
				Spec: svcapitypes.IPSetSpec{
					ForProvider: svcapitypes.IPSetParameters{
						Region:           "eu-central-1",
						Scope:            aws.String(scope),
						IPAddressVersion: aws.String("IPV4"),
						Addresses:        []*string{aws.String("10.0.0.0/8")},
						Name:             aws.String(ipSetName),
					},
				},
			},
			client: &fake.MockWAFV2Client{
				MockCreateIPSetWithContext: func(input *svcsdk.CreateIPSetInput) (*svcsdk.CreateIPSetOutput, error) {
					return &svcsdk.CreateIPSetOutput{
						Summary: &svcsdk.IPSetSummary{
							ARN:       aws.String(arn),
							Id:        aws.String(ipSetID),
							LockToken: aws.String(lockToken),
							Name:      aws.String(ipSetName),
						},
					}, nil
				},
			},
			want: want{err: nil},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &external{client: tc.client}
			_, err := e.Create(context.Background(), tc.cr)
			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("Create(...): -want error, +got error:\n%s", diff)
			}
			if err == nil {
				if aws.StringValue(tc.cr.Status.AtProvider.ARN) != arn {
					t.Errorf("Create(...): ARN = %v, want %v", aws.StringValue(tc.cr.Status.AtProvider.ARN), arn)
				}
				if aws.StringValue(tc.cr.Status.AtProvider.ID) != ipSetID {
					t.Errorf("Create(...): ID = %v, want %v", aws.StringValue(tc.cr.Status.AtProvider.ID), ipSetID)
				}
			}
		})
	}
}

func TestDelete(t *testing.T) {
	ipSetName := "test-ipset"
	ipSetID := "abc-123"
	lockToken := "lock-token-1"
	scope := "REGIONAL"

	type want struct {
		err error
	}

	cases := map[string]struct {
		cr     *svcapitypes.IPSet
		client *fake.MockWAFV2Client
		want   want
	}{
		"SuccessfulDelete": {
			cr: &svcapitypes.IPSet{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{meta.AnnotationKeyExternalName: ipSetName},
				},
				Spec: svcapitypes.IPSetSpec{
					ForProvider: svcapitypes.IPSetParameters{
						Scope: aws.String(scope),
					},
				},
				Status: svcapitypes.IPSetStatus{
					AtProvider: svcapitypes.IPSetObservation{
						ID:        aws.String(ipSetID),
						LockToken: aws.String(lockToken),
					},
				},
			},
			client: &fake.MockWAFV2Client{
				MockDeleteIPSetWithContext: func(input *svcsdk.DeleteIPSetInput) (*svcsdk.DeleteIPSetOutput, error) {
					return &svcsdk.DeleteIPSetOutput{}, nil
				},
			},
			want: want{err: nil},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &external{client: tc.client}
			_, err := e.Delete(context.Background(), tc.cr)
			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("Delete(...): -want error, +got error:\n%s", diff)
			}
		})
	}
}
