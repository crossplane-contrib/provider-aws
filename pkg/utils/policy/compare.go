package policy

import (
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

// ArePoliciesEqual determines if the two Policy objects can be considered
// equal.
func ArePoliciesEqual(a, b *Policy) (equal bool, diff string) {
	sortSlice := cmpopts.SortSlices(func(a, b string) bool { return a > b })
	diff = cmp.Diff(a, b, sortSlice)
	return diff == "", diff
}
