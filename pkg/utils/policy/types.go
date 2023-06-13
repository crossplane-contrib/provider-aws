package policy

import (
	"encoding/json"
	"strconv"
)

// Policy represents an AWS IAM policy.
type Policy struct {
	// Version is the current IAM policy version
	Version string `json:"Version"`

	// ID is the policy's optional identifier
	ID string `json:"Id,omitempty"`

	// Statements is the list of statement this policy applies.
	Statements StatementList `json:"Statement,omitempty"`
}

// StatementList is a list of statements.
// It implements a custom marshaller to support parsing from a single, non-list
// statement.
type StatementList []Statement

// UnmarshalJSON unmarshals data into s.
func (s *StatementList) UnmarshalJSON(data []byte) error {
	var single Statement
	if err := json.Unmarshal(data, &single); err == nil {
		*s = StatementList{single}
		return nil
	}
	list := []Statement{}
	if err := json.Unmarshal(data, &list); err != nil {
		return err
	}
	*s = list
	return nil
}

// StatementEffect specifies the effect of a policy statement.
type StatementEffect string

// Statement effect values.
const (
	StatementEffectAllow StatementEffect = "Allow"
	StatementEffectDeny  StatementEffect = "Deny"
)

// Statement defines an individual statement within the policy.
type Statement struct {
	// Optional identifier for this statement, must be unique within the
	// policy if provided.
	SID string `json:"Sid,omitempty"`

	// The effect is required and specifies whether the statement results
	// in an allow or an explicit deny.
	// Valid values for Effect are "Allow" and "Deny".
	Effect StatementEffect `json:"Effect"`

	// Used with the policy to specify the principal that is allowed
	// or denied access to a resource.
	Principal *Principal `json:"Principal,omitempty"`

	// Used with the S3 policy to specify the users which are not included
	// in this policy
	NotPrincipal *Principal `json:"NotPrincipal,omitempty"`

	// Action specifies the action or actions that will be allowed or denied
	// with this Statement.
	Action StringOrArray `json:"Action,omitempty"`

	// NotAction specifies each element that will allow the property to match
	// all but the listed actions.
	NotAction StringOrArray `json:"NotAction,omitempty"`

	// Resource specifies paths on which this statement will apply.
	Resource StringOrArray `json:"Resource,omitempty"`

	// NotResource explicitly specifies all resource paths that are defined in
	// this array.
	NotResource StringOrArray `json:"NotResource,omitempty"`

	// Condition specifies where conditions for policy are in effect.
	// https://docs.aws.amazon.com/AmazonS3/latest/dev/amazon-s3-policy-keys.html
	Condition ConditionMap `json:"Condition,omitempty"`
}

// Principal defines the principal users affected by
// the PolicyStatement
// Please see the AWS S3 docs for more information
// https://docs.aws.amazon.com/IAM/latest/UserGuide/reference_policies_elements_principal.html
type Principal struct {
	// This flag indicates if the policy should be made available
	// to all anonymous users. Also known as "*".
	// +optional
	AllowAnon bool `json:"-"`

	// This list contains the all of the AWS IAM users which are affected
	// by the policy statement.
	// +optional
	AWSPrincipals StringOrArray `json:"AWS,omitempty"`

	// This string contains the identifier for any federated web identity
	// provider.
	// +optional
	Federated string `json:"Federated,omitempty"`

	// Service define the services which can have access to this bucket
	// +optional
	Service StringOrArray `json:"Service,omitempty"`
}

type unmarshalPrinciple Principal

// UnmarshalJSON unmarshals data into p.
func (p *Principal) UnmarshalJSON(data []byte) error {
	var single string
	if err := json.Unmarshal(data, &single); err == nil {
		if single == "*" {
			p.AllowAnon = true
		}
		return nil
	}
	var res unmarshalPrinciple
	if err := json.Unmarshal(data, &res); err != nil {
		return err
	}
	*p = Principal(res)
	return nil
}

// ConditionMap is map with the operator as key and the setting as values.
// See https://docs.aws.amazon.com/IAM/latest/UserGuide/reference_policies_elements_condition.html
// for details.
type ConditionMap map[string]ConditionSettings

// ConditionSettings is a map of keys and values.
// Depending on the type of operation, the values can strings, integers,
// bools or lists of strings.
type ConditionSettings map[string]any

// UnmarshalJSON unmarshals data into m.
func (m *ConditionSettings) UnmarshalJSON(data []byte) error {
	res := map[string]any{}
	if err := json.Unmarshal(data, &res); err != nil {
		return err
	}
	for k, v := range res {
		// AWS converts bools into strings in conditions.
		if b, isBool := v.(bool); isBool {
			res[k] = strconv.FormatBool(b)
		}
	}
	*m = res
	return nil
}

// StringOrArray is a string array that supports parsing from a single string
// as a single entry array.
type StringOrArray []string

// UnmarshalJSON unmarshals data into s.
func (s *StringOrArray) UnmarshalJSON(data []byte) error {
	var single string
	if err := json.Unmarshal(data, &single); err == nil {
		*s = StringOrArray{single}
		return nil
	}
	list := []string{}
	if err := json.Unmarshal(data, &list); err != nil {
		return err
	}
	*s = list
	return nil
}
