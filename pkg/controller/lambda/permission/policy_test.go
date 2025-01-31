package permission

import (
	"testing"

	"github.com/crossplane/crossplane-runtime/pkg/test"
	"github.com/google/go-cmp/cmp"
)

type args struct {
	rawPolicy string
}

func stringPtr(v string) *string {
	return &v
}

func TestUnmarshalPolicyPrincipal(t *testing.T) {
	type want struct {
		result policyPrincipal
		err    error
	}

	cases := map[string]struct {
		args
		want
	}{
		"PrincipalJustWildcardString": {
			args: args{
				rawPolicy: `"foo"`,
			},
			want: want{
				result: policyPrincipal{
					Service: stringPtr("foo"),
				},
				err: nil,
			},
		},
		"PrincipalJustServiceString": {
			args: args{
				rawPolicy: `"s3.amazonaws.com"`,
			},
			want: want{
				result: policyPrincipal{
					Service: stringPtr("s3.amazonaws.com"),
				},
				err: nil,
			},
		},
		"PrincipalObjectService": {
			args: args{
				rawPolicy: `{"Service":"lambda.amazonaws.com"}`,
			},
			want: want{
				result: policyPrincipal{
					Service: stringPtr("lambda.amazonaws.com"),
				},
				err: nil,
			},
		},
		"PrincipalObjectAWS": {
			args: args{
				rawPolicy: `{"AWS":"aws:arn:iam:::role/test"}`,
			},
			want: want{
				result: policyPrincipal{
					AWS: stringPtr("aws:arn:iam:::role/test"),
				},
				err: nil,
			},
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			// Act
			result := policyPrincipal{}
			err := result.UnmarshalJSON([]byte(tc.args.rawPolicy))

			// Assert
			if diff := cmp.Diff(tc.want.result, result, test.EquateConditions()); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}

			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
		})
	}
}

func TestUnmarshalPolicy(t *testing.T) {
	type want struct {
		result *policyDocument
		err    error
	}

	cases := map[string]struct {
		args
		want
	}{
		"UnmarshalPolicyWithStringAsPrincipal": {
			args: args{
				rawPolicy: `{
					"Version":"version",
					"Id":"default",
					"Statement":[
						{
							"Sid": "sid",
							"Effect": "effect",
							"Principal": "service",
							"Action": "action",
							"Resource": "resource",
							"Condition": {
								"StringEquals": {
									"equals1": "foo",
									"equals2": "bar"
								},
								"ArnLike": {
									"like1": "foo",
									"like2": "bar"
								}
							}
						},
						{
							"Sid": "sid2",
							"Effect": "effect2",
							"Principal": "service2",
							"Action": "action2",
							"Resource": "resource2",
							"Condition": {
								"StringEquals": {
									"equals1": "foo"
								},
								"ArnLike": {
									"like2": "bar"
								}
							}
						}
					]
				}`,
			},
			want: want{
				result: &policyDocument{
					Version: "version",
					Statement: []policyStatement{
						{
							Sid:      "sid",
							Effect:   "effect",
							Action:   "action",
							Resource: "resource",
							Principal: policyPrincipal{
								Service: stringPtr("service"),
							},
							Condition: policyCondition{
								ArnLike: map[string]string{
									"like1": "foo",
									"like2": "bar",
								},
								StringEquals: map[string]string{
									"equals1": "foo",
									"equals2": "bar",
								},
							},
						},
						{
							Sid:      "sid2",
							Effect:   "effect2",
							Action:   "action2",
							Resource: "resource2",
							Principal: policyPrincipal{
								Service: stringPtr("service2"),
							},
							Condition: policyCondition{
								ArnLike: map[string]string{
									"like2": "bar",
								},
								StringEquals: map[string]string{
									"equals1": "foo",
								},
							},
						},
					},
				},
				err: nil,
			},
		},
		"UnmarshalPolicyWithObjectAsPrincipal": {
			args: args{
				rawPolicy: `{
					"Version":"version",
					"Id":"default",
					"Statement":[
						{
							"Sid": "sid",
							"Effect": "effect",
							"Principal": {
								"Service": "service"
							},
							"Action": "action",
							"Resource": "resource",
							"Condition": {
								"StringEquals": {
									"equals1": "foo"
								},
								"ArnLike": {
									"like2": "bar"
								}
							}
						}
					]
				}`,
			},
			want: want{
				result: &policyDocument{
					Version: "version",
					Statement: []policyStatement{
						{
							Sid:      "sid",
							Effect:   "effect",
							Action:   "action",
							Resource: "resource",
							Principal: policyPrincipal{
								Service: stringPtr("service"),
							},
							Condition: policyCondition{
								ArnLike: map[string]string{
									"like2": "bar",
								},
								StringEquals: map[string]string{
									"equals1": "foo",
								},
							},
						},
					},
				},
				err: nil,
			},
		},
		"UnmarshalPolicyWithAWSObjectAsPrincipal": {
			args: args{
				rawPolicy: `{
					"Version":"version",
					"Id":"default",
					"Statement":[
						{
							"Sid": "sid",
							"Effect": "effect",
							"Principal": {
								"AWS": "arn"
							},
							"Action": "action",
							"Resource": "resource",
							"Condition": {
								"StringEquals": {
									"equals1": "foo"
								},
								"ArnLike": {
									"like2": "bar"
								}
							}
						}
					]
				}`,
			},
			want: want{
				result: &policyDocument{
					Version: "version",
					Statement: []policyStatement{
						{
							Sid:      "sid",
							Effect:   "effect",
							Action:   "action",
							Resource: "resource",
							Principal: policyPrincipal{
								AWS: stringPtr("arn"),
							},
							Condition: policyCondition{
								ArnLike: map[string]string{
									"like2": "bar",
								},
								StringEquals: map[string]string{
									"equals1": "foo",
								},
							},
						},
					},
				},
				err: nil,
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			// Act
			result, err := parsePolicy(tc.args.rawPolicy)

			// Assert
			if diff := cmp.Diff(tc.want.result, result, test.EquateConditions()); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}

			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
		})
	}
}
