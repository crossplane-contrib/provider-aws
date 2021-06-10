package ecr

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/awserr"
	"github.com/aws/aws-sdk-go-v2/service/ecr"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/crossplane/provider-aws/apis/ecr/v1alpha1"
	awsclient "github.com/crossplane/provider-aws/pkg/clients"
)

const (
	// RepositoryPolicyNotFoundException policy was not found
	RepositoryPolicyNotFoundException = "RepositoryPolicyNotFoundException"

	errNotSpecified = "failed to format Repository Policy, no rawPolicy or policy specified"
)

// RepositoryPolicyClient is the external client used for Repository Policy Resource
type RepositoryPolicyClient interface {
	SetRepositoryPolicyRequest(input *ecr.SetRepositoryPolicyInput) ecr.SetRepositoryPolicyRequest
	DeleteRepositoryPolicyRequest(input *ecr.DeleteRepositoryPolicyInput) ecr.DeleteRepositoryPolicyRequest
	GetRepositoryPolicyRequest(input *ecr.GetRepositoryPolicyInput) ecr.GetRepositoryPolicyRequest
}

// GenerateSetRepositoryPolicyInput Generates the CreateRepositoryInput from the RepositoryPolicyParameters
func GenerateSetRepositoryPolicyInput(params *v1alpha1.RepositoryPolicyParameters, policy *string) *ecr.SetRepositoryPolicyInput {
	c := &ecr.SetRepositoryPolicyInput{
		RepositoryName: params.RepositoryName,
		RegistryId:     params.RegistryID,
		PolicyText:     policy,
		Force:          params.Force,
	}

	return c
}

// LateInitializeRepositoryPolicy fills the empty fields in *v1alpha1.RepositoryPolicyParameters with
// the values seen in ecr.GetRepositoryPolicyResponse.
func LateInitializeRepositoryPolicy(in *v1alpha1.RepositoryPolicyParameters, r *ecr.GetRepositoryPolicyResponse) { // nolint:gocyclo
	if r == nil {
		return
	}
	in.RegistryID = awsclient.LateInitializeStringPtr(in.RegistryID, r.RegistryId)
}

// IsPolicyNotFoundErr returns true if the error code indicates that the policy was not found
func IsPolicyNotFoundErr(err error) bool {
	if ecrErr, ok := err.(awserr.Error); ok && ecrErr.Code() == RepositoryPolicyNotFoundException {
		return true
	}
	return false
}

// Serialize is the custom marshaller for the RepositoryPolicyBody
func Serialize(p *v1alpha1.RepositoryPolicyBody) (interface{}, error) {
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
func SerializeRepositoryPolicyStatement(p v1alpha1.RepositoryPolicyStatement) (interface{}, error) { // nolint:gocyclo
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
func SerializeRepositoryPrincipal(p *v1alpha1.RepositoryPrincipal) (interface{}, error) {
	all := "*"
	if aws.BoolValue(p.AllowAnon) {
		return all, nil
	}
	m := make(map[string]interface{})
	if p.Service != nil {
		m["Service"] = tryFirst(p.Service)
	}

	if len(p.AWSPrincipals) == 1 {
		m["AWS"] = aws.StringValue(SerializeAWSPrincipal(p.AWSPrincipals[0]))
	} else if len(p.AWSPrincipals) > 1 {
		values := make([]interface{}, len(p.AWSPrincipals))
		for i := range p.AWSPrincipals {
			values[i] = aws.StringValue(SerializeAWSPrincipal(p.AWSPrincipals[i]))
		}
		m["AWS"] = values
	}
	return m, nil
}

// SerializeAWSPrincipal converts an AWSPrincipal to a string
func SerializeAWSPrincipal(p v1alpha1.AWSPrincipal) *string {
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
	case p.IAMUserARN != nil:
		return p.IAMUserARN
	default:
		return nil
	}
}

// SerializeRepositoryCondition converts the string -> Condition map
// into a serialized version
func SerializeRepositoryCondition(p []v1alpha1.Condition) (interface{}, error) {
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

// IsRepositoryPolicyUpToDate Marshall policies to json for a compare to get around string ordering
func IsRepositoryPolicyUpToDate(local, remote *string) bool {
	var localUnmarshalled interface{}
	var remoteUnmarshalled interface{}

	var err error
	err = json.Unmarshal([]byte(*local), &localUnmarshalled)
	if err != nil {
		return false
	}
	err = json.Unmarshal([]byte(*remote), &remoteUnmarshalled)
	if err != nil {
		return false
	}

	sortSlicesOpt := cmpopts.SortSlices(func(x, y interface{}) bool {
		if a, ok := x.(string); ok {
			if b, ok := y.(string); ok {
				return a < b
			}
		}
		// Note: Unknown types in slices will not cause a panic, but
		// may not be sorted correctly. Depending on how AWS handles
		// these, it may cause constant updates - but better this than
		// panicing.
		return false
	})
	return cmp.Equal(localUnmarshalled, remoteUnmarshalled, cmpopts.EquateEmpty(), sortSlicesOpt)
}

// RawPolicyData parses and formats the RepositoryPolicy struct
func RawPolicyData(original *v1alpha1.RepositoryPolicy) (string, error) {
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
