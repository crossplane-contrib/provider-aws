package iam

import (
	"net/url"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/iam"
	iamtypes "github.com/aws/aws-sdk-go-v2/service/iam/types"
	"github.com/aws/smithy-go/document"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"k8s.io/utils/ptr"

	"github.com/crossplane-contrib/provider-aws/apis/iam/v1beta1"
	"github.com/crossplane-contrib/provider-aws/pkg/utils/pointer"
)

var (
	roleARN                  = "some arn"
	description              = "some-description"
	assumeRolePolicyDocument = `{
		"Version": "2012-10-17",
		"Statement": [
		  {
			"Effect": "Allow",
			"Principal": {
			  "Service": "eks.amazonaws.com"
			},
			"Action": "sts:AssumeRole"
		  }
		]
	   }`
	assumeRolePolicyDocumentWithArrays = `{
		"Statement": [
			{
				"Effect": "Allow",
				"Principal": {
					"Service": ["eks.amazonaws.com"]
				},
				"Action": ["sts:AssumeRole"]
			}
		],
		"Version": "2012-10-17"
	}`
	assumeRolePolicyDocument2 = `{
		"Statement": [
		  {
			"Effect": "Allow",
			"Principal": {
			  "Service": "eks.amazonaws.com"
			},
			"Action": "sts:AssumeRole",
			"Condition": {
			  "StringEquals": {"foo": "bar"}
			}
		  }
		],
		"Version": "2012-10-17"
	}`
	roleID             = "some Id"
	roleName           = "some name"
	tagKey             = "key"
	tagValue           = "value"
	permissionBoundary = "arn:aws:iam::111111111111:policy/permission-boundary"
	createDate         = time.Now()
	region             = "us-east-1"
)

func roleParams(m ...func(*v1beta1.RoleParameters)) *v1beta1.RoleParameters {
	o := &v1beta1.RoleParameters{
		Description:              &description,
		AssumeRolePolicyDocument: assumeRolePolicyDocument,
		MaxSessionDuration:       pointer.ToIntAsInt32(1),
	}

	for _, f := range m {
		f(o)
	}

	return o
}

func role(m ...func(*iamtypes.Role)) *iamtypes.Role {
	o := &iamtypes.Role{
		Description:              &description,
		AssumeRolePolicyDocument: &assumeRolePolicyDocument,
		MaxSessionDuration:       pointer.ToIntAsInt32(1),
	}

	for _, f := range m {
		f(o)
	}

	return o
}

func addRoleOutputFields(r *iamtypes.Role) {
	r.Arn = pointer.ToOrNilIfZeroValue(roleARN)
	r.RoleId = pointer.ToOrNilIfZeroValue(roleID)
	r.CreateDate = &createDate
	r.RoleLastUsed = &iamtypes.RoleLastUsed{
		LastUsedDate: &createDate,
		Region:       &region,
	}
}

func roleObservation(m ...func(*v1beta1.RoleExternalStatus)) *v1beta1.RoleExternalStatus {
	o := &v1beta1.RoleExternalStatus{
		ARN:        roleARN,
		RoleID:     roleID,
		CreateDate: pointer.TimeToMetaTime(&createDate),
		RoleLastUsed: &v1beta1.RoleLastUsed{
			LastUsedDate: pointer.TimeToMetaTime(&createDate),
			Region:       &region,
		},
	}

	for _, f := range m {
		f(o)
	}

	return o
}

func TestGenerateCreateRoleInput(t *testing.T) {
	cases := map[string]struct {
		in  v1beta1.RoleParameters
		out iam.CreateRoleInput
	}{
		"FilledInput": {
			in: *roleParams(),
			out: iam.CreateRoleInput{
				RoleName:                 pointer.ToOrNilIfZeroValue(roleName),
				Description:              &description,
				AssumeRolePolicyDocument: pointer.ToOrNilIfZeroValue(assumeRolePolicyDocument),
				MaxSessionDuration:       pointer.ToIntAsInt32(1),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			r := GenerateCreateRoleInput(roleName, &tc.in)
			if diff := cmp.Diff(r, &tc.out, cmpopts.IgnoreTypes(document.NoSerde{})); diff != "" {
				t.Errorf("GenerateNetworkObservation(...): -want, +got:\n%s", diff)
			}
		})
	}
}

