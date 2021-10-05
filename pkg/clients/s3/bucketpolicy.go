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

package s3

import (
	"context"
	"errors"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	s3types "github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/aws/smithy-go"

	"github.com/crossplane/provider-aws/apis/s3/v1alpha3"
)

// BucketPolicyClient is the external client used for S3BucketPolicy Custom Resource
type BucketPolicyClient interface {
	GetBucketPolicy(ctx context.Context, input *s3.GetBucketPolicyInput, opts ...func(*s3.Options)) (*s3.GetBucketPolicyOutput, error)
	PutBucketPolicy(ctx context.Context, input *s3.PutBucketPolicyInput, opts ...func(*s3.Options)) (*s3.PutBucketPolicyOutput, error)
	DeleteBucketPolicy(ctx context.Context, input *s3.DeleteBucketPolicyInput, opts ...func(*s3.Options)) (*s3.DeleteBucketPolicyOutput, error)
}

// NewBucketPolicyClient returns a new client given an aws config
func NewBucketPolicyClient(cfg aws.Config) BucketPolicyClient {
	return s3.NewFromConfig(cfg)
}

// IsErrorPolicyNotFound returns true if the error code indicates that the item was not found
func IsErrorPolicyNotFound(err error) bool {
	if s3Err, ok := err.(smithy.APIError); ok {
		if s3Err.ErrorCode() == "NoSuchBucketPolicy" {
			return true
		}
	}
	return false
}

// IsErrorBucketNotFound returns true if the error code indicates that the bucket was not found
func IsErrorBucketNotFound(err error) bool {
	var nsb *s3types.NoSuchBucket
	return errors.As(err, &nsb)
}

// Serialize is the custom marshaller for the BucketPolicyParameters
func Serialize(p *v1alpha3.BucketPolicyBody) (interface{}, error) {
	m := make(map[string]interface{})
	m["Version"] = p.Version
	if p.ID != "" {
		m["Id"] = p.ID
	}
	slc := make([]interface{}, len(p.Statements))
	for i, v := range p.Statements {
		msg, err := SerializeBucketPolicyStatement(v)
		if err != nil {
			return nil, err
		}
		slc[i] = msg
	}
	m["Statement"] = slc
	return m, nil
}

// SerializeBucketPolicyStatement is the custom marshaller for the BucketPolicyStatement
func SerializeBucketPolicyStatement(p v1alpha3.BucketPolicyStatement) (interface{}, error) { // nolint:gocyclo
	m := make(map[string]interface{})
	if p.Principal != nil {
		principal, err := SerializeBucketPrincipal(p.Principal)
		if err != nil {
			return nil, err
		}
		m["Principal"] = principal
	}
	if p.NotPrincipal != nil {
		notPrincipal, err := SerializeBucketPrincipal(p.NotPrincipal)
		if err != nil {
			return nil, err
		}
		m["NotPrincipal"] = notPrincipal
	}
	if checkExistsArray(p.Action) {
		m["Action"] = tryFirst(p.Action)
	}
	if checkExistsArray(p.NotAction) {
		m["NotAction"] = tryFirst(p.NotAction)
	}
	if checkExistsArray(p.Resource) {
		m["Resource"] = tryFirst(p.Resource)
	}
	if checkExistsArray(p.NotResource) {
		m["NotResource"] = tryFirst(p.NotResource)
	}
	if p.Condition != nil {
		condition, err := SerializeBucketCondition(p.Condition)
		if err != nil {
			return nil, err
		}
		m["Condition"] = condition
	}
	m["Effect"] = p.Effect
	if p.SID != nil {
		m["Sid"] = *p.SID
	}
	return m, nil
}

// SerializeBucketPrincipal is the custom serializer for the BucketPrincipal
func SerializeBucketPrincipal(p *v1alpha3.BucketPrincipal) (interface{}, error) {
	all := "*"
	if p.AllowAnon {
		return all, nil
	}
	m := make(map[string]interface{})
	if p.Service != nil {
		m["Service"] = tryFirst(p.Service)
	}
	if p.Federated != nil {
		m["Federated"] = aws.ToString(p.Federated)
	}
	if len(p.AWSPrincipals) == 1 {
		m["AWS"] = aws.ToString(SerializeAWSPrincipal(p.AWSPrincipals[0]))
	} else if len(p.AWSPrincipals) > 1 {
		values := make([]interface{}, len(p.AWSPrincipals))
		for i := range p.AWSPrincipals {
			values[i] = aws.ToString(SerializeAWSPrincipal(p.AWSPrincipals[i]))
		}
		m["AWS"] = values
	}
	return m, nil
}

// SerializeAWSPrincipal converts an AWSPrincipal to a string
func SerializeAWSPrincipal(p v1alpha3.AWSPrincipal) *string {
	switch {
	case p.AWSAccountID != nil:
		return p.AWSAccountID
	case p.IAMRoleARN != nil:
		return p.IAMRoleARN
	case p.IAMUserARN != nil:
		return p.IAMUserARN
	default:
		return nil
	}
}

// SerializeBucketCondition converts the string -> Condition map
// into a serialized version
func SerializeBucketCondition(p []v1alpha3.Condition) (interface{}, error) {
	m := make(map[string]interface{})
	for _, v := range p {
		subMap := make(map[string]interface{})
		for _, c := range v.Conditions {
			switch {
			case c.ConditionStringValue != nil:
				subMap[c.ConditionKey] = *c.ConditionStringValue
			case c.ConditionBooleanValue != nil:
				subMap[c.ConditionKey] = *c.ConditionBooleanValue
			case c.ConditionNumericValue != nil:
				subMap[c.ConditionKey] = *c.ConditionNumericValue
			case c.ConditionDateValue != nil:
				subMap[c.ConditionKey] = c.ConditionDateValue.Time.Format("2006-01-02T15:04:05-0700")
			case c.ConditionListValue != nil:
				subMap[c.ConditionKey] = c.ConditionListValue
			default:
				return nil, fmt.Errorf("no value provided for key with value %s, condition %s", c.ConditionKey, v.OperatorKey)
			}
		}
		m[v.OperatorKey] = subMap
	}
	return m, nil
}

func checkExistsArray(slc []string) bool {
	return len(slc) != 0
}

func tryFirst(slc []string) interface{} {
	if len(slc) == 1 {
		return slc[0]
	}
	return slc
}
