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

package arn

import (
	"fmt"
	"regexp"

	"github.com/pkg/errors"
)

// ARN represents a
type ARN struct {
	Partition string
	Service   string
	Region    string
	AccountID string
	Resource  string
}

// String returns the string representation of a.
func (a *ARN) String() string {
	return fmt.Sprintf("arn:%s:%s:%s:%s:%s", a.Partition, a.Service, a.Region, a.AccountID, a.Resource)
}

var (
	arnRegex = regexp.MustCompile(`^arn:([\w\d-]*):([\w\d-]*):([\w\d-]*):([\w\d-]*):(.*)$`)
)

// ParseARN extract ARN information from s.
func ParseARN(s string) (ARN, error) {
	match := arnRegex.FindStringSubmatch(s)
	if match == nil {
		return ARN{}, errors.Errorf("%q is not a valid ARN", s)
	}
	return ARN{
		Partition: match[1],
		Service:   match[2],
		Region:    match[3],
		AccountID: match[4],
		Resource:  match[5],
	}, nil
}
