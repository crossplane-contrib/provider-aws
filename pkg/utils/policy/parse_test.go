package policy

import (
	_ "embed"
	"testing"

	"github.com/google/go-cmp/cmp"
)

var (
	//go:embed testdata/UnmarshalSinglesAsList.json
	policyUnmarshalSinglesAsList string

	//go:embed testdata/UnmarshalArrays.json
	policyUnmarshalArrays string

	//go:embed testdata/UnmarshalFromEmbeddedString.raw
	policyUnmarshalFromEmbeddedStringArrays string
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
					Statements: []Statement{
						{
							SID: "AllowPutObjectS3ServerAccessLogsPolicy",
							Principal: &Principal{
								Service: StringOrArray{
									"logging.s3.amazonaws.com",
								},
								Federated: "cognito-identity.amazonaws.com",
								AWSPrincipals: StringOrArray{
									"123456789012",
								},
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
									"aws:SourceAccount": "111111111111",
								},
								"ArnLike": ConditionSettings{
									"aws:SourceArn": "arn:aws:s3:::EXAMPLE-SOURCE-BUCKET",
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
					Statements: []Statement{
						{
							SID: "AllowPutObjectS3ServerAccessLogsPolicy",
							Principal: &Principal{
								Service: StringOrArray{
									"logging.s3.amazonaws.com",
									"s3.amazonaws.com",
								},
								Federated: "cognito-identity.amazonaws.com",
								AWSPrincipals: StringOrArray{
									"123456789012",
									"452356421222",
								},
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
									"aws:SourceArn": "arn:aws:s3:::EXAMPLE-SOURCE-BUCKET",
								},
							},
						},
						{
							SID:    "RestrictToS3ServerAccessLogs",
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
									"aws:PrincipalServiceNamesList": "logging.s3.amazonaws.com",
								},
							},
						},
					},
				},
			},
		},
		"UnmarshalFromEmbeddedStringArrays": {
			args: args{
				rawPolicy: policyUnmarshalFromEmbeddedStringArrays,
			},
			want: want{
				policy: &Policy{
					Version: "2012-10-17",
					Statements: []Statement{
						{
							Effect: "Allow",
							Action: StringOrArray{
								"glue:*",
								"s3:GetBucketLocation",
							},
							Resource: StringOrArray{"*"},
						},
						{
							Effect:   "Allow",
							Action:   StringOrArray{"s3:CreateBucket", "s3:PutBucketPublicAccessBlock"},
							Resource: StringOrArray{"arn:aws:s3:::aws-glue-*"},
						},
						{
							Effect:   "Allow",
							Action:   StringOrArray{"s3:GetObject", "s3:PutObject", "s3:DeleteObject"},
							Resource: StringOrArray{"arn:aws:s3:::aws-glue-*/*", "arn:aws:s3:::*/*aws-glue-*/*"},
						},
						{
							Effect: "Allow",
							Action: StringOrArray{"s3:GetObject"},
							Resource: StringOrArray{
								"arn:aws:s3:::crawler-public*",
								"arn:aws:s3:::aws-glue-*",
							},
						},
						{
							Effect: "Allow",
							Action: StringOrArray{
								"logs:CreateLogGroup", "logs:CreateLogStream", "logs:PutLogEvents",
								"logs:AssociateKmsKey",
							},
							Resource: StringOrArray{"arn:aws:logs:*:*:/aws-glue/*"},
						},
						{
							Effect: "Allow",
							Action: StringOrArray{"ec2:CreateTags", "ec2:DeleteTags"},
							Resource: StringOrArray{
								"arn:aws:ec2:*:*:network-interface/*", "arn:aws:ec2:*:*:security-group/*",
								"arn:aws:ec2:*:*:instance/*",
							},
							Condition: ConditionMap{
								"ForAllValues:StringEquals": {"aws:TagKeys": []any{string("aws-glue-service-resource")}},
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
