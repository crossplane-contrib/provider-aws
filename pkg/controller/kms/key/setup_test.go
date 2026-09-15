/*
Copyright 2023 The Crossplane Authors.

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

	"github.com/aws/aws-sdk-go/aws"
	svcsdk "github.com/aws/aws-sdk-go/service/kms"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/golang/mock/gomock"

	svcapitypes "github.com/crossplane-contrib/provider-aws/apis/kms/v1alpha1"
	mockkms "github.com/crossplane-contrib/provider-aws/pkg/clients/mock/kmsiface"
)

const testPolicy = `{"Version":"2012-10-17","Statement":[]}`

func keyWithOrigin(origin string) *svcapitypes.Key {
	cr := &svcapitypes.Key{}
	meta.SetExternalName(cr, "test-key")
	cr.Spec.ForProvider.Policy = aws.String(testPolicy)
	cr.Spec.ForProvider.Origin = aws.String(origin)
	cr.Spec.ForProvider.EnableKeyRotation = aws.Bool(false)
	return cr
}

// TestObserverIsUpToDateSkipsRotationForNonKmsOrigin verifies GetKeyRotationStatus
// is not called for keys whose origin is not AWS_KMS (e.g. AWS_CLOUDHSM or EXTERNAL),
// where the rotation APIs return UnsupportedOperationException.
func TestObserverIsUpToDateSkipsRotationForNonKmsOrigin(t *testing.T) {
	for _, origin := range []string{"AWS_CLOUDHSM", "EXTERNAL", "EXTERNAL_KEY_STORE"} {
		t.Run(origin, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mock := mockkms.NewMockKMSAPI(ctrl)
			o := &observer{client: mock}

			cr := keyWithOrigin(origin)
			obj := &svcsdk.DescribeKeyOutput{
				KeyMetadata: &svcsdk.KeyMetadata{
					KeyId:      aws.String("test-key"),
					KeyState:   aws.String("Enabled"),
					Origin:     aws.String(origin),
					Enabled:    aws.Bool(true),
					KeyManager: aws.String("CUSTOMER"),
				},
			}

			mock.EXPECT().GetKeyPolicy(gomock.Any()).Return(&svcsdk.GetKeyPolicyOutput{Policy: aws.String(testPolicy)}, nil)
			mock.EXPECT().ListResourceTags(gomock.Any()).Return(&svcsdk.ListResourceTagsOutput{}, nil)
			// GetKeyRotationStatus is intentionally NOT expected: gomock fails the
			// test if it is called.

			upToDate, _, err := o.isUpToDate(context.Background(), cr, obj)
			if err != nil {
				t.Fatalf("isUpToDate() unexpected error: %v", err)
			}
			if !upToDate {
				t.Errorf("isUpToDate() = %v, want true", upToDate)
			}
		})
	}
}

// TestUpdaterUpdateSkipsRotationForNonKmsOrigin verifies EnableKeyRotation /
// DisableKeyRotation are not called for non-AWS_KMS origin keys.
func TestUpdaterUpdateSkipsRotationForNonKmsOrigin(t *testing.T) {
	for _, origin := range []string{"AWS_CLOUDHSM", "EXTERNAL", "EXTERNAL_KEY_STORE"} {
		t.Run(origin, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mock := mockkms.NewMockKMSAPI(ctrl)
			u := &updater{client: mock}

			cr := keyWithOrigin(origin)

			mock.EXPECT().PutKeyPolicyWithContext(gomock.Any(), gomock.Any()).Return(&svcsdk.PutKeyPolicyOutput{}, nil)
			mock.EXPECT().ListResourceTagsWithContext(gomock.Any(), gomock.Any()).Return(&svcsdk.ListResourceTagsOutput{}, nil)
			// EnableKeyRotation / DisableKeyRotation are intentionally NOT expected.

			if _, err := u.update(context.Background(), cr); err != nil {
				t.Fatalf("update() unexpected error: %v", err)
			}
		})
	}
}
