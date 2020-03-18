/*
Copyright 2019 The Crossplane Authors.

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

package s3

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/google/go-cmp/cmp"

	"github.com/crossplane/provider-aws/apis/storage/v1beta1"
	aws "github.com/crossplane/provider-aws/pkg/clients"
)

const (
	bucketName = "some bucket"
	policy     = "some policy"
	region     = "us-east-2"
)

func bucketParams(m ...func(*v1beta1.S3BucketParameters)) *v1beta1.S3BucketParameters {
	o := &v1beta1.S3BucketParameters{
		Policy: aws.String(policy),
		Region: region,
	}

	for _, f := range m {
		f(o)
	}

	return o
}

func TestCreatePatch(t *testing.T) {
	type args struct {
		bucketPolicy string
		p            v1beta1.S3BucketParameters
	}

	type want struct {
		patch *v1beta1.S3BucketParameters
	}

	cases := map[string]struct {
		args
		want
	}{
		"SameFields": {
			args: args{
				bucketPolicy: policy,
				p: v1beta1.S3BucketParameters{
					Policy: aws.String(policy),
				},
			},
			want: want{
				patch: &v1beta1.S3BucketParameters{},
			},
		},
		"DifferentFields": {
			args: args{
				bucketPolicy: "",
				p: v1beta1.S3BucketParameters{
					Policy: aws.String(policy),
				},
			},
			want: want{
				patch: &v1beta1.S3BucketParameters{
					Policy: aws.String(policy),
				},
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			result, _ := CreatePatch(tc.args.p, &tc.args.bucketPolicy)
			if diff := cmp.Diff(tc.want.patch, result); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
		})
	}
}

func TestIsUpToDate(t *testing.T) {
	type args struct {
		p            v1beta1.S3BucketParameters
		bucketPolicy string
	}

	cases := map[string]struct {
		args args
		want bool
	}{
		"SameFields": {
			args: args{
				p: v1beta1.S3BucketParameters{
					Policy: aws.String(policy),
				},
				bucketPolicy: policy,
			},
			want: true,
		},
		"DifferentFields": {
			args: args{
				p: v1beta1.S3BucketParameters{
					Policy: aws.String(policy),
				},
				bucketPolicy: "",
			},
			want: false,
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			got, _ := IsUpToDate(tc.args.p, tc.args.bucketPolicy)
			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
		})
	}
}

func TestGenerateCreateBucketInput(t *testing.T) {
	cases := map[string]struct {
		in  v1beta1.S3BucketParameters
		out s3.CreateBucketInput
	}{
		"FilledInput": {
			in: *bucketParams(),
			out: s3.CreateBucketInput{
				Bucket:                    aws.String(bucketName),
				CreateBucketConfiguration: &s3.CreateBucketConfiguration{LocationConstraint: s3.BucketLocationConstraint(region)},
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			r := GenerateCreateBucketInput(bucketName, &tc.in)
			if diff := cmp.Diff(r, &tc.out); diff != "" {
				t.Errorf("GenerateNetworkObservation(...): -want, +got:\n%s", diff)
			}
		})
	}
}