func TestGenerateRoleObservation(t *testing.T) {
	cases := map[string]struct {
		in  iamtypes.Role
		out v1beta1.RoleExternalStatus
	}{
		"AllFilled": {
			in:  *role(addRoleOutputFields),
			out: *roleObservation(),
		},
		"NoRoleId": {
			in: *role(addRoleOutputFields, func(r *iamtypes.Role) {
				r.RoleId = nil
			}),
			out: *roleObservation(func(o *v1beta1.RoleExternalStatus) {
				o.RoleID = ""
			}),
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			r := GenerateRoleObservation(tc.in)
			if diff := cmp.Diff(r, tc.out); diff != "" {
				t.Errorf("GenerateNetworkObservation(...): -want, +got:\n%s", diff)
			}
		})
	}
}

func TestLateInitializeRole(t *testing.T) {
	type args struct {
		spec *v1beta1.RoleParameters
		in   iamtypes.Role
	}
	cases := map[string]struct {
		args args
		want *v1beta1.RoleParameters
	}{
		"AllFilledNoDiff": {
			args: args{
				spec: roleParams(),
				in:   *role(),
			},
			want: roleParams(),
		},
		"AllFilledExternalDiff": {
			args: args{
				spec: roleParams(),
				in: *role(func(r *iamtypes.Role) {
					r.CreateDate = &time.Time{}
				}),
			},
			want: roleParams(),
		},
		"PartialFilled": {
			args: args{
				spec: roleParams(func(p *v1beta1.RoleParameters) {
					p.Description = nil
				}),
				in: *role(),
			},
			want: roleParams(func(p *v1beta1.RoleParameters) {
				p.Description = &description
			}),
		},
		"PointerFields": {
			args: args{
				spec: roleParams(),
				in: *role(func(r *iamtypes.Role) {
					r.Tags = []iamtypes.Tag{
						{
							Key:   &tagKey,
							Value: &tagValue,
						},
					}
					r.PermissionsBoundary = &iamtypes.AttachedPermissionsBoundary{
						PermissionsBoundaryArn: &roleARN,
					}
				}),
			},
			want: roleParams(func(p *v1beta1.RoleParameters) {
				p.Tags = []v1beta1.Tag{
					{
						Key:   tagKey,
						Value: tagValue,
					},
				}
				p.PermissionsBoundary = &roleARN
			}),
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			LateInitializeRole(tc.args.spec, &tc.args.in)
			if diff := cmp.Diff(tc.args.spec, tc.want); diff != "" {
				t.Errorf("LateInitializeSpec(...): -want, +got:\n%s", diff)
			}
		})
	}
}

