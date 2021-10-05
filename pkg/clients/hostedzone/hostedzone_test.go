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

package hostedzone

import (
	"errors"
	"testing"

	route53types "github.com/aws/aws-sdk-go-v2/service/route53/types"
	"github.com/aws/smithy-go"
)

func TestIsErrorNoSuchHostedZone(t *testing.T) {
	tests := map[string]struct {
		err  error
		want bool
	}{
		"validError": {
			err:  &route53types.NoSuchHostedZone{},
			want: true,
		},
		"invalidAwsError": {
			err:  &smithy.GenericAPIError{Code: "something"},
			want: false,
		},
		"randomError": {
			err:  errors.New("the specified hosted zone does not exist"),
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.err.Error(), func(t *testing.T) {
			if got := IsNotFound(tt.err); got != tt.want {
				t.Errorf("IsNotFound() = %v, want %v", got, tt.want)
			}
		})
	}
}
