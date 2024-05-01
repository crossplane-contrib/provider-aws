package policy

import (
	"github.com/google/go-cmp/cmp"
)

// ArePoliciesEqal determines if the two Policy objects can be considered
// equal.
func ArePoliciesEqal(a, b *Policy) (equal bool, diff string) {
	diff = cmp.Diff(a, b)
	return diff == "", diff
}

// ArePolicyDocumentsEqual determines if the two policy documents can be considered equal.
func ArePolicyDocumentsEqual(a, b string) bool {
	policyA, err := ParsePolicyString(a)
	if err != nil {
		return a == b
	}
	policyB, err := ParsePolicyString(b)
	if err != nil {
		return false
	}
	eq, _ := ArePoliciesEqal(&policyA, &policyB)
	return eq
}