func TestIsRoleUpToDate(t *testing.T) {
	type args struct {
		role iamtypes.Role
		p    v1beta1.RoleParameters
	}

	cases := map[string]struct {
		args     args
		want     bool
		wantDiff []*regexp.Regexp
	}{
		"SameFields": {
			args: args{
				role: iamtypes.Role{
					AssumeRolePolicyDocument: ptr.To(url.QueryEscape(assumeRolePolicyDocument)),
					Description:              &description,
					MaxSessionDuration:       pointer.ToIntAsInt32(1),
					Path:                     pointer.ToOrNilIfZeroValue("/"),
					Tags: []iamtypes.Tag{{
						Key:   pointer.ToOrNilIfZeroValue("key1"),
						Value: pointer.ToOrNilIfZeroValue("value1"),
					}},
				},
				p: v1beta1.RoleParameters{
					Description:              &description,
					AssumeRolePolicyDocument: assumeRolePolicyDocument,
					MaxSessionDuration:       pointer.ToIntAsInt32(1),
					Path:                     pointer.ToOrNilIfZeroValue("/"),
					Tags: []v1beta1.Tag{{
						Key:   "key1",
						Value: "value1",
					}},
				},
			},
			want: true,
		},
		"SameFieldsWithDifferentPolicyFormat": {
			args: args{
				role: iamtypes.Role{
					AssumeRolePolicyDocument: ptr.To(url.QueryEscape(assumeRolePolicyDocumentWithArrays)),
					Description:              &description,
					MaxSessionDuration:       pointer.ToIntAsInt32(1),
					Path:                     pointer.ToOrNilIfZeroValue("/"),
					Tags: []iamtypes.Tag{{
						Key:   pointer.ToOrNilIfZeroValue("key1"),
						Value: pointer.ToOrNilIfZeroValue("value1"),
					}},
				},
				p: v1beta1.RoleParameters{
					Description:              &description,
					AssumeRolePolicyDocument: assumeRolePolicyDocument,
					MaxSessionDuration:       pointer.ToIntAsInt32(1),
					Path:                     pointer.ToOrNilIfZeroValue("/"),
					Tags: []v1beta1.Tag{{
						Key:   "key1",
						Value: "value1",
					}},
				},
			},
			want: true,
		},
		"AWSInitializedFields": {
			args: args{
				role: iamtypes.Role{
					AssumeRolePolicyDocument: ptr.To(url.QueryEscape(assumeRolePolicyDocument)),
					CreateDate:               &createDate,
					Description:              &description,
					MaxSessionDuration:       pointer.ToIntAsInt32(1),
					Path:                     pointer.ToOrNilIfZeroValue("/"),
					PermissionsBoundary: &iamtypes.AttachedPermissionsBoundary{
						PermissionsBoundaryArn:  &permissionBoundary,
						PermissionsBoundaryType: "Policy",
					},
					RoleLastUsed: &iamtypes.RoleLastUsed{
						LastUsedDate: &createDate,
						Region:       pointer.ToOrNilIfZeroValue("us-east-1"),
					},
					Tags: []iamtypes.Tag{{
						Key:   pointer.ToOrNilIfZeroValue("key1"),
						Value: pointer.ToOrNilIfZeroValue("value1"),
					}},
				},
				p: v1beta1.RoleParameters{
					Description:              &description,
					AssumeRolePolicyDocument: assumeRolePolicyDocument,
					MaxSessionDuration:       pointer.ToIntAsInt32(1),
					Path:                     pointer.ToOrNilIfZeroValue("/"),
					PermissionsBoundary:      &permissionBoundary,
					Tags: []v1beta1.Tag{{
						Key:   "key1",
						Value: "value1",
					}},
				},
			},
			want: true,
		},
		"DifferentPolicy": {
			args: args{
				role: iamtypes.Role{
					AssumeRolePolicyDocument: ptr.To(url.QueryEscape(assumeRolePolicyDocument)),
					Description:              &description,
					MaxSessionDuration:       pointer.ToIntAsInt32(1),
					Path:                     pointer.ToOrNilIfZeroValue("/"),
				},
				p: v1beta1.RoleParameters{
					Description:              &description,
					AssumeRolePolicyDocument: assumeRolePolicyDocument2,
					MaxSessionDuration:       pointer.ToIntAsInt32(1),
					Path:                     pointer.ToOrNilIfZeroValue("/"),
				},
			},
			want: false,
			wantDiff: []*regexp.Regexp{
				regexp.MustCompile("Found observed difference in IAM role"),
				regexp.MustCompile(`- AssumeRolePolicyDocument: &"(%\w\w)+Statement`),
				regexp.MustCompile(`\+ AssumeRolePolicyDocument: &"(%\w\w)+Version`),
			},
		},
		"DifferentFields": {
			args: args{
				role: iamtypes.Role{
					AssumeRolePolicyDocument: ptr.To(url.QueryEscape(assumeRolePolicyDocument)),
					Description:              &description,
					MaxSessionDuration:       pointer.ToIntAsInt32(1),
					Path:                     pointer.ToOrNilIfZeroValue("//"),
					Tags: []iamtypes.Tag{{
						Key:   pointer.ToOrNilIfZeroValue("key1"),
						Value: pointer.ToOrNilIfZeroValue("value1"),
					}},
				},
				p: v1beta1.RoleParameters{
					Description:              &description,
					AssumeRolePolicyDocument: assumeRolePolicyDocument,
					MaxSessionDuration:       pointer.ToIntAsInt32(1),
					Path:                     pointer.ToOrNilIfZeroValue("/"),
					Tags: []v1beta1.Tag{{
						Key:   "key1",
						Value: "value1",
					}},
				},
			},
			want: false,
			wantDiff: []*regexp.Regexp{
				regexp.MustCompile("Found observed difference in IAM role"),
				regexp.MustCompile(`- Path: &"/"`),
				regexp.MustCompile(`\+ Path: &"//"`),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			got, testDiff, err := IsRoleUpToDate(tc.args.p, tc.args.role)
			if err != nil {
				t.Errorf("r: unexpected error: %v", err)
			}
			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
			if tc.wantDiff == nil {
				if diff := cmp.Diff("", testDiff); diff != "" {
					t.Errorf("r: -want, +got:\n%s", diff)
				}
			} else {
				// cmp randomly uses either regular or non-breaking spaces.
				// Replace them all with regular spaces.
				compactDiff := strings.Join(strings.Fields(testDiff), " ")
				for _, wantDiff := range tc.wantDiff {
					if !wantDiff.MatchString(compactDiff) {
						t.Errorf("expected:\n%s\nto match:\n%s", testDiff, wantDiff.String())
					}
				}
			}
		})
	}
}
