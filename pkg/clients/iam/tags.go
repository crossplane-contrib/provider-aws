package iam

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	iamtypes "github.com/aws/aws-sdk-go-v2/service/iam/types"

	"github.com/crossplane/provider-aws/apis/iam/v1beta1"
)

// DiffIAMTags returns the lists of tags that need to be removed and added according
// to current and desired states, also returns if desired state needs to be updated
func DiffIAMTags(local map[string]string, remote []iamtypes.Tag) (add []iamtypes.Tag, remove []string, areTagsUpToDate bool) {
	addMap, remove, areTagsUpToDate := diffTags(local, remote)

	for k, v := range addMap {
		add = append(add, iamtypes.Tag{Key: aws.String(k), Value: aws.String(v)})
	}
	return add, remove, areTagsUpToDate
}

// DiffIAMTagsWithUpdates returns the lists of tags that need to be removed and added according
// to current and desired states;
// tags that have changed will be returned in the addOrUpdate return parameter, but not included in the `remove` return parameters
// it also returns if desired state needs to be updated
func DiffIAMTagsWithUpdates(local []v1beta1.Tag, remote []iamtypes.Tag) (addOrUpdate []iamtypes.Tag, remove []string, areTagsUpToDate bool) {
	crTagMap := make(map[string]string, len(local))
	for _, v := range local {
		crTagMap[v.Key] = v.Value
	}
	toAdd, toRemove, areTagsUpToDate := diffTags(crTagMap, remote)

	for _, k := range toRemove {
		if _, ok := toAdd[k]; !ok {
			remove = append(remove, k)
		}
	}
	for k, v := range toAdd {
		addOrUpdate = append(addOrUpdate, iamtypes.Tag{Key: aws.String(k), Value: aws.String(v)})
	}

	return addOrUpdate, remove, areTagsUpToDate
}

func diffTags(local map[string]string, remote []iamtypes.Tag) (add map[string]string, remove []string, areTagsUpToDate bool) {
	removeMap := map[string]struct{}{}
	for _, t := range remote {
		if local[aws.ToString(t.Key)] == aws.ToString(t.Value) {
			delete(local, aws.ToString(t.Key))
			continue
		}
		removeMap[aws.ToString(t.Key)] = struct{}{}
	}

	add = make(map[string]string, len(local))
	for k, v := range local {
		add[k] = v
	}
	for k := range removeMap {
		remove = append(remove, k)
	}
	areTagsUpToDate = len(add) == 0 && len(remove) == 0

	return add, remove, areTagsUpToDate
}
