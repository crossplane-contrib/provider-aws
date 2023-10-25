package ecr

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"

	"github.com/aws/aws-sdk-go-v2/service/ecr"
	awsecrtypes "github.com/aws/aws-sdk-go-v2/service/ecr/types"

	"github.com/crossplane-contrib/provider-aws/apis/ecr/v1beta1"
	"github.com/crossplane-contrib/provider-aws/pkg/utils/pointer"
)

const (
	errNotSpecified = "failed to format Repository Policy, no rawPolicy or policy specified"
)

// RepositoryPolicyClient is the external client used for Repository Policy Resource
type RepositoryPolicyClient interface {
	SetRepositoryPolicy(ctx context.Context, input *ecr.SetRepositoryPolicyInput, opts ...func(*ecr.Options)) (*ecr.SetRepositoryPolicyOutput, error)
	DeleteRepositoryPolicy(ctx context.Context, input *ecr.DeleteRepositoryPolicyInput, opts ...func(*ecr.Options)) (*ecr.DeleteRepositoryPolicyOutput, error)
	GetRepositoryPolicy(ctx context.Context, input *ecr.GetRepositoryPolicyInput, opts ...func(*ecr.Options)) (*ecr.GetRepositoryPolicyOutput, error)
}

// GenerateSetRepositoryPolicyInput Generates the CreateRepositoryInput from the RepositoryPolicyParameters
func GenerateSetRepositoryPolicyInput(params *v1beta1.RepositoryPolicyParameters, policy *string) *ecr.SetRepositoryPolicyInput {
	c := &ecr.SetRepositoryPolicyInput{
		RepositoryName: params.RepositoryName,
		RegistryId:     params.RegistryID,
		PolicyText:     policy,
		Force:          pointer.BoolValue(params.Force),
	}

	return c
}

// LateInitializeRepositoryPolicy fills the empty fields in *v1alpha1.RepositoryPolicyParameters with
// the values seen in ecr.GetRepositoryPolicyResponse.
func LateInitializeRepositoryPolicy(in *v1beta1.RepositoryPolicyParameters, r *ecr.GetRepositoryPolicyOutput) {
	if r == nil {
		return
	}
	in.RegistryID = pointer.LateInitialize(in.RegistryID, r.RegistryId)
}

// IsPolicyNotFoundErr returns true if the error code indicates that the policy was not found
func IsPolicyNotFoundErr(err error) bool {
	var notFoundError *awsecrtypes.RepositoryPolicyNotFoundException
	return errors.As(err, &notFoundError)
}

// Serialize is the custom marshaller for the RepositoryPolicyBody
func Serialize(p *v1beta1.RepositoryPolicyBody) (interface{}, error) {
	m := make(map[string]interface{})
	m["Version"] = p.Version
	if p.ID != nil && *p.ID != "" {
		m["Id"] = p.ID
	}
	slc := make([]interface{}, len(p.Statements))
	for i, v := range p.Statements {
		msg, err := SerializeRepositoryPolicyStatement(v)
		if err != nil {
			return nil, err
		}
		slc[i] = msg
	}
	m["Statement"] = slc
	return m, nil
}

// SerializeRepositoryPolicyStatement is the custom marshaller for the RepositoryPolicyStatement
func SerializeRepositoryPolicyStatement(p v1beta1.RepositoryPolicyStatement) (interface{}, error) { //nolint:gocyclo
	m := make(map[string]interface{})
	if p.Principal != nil {
		principal, err := SerializeRepositoryPrincipal(p.Principal)
		if err != nil {
			return nil, err
		}
		m["Principal"] = principal
	}
	if p.NotPrincipal != nil {
		notPrincipal, err := SerializeRepositoryPrincipal(p.NotPrincipal)
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
		condition, err := SerializeRepositoryCondition(p.Condition)
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

// SerializeRepositoryPrincipal is the custom serializer for the RepositoryPrincipal
func SerializeRepositoryPrincipal(p *v1beta1.RepositoryPrincipal) (interface{}, error) {
	all := "*"
	if pointer.BoolValue(p.AllowAnon) {
		return all, nil
	}
	m := make(map[string]interface{})
	if p.Service != nil {
		m["Service"] = tryFirst(p.Service)
	}

	if len(p.AWSPrincipals) == 1 {
		m["AWS"] = pointer.StringValue(SerializeAWSPrincipal(p.AWSPrincipals[0]))
	} else if len(p.AWSPrincipals) > 1 {
		values := make([]interface{}, len(p.AWSPrincipals))
		for i := range p.AWSPrincipals {
			values[i] = pointer.StringValue(SerializeAWSPrincipal(p.AWSPrincipals[i]))
		}
		m["AWS"] = values
	}
	return m, nil
}

// SerializeAWSPrincipal converts an AWSPrincipal to a string
func SerializeAWSPrincipal(p v1beta1.AWSPrincipal) *string {
	switch {
	case p.AWSAccountID != nil:
		// Note: AWS Docs say you can specify the account ID either
		// raw or as an ARN, but AWS actually converts internally to
		// the ARN format, which is problematic for checking if we're
		// up to date. So here we just do the conversion ourselves if
		// we were given a string containing a number that looks like an
		// AWS account ID (looks like a 12-digit integer).
		if _, err := strconv.ParseInt(*p.AWSAccountID, 10, 64); err == nil {
			s := fmt.Sprintf("arn:aws:iam::%s:root", *p.AWSAccountID)
			return &s
		}
		return p.AWSAccountID
	case p.IAMRoleARN != nil:
		return p.IAMRoleARN
	case p.UserARN != nil:
		return p.UserARN
	default:
		return nil
	}
}

// SerializeRepositoryCondition converts the string -> Condition map
// into a serialized version
func SerializeRepositoryCondition(p []v1beta1.Condition) (interface{}, error) {
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

// RawPolicyData parses and formats the RepositoryPolicy struct
func RawPolicyData(original *v1beta1.RepositoryPolicy) (string, error) {
	if original == nil {
		return "", errors.New(errNotSpecified)
	}
	switch {
	case original.Spec.ForProvider.RawPolicy != nil:
		return *original.Spec.ForProvider.RawPolicy, nil
	case original.Spec.ForProvider.Policy != nil:
		c := original.DeepCopy()
		body, err := Serialize(c.Spec.ForProvider.Policy)
		if err != nil {
			return "", err
		}
		byteData, err := json.Marshal(body)
		if err != nil {
			return "", err
		}
		str := string(byteData)
		return str, nil
	}
	return "", errors.New(errNotSpecified)
}
