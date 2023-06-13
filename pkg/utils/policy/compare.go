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
