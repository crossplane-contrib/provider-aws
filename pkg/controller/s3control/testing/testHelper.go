package testing

import (
	"github.com/aws/aws-sdk-go/aws/awserr"
	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/meta"

	"github.com/crossplane-contrib/provider-aws/apis/s3/common"
	svcapitypes "github.com/crossplane-contrib/provider-aws/apis/s3control/v1alpha1"
	awsClient "github.com/crossplane-contrib/provider-aws/pkg/clients"
)

// AccessPointModifier is a function which modifies the AccessPoint for testing
type AccessPointModifier func(bucket *svcapitypes.AccessPoint)

// WithConditions sets the Conditions for an AccessPoint
func WithConditions(c ...xpv1.Condition) AccessPointModifier {
	return func(r *svcapitypes.AccessPoint) { r.Status.ConditionedStatus.Conditions = c }
}

// WithPolicy sets the policy for an AccessPoint
func WithPolicy(s *common.BucketPolicyBody) AccessPointModifier { // nolint
	return func(r *svcapitypes.AccessPoint) { r.Spec.ForProvider.Policy = s }
}

// AccessPoint creates a AccessPoint for use in testing
func AccessPoint(m ...AccessPointModifier) *svcapitypes.AccessPoint {
	cr := &svcapitypes.AccessPoint{
		Spec: svcapitypes.AccessPointSpec{
			ForProvider: svcapitypes.AccessPointParameters{
				Region:    "us-east-1",
				AccountID: awsClient.String("1234567890"),
				CustomAccessPointParameters: svcapitypes.CustomAccessPointParameters{
					BucketName: awsClient.String("test.bucket.name"),
				},
			},
		},
		Status: svcapitypes.AccessPointStatus{},
	}
	for _, f := range m {
		f(cr)
	}
	meta.SetExternalName(cr, "test.accessPoint.name")
	return cr
}

// NoSuchAccessPoint creates an error for use in testing
func NoSuchAccessPoint() awserr.Error {
	return awserr.New("NoSuchAccessPoint", "", nil)
}

// NoSuchAccessPointPolicy creates an error for use in testing
func NoSuchAccessPointPolicy() awserr.Error {
	return awserr.New("NoSuchAccessPointPolicy", "", nil)
}
