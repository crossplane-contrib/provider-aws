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

package repository

import (
	"encoding/json"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/crossplane/provider-aws/apis/ecr/v1alpha1"
)

func TestDecodeEncodeLifecyclePolicy(t *testing.T) {
	var policies v1alpha1.LifecyclePolicyRule
	if err := json.Unmarshal([]byte(lifecyclePolicyString), &policies); err != nil {
		t.Errorf("could not decode policystring \n%s\n, err: %+v", lifecyclePolicyString, err)
	}

	var multiplePolicies v1alpha1.LifecyclePolicy
	if err := json.Unmarshal([]byte(multipleLifecyclePoliciesString), &multiplePolicies); err != nil {
		t.Errorf("could not decode multiple \n%s\n policystrings, err: %+v", multipleLifecyclePoliciesString, err)
	}

	lifecyclePolicy := v1alpha1.LifecyclePolicy{
		Rules: []v1alpha1.LifecyclePolicyRule{
			{
				RulePriority: 1,
				Description:  "Expire images older than 14 days",
				Selection: v1alpha1.LifecyclePolicySelection{
					TagStatus:   "untagged",
					CountType:   "sinceImagePushed",
					CountUnit:   "days",
					CountNumber: 14,
				},
				Action: v1alpha1.LifecyclePolicyAction{
					Type: "expire",
				},
			},
		},
	}

	policy, err := json.Marshal(lifecyclePolicy)
	if err != nil {
		t.Errorf("could not marshal policy as json, err: %+v", err)
	}
	if diff := cmp.Diff(lifecyclePolicyString, string(policy)); diff != "" {
		t.Errorf("marshal diff: -want, +got:\n%s", diff)
	}

	lifecyclePolicyMultiple := v1alpha1.LifecyclePolicy{
		Rules: []v1alpha1.LifecyclePolicyRule{
			{
				RulePriority: 1,
				Description:  "Expire images older than 14 days",
				Selection: v1alpha1.LifecyclePolicySelection{
					TagStatus:   "untagged",
					CountType:   "sinceImagePushed",
					CountUnit:   "days",
					CountNumber: 14,
				},
				Action: v1alpha1.LifecyclePolicyAction{
					Type: "expire",
				},
			},
		},
	}
	policyMultiple, err := json.Marshal(lifecyclePolicyMultiple)
	if err != nil {
		t.Errorf("could not marshal policy as json, err: %+v", err)
	}
	if diff := cmp.Diff(lifecyclePolicyString, string(policyMultiple)); diff != "" {
		t.Errorf("marshal diff: -want, +got:\n%s", diff)
	}

}
