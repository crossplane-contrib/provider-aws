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

package key

import (
	"context"
	"testing"

	svcsdk "github.com/aws/aws-sdk-go/service/kms"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/golang/mock/gomock"

	svcapitypes "github.com/crossplane-contrib/provider-aws/apis/kms/v1alpha1"
	mockkms "github.com/crossplane-contrib/provider-aws/pkg/clients/mock/kmsiface"
	"github.com/crossplane-contrib/provider-aws/pkg/utils/pointer"
)

const testKeyID = "test-key-id"

func newTestKey(origin *string) *svcapitypes.Key {
	cr := &svcapitypes.Key{
		Spec: svcapitypes.KeySpec{
			ForProvider: svcapitypes.KeyParameters{
				Origin: origin,
				CustomKeyParameters: svcapitypes.CustomKeyParameters{
					EnableKeyRotation: pointer.ToOrNilIfZeroValue(true),
					Enabled:           pointer.ToOrNilIfZeroValue(true),
				},
			},
		},
		Status: svcapitypes.KeyStatus{
			AtProvider: svcapitypes.KeyObservation{
				Enabled: pointer.ToOrNilIfZeroValue(true),
			},
		},
	}
	meta.SetExternalName(cr, testKeyID)
	return cr
}

func TestIsUpToDateSkipsKeyRotationForNonAWSKMSOrigin(t *testing.T) {
	cases := map[string]struct {
		origin *string
		setup  func(mock *mockkms.MockKMSAPI)
	}{
		"AWSKMSOriginChecksRotationStatus": {
			origin: pointer.ToOrNilIfZeroValue(string(svcapitypes.OriginType_AWS_KMS)),
			setup: func(mock *mockkms.MockKMSAPI) {
				mock.EXPECT().GetKeyRotationStatus(gomock.Any()).Return(&svcsdk.GetKeyRotationStatusOutput{
					KeyRotationEnabled: pointer.ToOrNilIfZeroValue(true),
				}, nil)
			},
		},
		"CloudHSMOriginSkipsRotationStatus": {
			origin: pointer.ToOrNilIfZeroValue(string(svcapitypes.OriginType_AWS_CLOUDHSM)),
			setup:  func(mock *mockkms.MockKMSAPI) {},
		},
		"ExternalOriginSkipsRotationStatus": {
			origin: pointer.ToOrNilIfZeroValue(string(svcapitypes.OriginType_EXTERNAL)),
			setup:  func(mock *mockkms.MockKMSAPI) {},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			mockClient := mockkms.NewMockKMSAPI(gomock.NewController(t))
			mockClient.EXPECT().GetKeyPolicy(gomock.Any()).Return(&svcsdk.GetKeyPolicyOutput{}, nil)
			mockClient.EXPECT().ListResourceTags(gomock.Any()).Return(&svcsdk.ListResourceTagsOutput{}, nil)
			tc.setup(mockClient)

			o := &observer{client: mockClient}
			cr := newTestKey(tc.origin)
			obj := &svcsdk.DescribeKeyOutput{
				KeyMetadata: &svcsdk.KeyMetadata{
					Origin: tc.origin,
				},
			}

			// gomock fails the test itself if an unexpected call (e.g.
			// GetKeyRotationStatus for a non-AWS_KMS origin) is made, so we
			// only need to assert that isUpToDate returns without error.
			if _, _, err := o.isUpToDate(context.Background(), cr, obj); err != nil {
				t.Errorf("isUpToDate(...): unexpected error: %v", err)
			}
		})
	}
}

func TestLateInitializeSetsOrigin(t *testing.T) {
	mockClient := mockkms.NewMockKMSAPI(gomock.NewController(t))
	mockClient.EXPECT().GetKeyPolicy(gomock.Any()).Return(&svcsdk.GetKeyPolicyOutput{}, nil)
	mockClient.EXPECT().ListResourceTags(gomock.Any()).Return(&svcsdk.ListResourceTagsOutput{}, nil)

	o := &observer{client: mockClient}
	in := &svcapitypes.KeyParameters{}
	obj := &svcsdk.DescribeKeyOutput{
		KeyMetadata: &svcsdk.KeyMetadata{
			KeyId:  pointer.ToOrNilIfZeroValue(testKeyID),
			Origin: pointer.ToOrNilIfZeroValue(string(svcapitypes.OriginType_AWS_CLOUDHSM)),
		},
	}

	if err := o.lateInitialize(in, obj); err != nil {
		t.Fatalf("lateInitialize(...): unexpected error: %v", err)
	}
	if pointer.StringValue(in.Origin) != string(svcapitypes.OriginType_AWS_CLOUDHSM) {
		t.Errorf("lateInitialize(...): got origin %q, want %q", pointer.StringValue(in.Origin), svcapitypes.OriginType_AWS_CLOUDHSM)
	}
}

func TestUpdateSkipsKeyRotationForNonAWSKMSOrigin(t *testing.T) {
	cases := map[string]struct {
		origin *string
		setup  func(mock *mockkms.MockKMSAPI)
	}{
		"NilOriginPreservesLegacyBehaviour": {
			origin: nil,
			setup: func(mock *mockkms.MockKMSAPI) {
				mock.EXPECT().EnableKeyRotationWithContext(gomock.Any(), gomock.Any()).Return(&svcsdk.EnableKeyRotationOutput{}, nil)
			},
		},
		"AWSKMSOriginEnablesRotation": {
			origin: pointer.ToOrNilIfZeroValue(string(svcapitypes.OriginType_AWS_KMS)),
			setup: func(mock *mockkms.MockKMSAPI) {
				mock.EXPECT().EnableKeyRotationWithContext(gomock.Any(), gomock.Any()).Return(&svcsdk.EnableKeyRotationOutput{}, nil)
			},
		},
		"CloudHSMOriginSkipsRotationCall": {
			origin: pointer.ToOrNilIfZeroValue(string(svcapitypes.OriginType_AWS_CLOUDHSM)),
			setup:  func(mock *mockkms.MockKMSAPI) {},
		},
		"ExternalOriginSkipsRotationCall": {
			origin: pointer.ToOrNilIfZeroValue(string(svcapitypes.OriginType_EXTERNAL)),
			setup:  func(mock *mockkms.MockKMSAPI) {},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			mockClient := mockkms.NewMockKMSAPI(gomock.NewController(t))
			mockClient.EXPECT().PutKeyPolicyWithContext(gomock.Any(), gomock.Any()).Return(&svcsdk.PutKeyPolicyOutput{}, nil)
			mockClient.EXPECT().ListResourceTagsWithContext(gomock.Any(), gomock.Any()).Return(&svcsdk.ListResourceTagsOutput{}, nil)
			tc.setup(mockClient)

			u := &updater{client: mockClient}
			cr := newTestKey(tc.origin)

			// gomock fails the test itself if an unexpected call (e.g.
			// EnableKeyRotationWithContext for a non-AWS_KMS origin) is made.
			if _, err := u.update(context.Background(), cr); err != nil {
				t.Errorf("update(...): unexpected error: %v", err)
			}
		})
	}
}
