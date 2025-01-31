/*
Copyright 2022 The Crossplane Authors.

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

package permission

import (
	"encoding/json"
)

func (p *policyPrincipal) UnmarshalJSON(data []byte) error {
	// Check if its a single action
	var str string

	if err := json.Unmarshal(data, &str); err == nil {
		p.Service = &str
		return nil
	}

	pp := _policyPrincipal{}
	if err := json.Unmarshal(data, &pp); err != nil {
		return err
	}

	p.Service = pp.Service
	p.AWS = pp.AWS
	return nil
}

type policyCondition struct {
	ArnLike      map[string]string `json:"ArnLike,omitempty"`
	StringEquals map[string]string `json:"StringEquals,omitempty"`
}

type policyPrincipal struct {
	Service *string `json:"Service,omitempty"`
	AWS     *string `json:"AWS,omitempty"`
}

type _policyPrincipal policyPrincipal

type policyStatement struct {
	Sid       string          `json:"Sid"`
	Effect    string          `json:"Effect"`
	Action    string          `json:"Action"`
	Resource  string          `json:"Resource"`
	Principal policyPrincipal `json:"Principal"`
	Condition policyCondition `json:"Condition"`
}

func (p *policyStatement) GetPrincipal() string {
	if p.Principal.Service != nil {
		return *p.Principal.Service
	}
	if p.Principal.AWS != nil {
		return *p.Principal.AWS
	}
	return ""
}

func (p *policyStatement) GetPrincipalOrgID() *string {
	return tryGet(p.Condition.StringEquals, "aws:PrincipalOrgID")
}

func (p *policyStatement) GetEventSourceToken() *string {
	return tryGet(p.Condition.StringEquals, "lambda:EventSourceToken")
}

func (p *policyStatement) GetSourceAccount() *string {
	return tryGet(p.Condition.StringEquals, "AWS:SourceAccount")
}

func (p *policyStatement) GetSourceARN() *string {
	return tryGet(p.Condition.ArnLike, "AWS:SourceArn")
}

func tryGet(m map[string]string, key string) *string {
	if m == nil {
		return nil
	}
	if val, ok := m[key]; ok {
		return &val
	}
	return nil
}

type policyDocument struct {
	Version   string            `json:"Version"`
	Statement []policyStatement `json:"Statement"`
}

func (p *policyDocument) StatementByID(sid string) *policyStatement {
	for _, s := range p.Statement {
		if s.Sid == sid {
			return &s
		}
	}
	return nil
}

func parsePolicy(raw string) (*policyDocument, error) {
	document := &policyDocument{}

	if err := json.Unmarshal([]byte(raw), document); err != nil {
		return nil, err
	}
	return document, nil
}
