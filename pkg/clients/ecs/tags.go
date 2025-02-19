package ecs

import (
	svcsdk "github.com/aws/aws-sdk-go/service/ecs"

	svcapitypes "github.com/crossplane-contrib/provider-aws/apis/ecs/v1alpha1"
	"github.com/crossplane-contrib/provider-aws/pkg/utils/pointer"
)

// DiffTags returns tags that should be added or removed.
func DiffTags(spec []*svcapitypes.Tag, current []*svcapitypes.Tag) (addTags []*svcsdk.Tag, remove []*string) {
	addMap := make(map[string]string, len(spec))
	for _, t := range spec {
		addMap[pointer.StringValue(t.Key)] = pointer.StringValue(t.Value)
	}
	removeMap := make(map[string]string, len(spec))
	for _, t := range current {
		if addMap[pointer.StringValue(t.Key)] == pointer.StringValue(t.Value) {
			delete(addMap, pointer.StringValue(t.Key))
			continue
		}
		removeMap[pointer.StringValue(t.Key)] = pointer.StringValue(t.Value)
	}
	for k, v := range addMap {
		addTags = append(addTags, &svcsdk.Tag{Key: pointer.ToOrNilIfZeroValue(k), Value: pointer.ToOrNilIfZeroValue(v)})
	}
	for k := range removeMap {
		remove = append(remove, pointer.ToOrNilIfZeroValue(k))
	}
	return
}
