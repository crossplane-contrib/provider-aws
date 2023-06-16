package policy

import "encoding/json"

// ParsePolicyBytes from a byte array representing a raw JSOn string.
func ParsePolicyBytes(raw []byte) (Policy, error) {
	policy := Policy{}
	err := json.Unmarshal(raw, &policy)
	return policy, err
}

// ParsePolicyString from a raw JSON string.
func ParsePolicyString(raw string) (Policy, error) {
	return ParsePolicyBytes([]byte(raw))
}

// ParsePolicyObject parses a policy from an object (i.e. an API struct) which
// can be marshalled into JSON.
func ParsePolicyObject(obj any) (Policy, error) {
	input, err := json.Marshal(obj)
	if err != nil {
		return Policy{}, err
	}
	return ParsePolicyBytes(input)
}
