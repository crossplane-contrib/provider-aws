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
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/awserr"
	"github.com/aws/aws-sdk-go-v2/service/s3"

	"github.com/crossplane/provider-aws/apis/s3/v1alpha1"
)

// BucketPolicyClient is the external client used for S3BucketPolicy Custom Resource
type BucketPolicyClient interface {
	GetBucketPolicyRequest(input *s3.GetBucketPolicyInput) s3.GetBucketPolicyRequest
	PutBucketPolicyRequest(input *s3.PutBucketPolicyInput) s3.PutBucketPolicyRequest
	DeleteBucketPolicyRequest(input *s3.DeleteBucketPolicyInput) s3.DeleteBucketPolicyRequest
}

// NewBucketPolicyClient returns a new client given an aws config
func NewBucketPolicyClient(cfg aws.Config) BucketPolicyClient {
	return s3.New(cfg)
}

// IsErrorPolicyNotFound returns true if the error code indicates that the item was not found
func IsErrorPolicyNotFound(err error) bool {
	if s3Err, ok := err.(awserr.Error); ok && s3Err.Code() == "NoSuchBucketPolicy" {
		return true
	}
	return false
}

// IsErrorBucketNotFound returns true if the error code indicates that the bucket was not found
func IsErrorBucketNotFound(err error) bool {
	if s3Err, ok := err.(awserr.Error); ok && s3Err.Code() == s3.ErrCodeNoSuchBucket {
		return true
	}
	return false
}

// Serialize is the custom marshaller for the BucketPolicyParameters
func Serialize(p v1alpha1.BucketPolicyParameters) (interface{}, error) {
	m := make(map[string]interface{})
	m["Version"] = p.PolicyVersion
	if p.PolicyID != "" {
		m["Id"] = p.PolicyID
	}
	slc := make([]interface{}, len(p.PolicyStatement))
	for i, v := range p.PolicyStatement {
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
func SerializeBucketPolicyStatement(p v1alpha1.BucketPolicyStatement) (interface{}, error) { // nolint:gocyclo
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
	if checkExistsArray(p.PolicyAction) {
		m["Action"] = tryFirst(p.PolicyAction)
	}
	if checkExistsArray(p.NotPolicyAction) {
		m["NotAction"] = tryFirst(p.NotPolicyAction)
	}
	if checkExistsArray(p.ResourcePath) {
		m["Resource"] = tryFirst(p.ResourcePath)
	}
	if checkExistsArray(p.NotResourcePath) {
		m["NotResource"] = tryFirst(p.NotResourcePath)
	}
	if p.ConditionBlock != nil {
		condition, err := SerializeBucketCondition(p.ConditionBlock)
		if err != nil {
			return nil, err
		}
		m["Condition"] = condition
	}
	m["Effect"] = p.Effect
	if p.StatementID != nil {
		m["Sid"] = *p.StatementID
	}
	return m, nil
}

// SerializeBucketPrincipal is the custom serializer for the BucketPrincipal
func SerializeBucketPrincipal(p *v1alpha1.BucketPrincipal) (interface{}, error) {
	all := "*"
	if p.AllowAnon {
		return all, nil
	}
	m := make(map[string]interface{})
	if p.Service != nil {
		m["Service"] = tryFirst(p.Service)
	}
	if p.Federated != nil {
		m["Federated"] = aws.StringValue(p.Federated)
	}
	if len(p.AWSPrincipals) == 1 {
		m["AWS"] = aws.StringValue(SerializeAWSPrincipal(p.AWSPrincipals[0]))
	} else if len(p.AWSPrincipals) > 1 {
		values := make([]string, len(p.AWSPrincipals))
		for i := range p.AWSPrincipals {
			values[i] = aws.StringValue(SerializeAWSPrincipal(p.AWSPrincipals[0]))
		}
		m["AWS"] = values
	}
	return m, nil
}

// SerializeAWSPrincipal converts an AWSPrincipal to a string
func SerializeAWSPrincipal(p v1alpha1.AWSPrincipal) *string {
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
func SerializeBucketCondition(p map[string]v1alpha1.Condition) (interface{}, error) {
	m := make(map[string]interface{})
	for k, v := range p {
		subMap := make(map[string]interface{})
		switch {
		case v.ConditionStringValue != nil:
			subMap[v.ConditionKey] = *v.ConditionStringValue
		case v.ConditionBooleanValue != nil:
			subMap[v.ConditionKey] = *v.ConditionBooleanValue
		case v.ConditionNumericValue != nil:
			subMap[v.ConditionKey] = *v.ConditionNumericValue
		case v.ConditionDateValue != nil:
			subMap[v.ConditionKey] = v.ConditionDateValue.Time.Format("2006-01-02T15:04:05-0700")
		default:
			return nil, fmt.Errorf("no value provided for key with value %s, condition %s", v.ConditionKey, k)
		}
		m[k] = subMap
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
