package policy

import (
	_ "embed"
	"testing"

	"github.com/google/go-cmp/cmp"
	"k8s.io/utils/ptr"
)

var (
	//go:embed testdata/UnmarshalSinglesAsList.json
	policyUnmarshalSinglesAsList string

	//go:embed testdata/UnmarshalArrays.json
	policyUnmarshalArrays string
)

func TestParsePolicy(t *testing.T) {
	type args struct {
		rawPolicy string
	}
	type want struct {
		policy *Policy
		err    error
	}
	cases := map[string]struct {
		want
		args
	}{
		"UnmarshalSinglesAsList": {
			args: args{
				rawPolicy: policyUnmarshalSinglesAsList,
			},
			want: want{
				policy: &Policy{
					Version: "2012-10-17",
					Statements: StatementList{
						{
							SID: ptr.To("AllowPutObjectS3ServerAccessLogsPolicy"),
							Principal: &Principal{
								Service: NewStringOrSet(
									"logging.s3.amazonaws.com",
								),
								Federated: ptr.To("cognito-identity.amazonaws.com"),
								AWSPrincipals: NewStringOrSet(
									"123456789012",
								),
							},
							Effect: "Allow",
							Action: StringOrArray{
								"s3:PutObject",
							},
							Resource: StringOrArray{
								"arn:aws:s3:::DOC-EXAMPLE-BUCKET-logs/*",
							},
							Condition: ConditionMap{
								"StringEquals": ConditionSettings{
									"aws:SourceAccount": ConditionSettingsValue{"111111111111"},
								},
								"ArnLike": ConditionSettings{
									"aws:SourceArn": ConditionSettingsValue{"arn:aws:s3:::EXAMPLE-SOURCE-BUCKET"},
								},
							},
						},
					},
				},
			},
		},
		"UnmarshalArrays": {
			args: args{
				rawPolicy: policyUnmarshalArrays,
			},
			want: want{
				policy: &Policy{
					Version: "2012-10-17",
					Statements: StatementList{
						{
							SID: ptr.To("AllowPutObjectS3ServerAccessLogsPolicy"),
							Principal: &Principal{
								Service: NewStringOrSet(
									"logging.s3.amazonaws.com",
									"s3.amazonaws.com",
								),
								Federated: ptr.To("cognito-identity.amazonaws.com"),
								AWSPrincipals: NewStringOrSet(
									"123456789012",
									"452356421222",
								),
							},
							Effect: "Allow",
							Action: StringOrArray{
								"s3:PutObject",
								"s3:GetObject",
							},
							Resource: StringOrArray{
								"arn:aws:s3:::DOC-EXAMPLE-BUCKET-logs/*",
							},
							Condition: ConditionMap{
								"StringEquals": ConditionSettings{
									"aws:SourceAccount": []any{
										"111111111111",
										"111111111112",
									},
								},
								"ArnLike": ConditionSettings{
									"aws:SourceArn": ConditionSettingsValue{"arn:aws:s3:::EXAMPLE-SOURCE-BUCKET"},
								},
							},
						},
						{
							SID:    ptr.To("RestrictToS3ServerAccessLogs"),
							Effect: "Deny",
							Principal: &Principal{
								AllowAnon: true,
							},
							Action: StringOrArray{
								"s3:PutObject",
							},
							Resource: StringOrArray{
								"arn:aws:s3:::DOC-EXAMPLE-BUCKET-logs/*",
							},
							Condition: ConditionMap{
								"ForAllValues:StringNotEquals": ConditionSettings{
									"aws:PrincipalServiceNamesList": ConditionSettingsValue{
										"logging.s3.amazonaws.com",
										"s3.amazonaws.com",
									},
								},
							},
						},
					},
				},
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			policy, err := ParsePolicyString(tc.args.rawPolicy)

			if diff := cmp.Diff(tc.want.policy, &policy); diff != "" {
				t.Errorf("ParsePolicyString(...): -want, +got\n:%s", diff)
			}
			if diff := cmp.Diff(tc.err, err); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
				return
			}
		})
	}
}
