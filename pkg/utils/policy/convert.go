/*
Copyright 2023 The Crossplane Authors.

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

package policy

import (
	"encoding/json"

	"github.com/crossplane-contrib/provider-aws/apis/common"
)

// ConvertResourcePolicyToPolicyString converts a ResourcePolicy to its JSON
// string representation that can be sent to AWS.
func ConvertResourcePolicyToPolicyString(rp *common.ResourcePolicy) (*string, error) {
	raw, err := ConvertResourcePolicyToPolicyBytes(rp)
	if err != nil {
		return nil, err
	}
	if raw == nil {
		return nil, nil
	}
	text := string(raw)
	return &text, nil
}

// ConvertResourcePolicyToPolicyBytes converts a ResourcePolicy to its JSON
// representation that can be sent to AWS.
func ConvertResourcePolicyToPolicyBytes(rp *common.ResourcePolicy) ([]byte, error) {
	policy := ConvertResourcePolicyToPolicy(rp)
	if policy == nil {
		return nil, nil
	}
	return json.Marshal(policy)
}

// ConvertResourcePolicyToPolicy converts a ResourcePolicy to a Policy object
// to be better comparable.
func ConvertResourcePolicyToPolicy(rp *common.ResourcePolicy) *Policy {
	if rp == nil {
		return nil
	}

	res := Policy{
		Version: rp.Version,
		ID:      rp.ID,
	}
	for _, sm := range rp.Statements {
		resSm := Statement{
			SID:          sm.SID,
			Effect:       StatementEffect(sm.Effect),
			Action:       sm.Action,
			NotAction:    sm.NotAction,
			Resource:     sm.Resource,
			NotResource:  sm.NotResource,
			Principal:    convertResourcePolicyPrincipal(sm.Principal),
			NotPrincipal: convertResourcePolicyPrincipal(sm.NotPrincipal),
			Condition:    convertResourcePolicyConditions(sm.Condition),
		}
		res.Statements = append(res.Statements, resSm)
	}
	return &res
}

func convertResourcePolicyPrincipal(p *common.ResourcePrincipal) *Principal {
	if p == nil {
		return nil
	}

	res := Principal{
		AllowAnon: p.AllowAnon,
		Federated: p.Federated,
		Service:   NewStringOrSet(p.Service...),
	}
	for _, pr := range p.AWSPrincipals {
		switch {
		case pr.AWSAccountID != nil:
			res.AWSPrincipals = res.AWSPrincipals.Add(*pr.AWSAccountID)
		case pr.IAMRoleARN != nil:
			res.AWSPrincipals = res.AWSPrincipals.Add(*pr.IAMRoleARN)
		case pr.UserARN != nil:
			res.AWSPrincipals = res.AWSPrincipals.Add(*pr.UserARN)
		}
	}
	return &res
}

func convertResourcePolicyConditions(conditions []common.Condition) ConditionMap {
	if conditions == nil {
		return nil
	}

	m := ConditionMap{}
	for _, cc := range conditions {
		set := ConditionSettings{}
		for _, c := range cc.Conditions {

			switch {
			case c.ConditionStringValue != nil:
				set[c.ConditionKey] = ConditionSettingsValue{*c.ConditionStringValue}
			case c.ConditionBooleanValue != nil:
				set[c.ConditionKey] = ConditionSettingsValue{*c.ConditionBooleanValue}
			case c.ConditionNumericValue != nil:
				set[c.ConditionKey] = ConditionSettingsValue{*c.ConditionNumericValue}
			case c.ConditionDateValue != nil:
				set[c.ConditionKey] = ConditionSettingsValue{c.ConditionDateValue.Time.Format("2006-01-02T15:04:05-0700")}
			case c.ConditionListValue != nil:
				listVal := make(ConditionSettingsValue, len(c.ConditionListValue))
				for i, val := range c.ConditionListValue {
					listVal[i] = val
				}
				set[c.ConditionKey] = listVal
			}
		}
		m[cc.OperatorKey] = set
	}
	return m
}
