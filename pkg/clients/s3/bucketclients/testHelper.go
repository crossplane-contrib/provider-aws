package bucketclients

import (
	"errors"
	"net/http"

	"github.com/aws/aws-sdk-go-v2/aws"
	corev1alpha1 "github.com/crossplane/crossplane-runtime/apis/core/v1alpha1"

	"github.com/crossplane/provider-aws/apis/s3/v1beta1"
)

var (
	enabled           = "Enabled"
	suspended         = "Suspended"
	errBoom           = errors.New("boom")
	accelGetFailed    = "cannot get bucket accelerate configuration"
	accelPutFailed    = "cannot put bucket acceleration configuration"
	accelDeleteFailed = "cannot delete bucket acceleration configuration"
	corsGetFailed     = "cannot get bucket CORS configuration"
	corsPutFailed     = "cannot put bucket cors"
	corsDeleteFailed  = "cannot delete bucket CORS configuration"
)

type bucketModifier func(policy *v1beta1.Bucket)

func withConditions(c ...corev1alpha1.Condition) bucketModifier { //nolint
	return func(r *v1beta1.Bucket) { r.Status.ConditionedStatus.Conditions = c }
}

func withAccelerationConfig(s *v1beta1.AccelerateConfiguration) bucketModifier { //nolint
	return func(r *v1beta1.Bucket) { r.Spec.Parameters.AccelerateConfiguration = s }
}

func withSSEConfig(s *v1beta1.ServerSideEncryptionConfiguration) bucketModifier { //nolint
	return func(r *v1beta1.Bucket) { r.Spec.Parameters.ServerSideEncryptionConfiguration = s }
}

func withVersioningConfig(s *v1beta1.VersioningConfiguration) bucketModifier { //nolint
	return func(r *v1beta1.Bucket) { r.Spec.Parameters.VersioningConfiguration = s }
}

func withCORSConfig(s *v1beta1.CORSConfiguration) bucketModifier { //nolint
	return func(r *v1beta1.Bucket) { r.Spec.Parameters.CORSConfiguration = s }
}

func withWebConfig(s *v1beta1.WebsiteConfiguration) bucketModifier { //nolint
	return func(r *v1beta1.Bucket) { r.Spec.Parameters.WebsiteConfiguration = s }
}

func withLoggingConfig(s *v1beta1.LoggingConfiguration) bucketModifier { //nolint
	return func(r *v1beta1.Bucket) { r.Spec.Parameters.LoggingConfiguration = s }
}

func withPayerConfig(s *v1beta1.PaymentConfiguration) bucketModifier { //nolint
	return func(r *v1beta1.Bucket) { r.Spec.Parameters.PayerConfiguration = s }
}

func withTaggingConfig(s *v1beta1.Tagging) bucketModifier { //nolint
	return func(r *v1beta1.Bucket) { r.Spec.Parameters.BucketTagging = s }
}

func withReplConfig(s *v1beta1.ReplicationConfiguration) bucketModifier { //nolint
	return func(r *v1beta1.Bucket) { r.Spec.Parameters.ReplicationConfiguration = s }
}

func withLifecycleConfig(s *v1beta1.BucketLifecycleConfiguration) bucketModifier { //nolint
	return func(r *v1beta1.Bucket) { r.Spec.Parameters.LifecycleConfiguration = s }
}

func withNotificationConfig(s *v1beta1.NotificationConfiguration) bucketModifier { //nolint
	return func(r *v1beta1.Bucket) { r.Spec.Parameters.NotificationConfiguration = s }
}

func bucket(m ...bucketModifier) *v1beta1.Bucket {
	cr := &v1beta1.Bucket{
		Spec: v1beta1.BucketSpec{
			Parameters: v1beta1.BucketParameters{
				ACL:                               aws.String("private"),
				LocationConstraint:                aws.String("us-east-1"),
				GrantFullControl:                  nil,
				GrantRead:                         nil,
				GrantReadACP:                      nil,
				GrantWrite:                        nil,
				GrantWriteACP:                     nil,
				ObjectLockEnabledForBucket:        nil,
				ServerSideEncryptionConfiguration: nil,
				VersioningConfiguration:           nil,
				AccelerateConfiguration:           nil,
				CORSConfiguration:                 nil,
				WebsiteConfiguration:              nil,
				LoggingConfiguration:              nil,
				PayerConfiguration:                nil,
				BucketTagging:                     nil,
				ReplicationConfiguration:          nil,
				LifecycleConfiguration:            nil,
				NotificationConfiguration:         nil,
			},
		},
	}
	for _, f := range m {
		f(cr)
	}
	return cr
}

func createRequest(err error, data interface{}) *aws.Request {
	return &aws.Request{HTTPRequest: &http.Request{}, Retryer: aws.NoOpRetryer{}, Error: err, Data: data}
}
