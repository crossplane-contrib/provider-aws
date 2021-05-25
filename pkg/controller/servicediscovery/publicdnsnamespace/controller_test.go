package publicdnsnamespace

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	svcsdk "github.com/aws/aws-sdk-go/service/servicediscovery"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/test"

	svcapitypes "github.com/crossplane/provider-aws/apis/servicediscovery/v1alpha1"
	awsclient "github.com/crossplane/provider-aws/pkg/clients"
	svcclient "github.com/crossplane/provider-aws/pkg/clients/servicediscovery"
	"github.com/crossplane/provider-aws/pkg/clients/servicediscovery/fake"
)

const (
	validOpID             string = "123"
	validNSID             string = "ns-id"
	validDescription      string = "valid description"
	validCreatorRequestID string = "valid:creator:request:id"
	validArn              string = "arn:string"
)

type args struct {
	client svcclient.Client
	kube   client.Client
	cr     *svcapitypes.PublicDNSNamespace
}

func TestObserve(t *testing.T) {
	type want struct {
		cr     *svcapitypes.PublicDNSNamespace
		result managed.ExternalObservation
		err    error
	}

	cases := map[string]struct {
		args
		want
	}{
		"NewNoOpID": {
			args: args{
				client: &fake.MockServicediscoveryClient{},
				kube:   nil,
				cr: &svcapitypes.PublicDNSNamespace{
					Spec: svcapitypes.PublicDNSNamespaceSpec{
						ForProvider: svcapitypes.PublicDNSNamespaceParameters{
							Region: "eu-central-1",
							Name:   aws.String("test"),
						},
					},
				},
			},
			want: want{
				cr: &svcapitypes.PublicDNSNamespace{
					Spec: svcapitypes.PublicDNSNamespaceSpec{
						ForProvider: svcapitypes.PublicDNSNamespaceParameters{
							Region: "eu-central-1",
							Name:   aws.String("test"),
						},
					},
					Status: svcapitypes.PublicDNSNamespaceStatus{},
				},
				result: managed.ExternalObservation{
					ResourceExists: false,
				},
			},
		},
		"NewOperationSubmitted": {
			args: args{
				client: &fake.MockServicediscoveryClient{
					MockGetOperation: func(input *svcsdk.GetOperationInput) (*svcsdk.GetOperationOutput, error) {
						if awsclient.StringValue(input.OperationId) != validOpID {
							return &svcsdk.GetOperationOutput{}, nil
						}
						return &svcsdk.GetOperationOutput{
							Operation: &svcsdk.Operation{
								Status:  aws.String("SUBMITTED"),
								Targets: map[string]*string{"NAMESPACE": aws.String(validNSID)},
							},
						}, nil
					},
				},
				kube: nil,
				cr: &svcapitypes.PublicDNSNamespace{
					Spec: svcapitypes.PublicDNSNamespaceSpec{
						ForProvider: svcapitypes.PublicDNSNamespaceParameters{
							Region: "eu-central-1",
							Name:   aws.String("test"),
						},
					},
					Status: svcapitypes.PublicDNSNamespaceStatus{
						AtProvider: svcapitypes.PublicDNSNamespaceObservation{
							OperationID: aws.String(validOpID),
						},
					},
				},
			},
			want: want{
				cr: &svcapitypes.PublicDNSNamespace{
					Spec: svcapitypes.PublicDNSNamespaceSpec{
						ForProvider: svcapitypes.PublicDNSNamespaceParameters{
							Region: "eu-central-1",
							Name:   aws.String("test"),
						},
					},
					Status: svcapitypes.PublicDNSNamespaceStatus{
						AtProvider: svcapitypes.PublicDNSNamespaceObservation{
							OperationID: aws.String(validOpID),
						},
					},
				},
				result: managed.ExternalObservation{
					ResourceExists: true,
				},
			},
		},
		"NewOperationPending": {
			args: args{
				client: &fake.MockServicediscoveryClient{
					MockGetOperation: func(input *svcsdk.GetOperationInput) (*svcsdk.GetOperationOutput, error) {
						if awsclient.StringValue(input.OperationId) != validOpID {
							return &svcsdk.GetOperationOutput{}, nil
						}
						return &svcsdk.GetOperationOutput{
							Operation: &svcsdk.Operation{
								Status:  aws.String("PENDING"),
								Targets: map[string]*string{"NAMESPACE": aws.String(validNSID)},
							},
						}, nil
					},
				},
				kube: nil,
				cr: &svcapitypes.PublicDNSNamespace{
					Spec: svcapitypes.PublicDNSNamespaceSpec{
						ForProvider: svcapitypes.PublicDNSNamespaceParameters{
							Region: "eu-central-1",
							Name:   aws.String("test"),
						},
					},
					Status: svcapitypes.PublicDNSNamespaceStatus{
						AtProvider: svcapitypes.PublicDNSNamespaceObservation{
							OperationID: aws.String(validOpID),
						},
					},
				},
			},
			want: want{
				cr: &svcapitypes.PublicDNSNamespace{
					Spec: svcapitypes.PublicDNSNamespaceSpec{
						ForProvider: svcapitypes.PublicDNSNamespaceParameters{
							Region: "eu-central-1",
							Name:   aws.String("test"),
						},
					},
					Status: svcapitypes.PublicDNSNamespaceStatus{
						AtProvider: svcapitypes.PublicDNSNamespaceObservation{
							OperationID: aws.String(validOpID),
						},
					},
				},
				result: managed.ExternalObservation{
					ResourceExists: true,
				},
			},
		},
		"NewOperationFailed": {
			args: args{
				client: &fake.MockServicediscoveryClient{
					MockGetOperation: func(input *svcsdk.GetOperationInput) (*svcsdk.GetOperationOutput, error) {
						if awsclient.StringValue(input.OperationId) != validOpID {
							return &svcsdk.GetOperationOutput{}, nil
						}
						return &svcsdk.GetOperationOutput{
							Operation: &svcsdk.Operation{
								Status:  aws.String("FAIL"),
								Targets: map[string]*string{"NAMESPACE": aws.String(validNSID)},
							},
						}, nil
					},
				},
				kube: nil,
				cr: &svcapitypes.PublicDNSNamespace{
					Spec: svcapitypes.PublicDNSNamespaceSpec{
						ForProvider: svcapitypes.PublicDNSNamespaceParameters{
							Region: "eu-central-1",
							Name:   aws.String("test"),
						},
					},
					Status: svcapitypes.PublicDNSNamespaceStatus{
						AtProvider: svcapitypes.PublicDNSNamespaceObservation{
							OperationID: aws.String(validOpID),
						},
					},
				},
			},
			want: want{
				cr: &svcapitypes.PublicDNSNamespace{
					Spec: svcapitypes.PublicDNSNamespaceSpec{
						ForProvider: svcapitypes.PublicDNSNamespaceParameters{
							Region: "eu-central-1",
							Name:   aws.String("test"),
						},
					},
					Status: svcapitypes.PublicDNSNamespaceStatus{
						AtProvider: svcapitypes.PublicDNSNamespaceObservation{
							OperationID: aws.String(validOpID),
						},
						ResourceStatus: xpv1.ResourceStatus{
							ConditionedStatus: xpv1.ConditionedStatus{
								Conditions: []xpv1.Condition{xpv1.Unavailable()},
							},
						},
					},
				},
				result: managed.ExternalObservation{
					ResourceExists: true,
				},
			},
		},
		"DeletingOperationFail": {
			args: args{
				client: &fake.MockServicediscoveryClient{
					MockGetOperation: func(input *svcsdk.GetOperationInput) (*svcsdk.GetOperationOutput, error) {
						if awsclient.StringValue(input.OperationId) != validOpID {
							return &svcsdk.GetOperationOutput{}, nil
						}
						return &svcsdk.GetOperationOutput{
							Operation: &svcsdk.Operation{
								Status:  aws.String("FAIL"),
								Targets: map[string]*string{"NAMESPACE": aws.String(validNSID)},
							},
						}, nil
					},
				},
				kube: nil,
				cr: &svcapitypes.PublicDNSNamespace{
					Spec: svcapitypes.PublicDNSNamespaceSpec{
						ForProvider: svcapitypes.PublicDNSNamespaceParameters{
							Region: "eu-central-1",
							Name:   aws.String("test"),
						},
					},
					Status: svcapitypes.PublicDNSNamespaceStatus{
						AtProvider: svcapitypes.PublicDNSNamespaceObservation{
							OperationID: aws.String(validOpID),
						},
						ResourceStatus: xpv1.ResourceStatus{
							ConditionedStatus: xpv1.ConditionedStatus{
								Conditions: []xpv1.Condition{xpv1.Deleting()},
							},
						},
					},
				},
			},
			want: want{
				cr: &svcapitypes.PublicDNSNamespace{
					Spec: svcapitypes.PublicDNSNamespaceSpec{
						ForProvider: svcapitypes.PublicDNSNamespaceParameters{
							Region: "eu-central-1",
							Name:   aws.String("test"),
						},
					},
					Status: svcapitypes.PublicDNSNamespaceStatus{
						AtProvider: svcapitypes.PublicDNSNamespaceObservation{
							OperationID: aws.String(validOpID),
						},
						ResourceStatus: xpv1.ResourceStatus{
							ConditionedStatus: xpv1.ConditionedStatus{
								Conditions: []xpv1.Condition{xpv1.Unavailable()},
							},
						},
					},
				},
				result: managed.ExternalObservation{
					ResourceExists: false,
				},
			},
		},
		"NewOpDoneNSNotFound": {
			args: args{
				client: &fake.MockServicediscoveryClient{
					MockGetOperation: func(input *svcsdk.GetOperationInput) (*svcsdk.GetOperationOutput, error) {
						if awsclient.StringValue(input.OperationId) != validOpID {
							return &svcsdk.GetOperationOutput{}, nil
						}
						return &svcsdk.GetOperationOutput{
							Operation: &svcsdk.Operation{
								Status:  aws.String("SUCCESS"),
								Targets: map[string]*string{"NAMESPACE": aws.String(validNSID)},
							},
						}, nil
					},
					MockGetNamespace: func(input *svcsdk.GetNamespaceInput) (*svcsdk.GetNamespaceOutput, error) {
						return &svcsdk.GetNamespaceOutput{}, awserr.New(svcsdk.ErrCodeNamespaceNotFound, "namespace not found", fmt.Errorf("err"))
					},
				},
				kube: nil,
				cr: &svcapitypes.PublicDNSNamespace{
					Spec: svcapitypes.PublicDNSNamespaceSpec{
						ForProvider: svcapitypes.PublicDNSNamespaceParameters{
							Region: "eu-central-1",
							Name:   aws.String("test"),
						},
					},
					Status: svcapitypes.PublicDNSNamespaceStatus{
						AtProvider: svcapitypes.PublicDNSNamespaceObservation{
							OperationID: aws.String(validOpID),
						},
					},
				},
			},
			want: want{
				cr: &svcapitypes.PublicDNSNamespace{
					Spec: svcapitypes.PublicDNSNamespaceSpec{
						ForProvider: svcapitypes.PublicDNSNamespaceParameters{
							Region: "eu-central-1",
							Name:   aws.String("test"),
						},
					},
					Status: svcapitypes.PublicDNSNamespaceStatus{
						AtProvider: svcapitypes.PublicDNSNamespaceObservation{
							OperationID: aws.String(validOpID),
						},
						ResourceStatus: xpv1.ResourceStatus{
							ConditionedStatus: xpv1.ConditionedStatus{
								Conditions: []xpv1.Condition{xpv1.Unavailable()},
							},
						},
					},
				},
				result: managed.ExternalObservation{
					ResourceExists: false,
				},
			},
		},
		"NewExternalNameNotSet": {
			args: args{
				client: &fake.MockServicediscoveryClient{
					MockGetOperation: func(input *svcsdk.GetOperationInput) (*svcsdk.GetOperationOutput, error) {
						if awsclient.StringValue(input.OperationId) != validOpID {
							return &svcsdk.GetOperationOutput{}, nil
						}
						return &svcsdk.GetOperationOutput{
							Operation: &svcsdk.Operation{
								Status:  aws.String("SUCCESS"),
								Targets: map[string]*string{"NAMESPACE": aws.String(validNSID)},
							},
						}, nil
					},
					MockGetNamespace: func(input *svcsdk.GetNamespaceInput) (*svcsdk.GetNamespaceOutput, error) {
						if awsclient.StringValue(input.Id) != validNSID {
							return &svcsdk.GetNamespaceOutput{}, nil
						}
						return &svcsdk.GetNamespaceOutput{
							Namespace: &svcsdk.Namespace{
								Arn:              aws.String(validArn),
								Name:             aws.String(validNSID),
								Description:      aws.String(validDescription),
								CreatorRequestId: aws.String(validCreatorRequestID),
							},
						}, nil
					},
				},
				kube: nil,
				cr: &svcapitypes.PublicDNSNamespace{
					Spec: svcapitypes.PublicDNSNamespaceSpec{
						ForProvider: svcapitypes.PublicDNSNamespaceParameters{
							Region:      "eu-central-1",
							Name:        aws.String("test"),
							Description: aws.String(validDescription),
						},
					},
					Status: svcapitypes.PublicDNSNamespaceStatus{
						AtProvider: svcapitypes.PublicDNSNamespaceObservation{
							OperationID: aws.String(validOpID),
						},
					},
				},
			},
			want: want{
				cr: &svcapitypes.PublicDNSNamespace{
					ObjectMeta: v1.ObjectMeta{
						Annotations: map[string]string{"crossplane.io/external-name": validNSID},
					},
					Spec: svcapitypes.PublicDNSNamespaceSpec{
						ForProvider: svcapitypes.PublicDNSNamespaceParameters{
							Region:           "eu-central-1",
							Name:             aws.String("test"),
							Description:      aws.String(validDescription),
							CreatorRequestID: aws.String(validCreatorRequestID),
						},
					},
					Status: svcapitypes.PublicDNSNamespaceStatus{
						AtProvider: svcapitypes.PublicDNSNamespaceObservation{
							OperationID: aws.String(validOpID),
						},
						ResourceStatus: xpv1.ResourceStatus{
							ConditionedStatus: xpv1.ConditionedStatus{
								Conditions: []xpv1.Condition{xpv1.Available()},
							},
						},
					},
				},
				result: managed.ExternalObservation{
					ResourceExists:          true,
					ResourceLateInitialized: true,
				},
			},
		},
		"NewWithExternalName": {
			args: args{
				client: &fake.MockServicediscoveryClient{
					MockGetOperation: func(input *svcsdk.GetOperationInput) (*svcsdk.GetOperationOutput, error) {
						if awsclient.StringValue(input.OperationId) != validOpID {
							return &svcsdk.GetOperationOutput{}, nil
						}
						return &svcsdk.GetOperationOutput{
							Operation: &svcsdk.Operation{
								Status:  aws.String("SUCCESS"),
								Targets: map[string]*string{"NAMESPACE": aws.String(validNSID)},
							},
						}, nil
					},
					MockGetNamespace: func(input *svcsdk.GetNamespaceInput) (*svcsdk.GetNamespaceOutput, error) {
						if awsclient.StringValue(input.Id) != validNSID {
							return &svcsdk.GetNamespaceOutput{}, nil
						}
						return &svcsdk.GetNamespaceOutput{
							Namespace: &svcsdk.Namespace{
								Arn:              aws.String(validArn),
								Name:             aws.String(validNSID),
								Description:      aws.String(validDescription),
								CreatorRequestId: aws.String(validCreatorRequestID),
							},
						}, nil
					},
				},
				kube: nil,
				cr: &svcapitypes.PublicDNSNamespace{
					ObjectMeta: v1.ObjectMeta{
						Annotations: map[string]string{"crossplane.io/external-name": validNSID},
					},
					Spec: svcapitypes.PublicDNSNamespaceSpec{
						ForProvider: svcapitypes.PublicDNSNamespaceParameters{
							Region: "eu-central-1",
							Name:   aws.String("test"),
						},
					},
					Status: svcapitypes.PublicDNSNamespaceStatus{
						AtProvider: svcapitypes.PublicDNSNamespaceObservation{
							OperationID: aws.String(validOpID),
						},
					},
				},
			},
			want: want{
				cr: &svcapitypes.PublicDNSNamespace{
					ObjectMeta: v1.ObjectMeta{
						Annotations: map[string]string{"crossplane.io/external-name": validNSID},
					},
					Spec: svcapitypes.PublicDNSNamespaceSpec{
						ForProvider: svcapitypes.PublicDNSNamespaceParameters{
							Region:           "eu-central-1",
							Name:             aws.String("test"),
							Description:      aws.String(validDescription),
							CreatorRequestID: aws.String(validCreatorRequestID),
						},
					},
					Status: svcapitypes.PublicDNSNamespaceStatus{
						AtProvider: svcapitypes.PublicDNSNamespaceObservation{
							OperationID: aws.String(validOpID),
						},
						ResourceStatus: xpv1.ResourceStatus{
							ConditionedStatus: xpv1.ConditionedStatus{
								Conditions: []xpv1.Condition{xpv1.Available()},
							},
						},
					},
				},
				result: managed.ExternalObservation{
					ResourceExists:          true,
					ResourceLateInitialized: true,
				},
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &external{kube: tc.kube, client: tc.client}
			useHooks(e)

			o, err := e.Observe(context.Background(), tc.args.cr)

			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
			if diff := cmp.Diff(tc.want.cr, tc.args.cr, test.EquateConditions(),
				cmpopts.IgnoreFields(v1.Condition{}, "LastTransitionTime"),
			); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
			if diff := cmp.Diff(tc.want.result, o); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
		})
	}
}

func TestCreate(t *testing.T) {
	type want struct {
		cr     *svcapitypes.PublicDNSNamespace
		result managed.ExternalCreation
		err    error
	}

	cases := map[string]struct {
		args
		want
	}{
		"NewPublicDNSNamespace": {
			args: args{
				client: &fake.MockServicediscoveryClient{
					MockCreatePublicDNSNamespace: func(input *svcsdk.CreatePublicDnsNamespaceInput) (*svcsdk.CreatePublicDnsNamespaceOutput, error) {
						return &svcsdk.CreatePublicDnsNamespaceOutput{
							OperationId: aws.String(validOpID),
						}, nil
					},
				},
				kube: nil,
				cr: &svcapitypes.PublicDNSNamespace{
					Spec: svcapitypes.PublicDNSNamespaceSpec{
						ForProvider: svcapitypes.PublicDNSNamespaceParameters{
							Region: "eu-central-1",
							Name:   aws.String("test"),
							Tags:   []*svcapitypes.Tag{{Key: aws.String("key"), Value: aws.String("value")}},
						},
					},
				},
			},
			want: want{
				cr: &svcapitypes.PublicDNSNamespace{
					Spec: svcapitypes.PublicDNSNamespaceSpec{
						ForProvider: svcapitypes.PublicDNSNamespaceParameters{
							Region: "eu-central-1",
							Name:   aws.String("test"),
							Tags:   []*svcapitypes.Tag{{Key: aws.String("key"), Value: aws.String("value")}},
						},
					},
					Status: svcapitypes.PublicDNSNamespaceStatus{
						AtProvider: svcapitypes.PublicDNSNamespaceObservation{
							OperationID: aws.String(validOpID),
						},
						ResourceStatus: xpv1.ResourceStatus{
							ConditionedStatus: xpv1.ConditionedStatus{
								Conditions: []xpv1.Condition{xpv1.Creating()},
							},
						},
					},
				},
				result: managed.ExternalCreation{},
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &external{kube: tc.kube, client: tc.client}
			useHooks(e)

			o, err := e.Create(context.Background(), tc.args.cr)

			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
			if diff := cmp.Diff(tc.want.cr, tc.args.cr, test.EquateConditions(),
				cmpopts.IgnoreFields(v1.Condition{}, "LastTransitionTime"),
			); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
			if diff := cmp.Diff(tc.want.result, o); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
		})
	}
}

func TestDelete(t *testing.T) {
	type want struct {
		cr  *svcapitypes.PublicDNSNamespace
		err error
	}

	cases := map[string]struct {
		args
		want
	}{
		"NewPublicDNSNamespace": {
			args: args{
				client: &fake.MockServicediscoveryClient{
					MockDeleteNamespace: func(input *svcsdk.DeleteNamespaceInput) (*svcsdk.DeleteNamespaceOutput, error) {
						return &svcsdk.DeleteNamespaceOutput{
							OperationId: aws.String(validOpID),
						}, nil
					},
				},
				kube: nil,
				cr: &svcapitypes.PublicDNSNamespace{
					Spec: svcapitypes.PublicDNSNamespaceSpec{
						ForProvider: svcapitypes.PublicDNSNamespaceParameters{
							Region: "eu-central-1",
							Name:   aws.String("test"),
						},
					},
					Status: svcapitypes.PublicDNSNamespaceStatus{
						AtProvider: svcapitypes.PublicDNSNamespaceObservation{
							OperationID: aws.String(validOpID),
						},
					},
				},
			},
			want: want{
				cr: &svcapitypes.PublicDNSNamespace{
					Spec: svcapitypes.PublicDNSNamespaceSpec{
						ForProvider: svcapitypes.PublicDNSNamespaceParameters{
							Region: "eu-central-1",
							Name:   aws.String("test"),
						},
					},
					Status: svcapitypes.PublicDNSNamespaceStatus{
						AtProvider: svcapitypes.PublicDNSNamespaceObservation{
							OperationID: aws.String(validOpID),
						},
						ResourceStatus: xpv1.ResourceStatus{
							ConditionedStatus: xpv1.ConditionedStatus{
								Conditions: []xpv1.Condition{xpv1.Deleting()},
							},
						},
					},
				},
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &external{kube: tc.kube, client: tc.client}
			useHooks(e)

			err := e.Delete(context.Background(), tc.args.cr)

			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
			if diff := cmp.Diff(tc.want.cr, tc.args.cr, test.EquateConditions(),
				cmpopts.IgnoreFields(v1.Condition{}, "LastTransitionTime"),
			); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
		})
	}
}
