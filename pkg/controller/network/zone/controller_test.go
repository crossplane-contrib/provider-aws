package zone

import (
	"context"
	"net/http"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsroute53 "github.com/aws/aws-sdk-go-v2/service/route53"
	"github.com/aws/aws-sdk-go/aws/awserr"
	runtimev1alpha1 "github.com/crossplane/crossplane-runtime/apis/core/v1alpha1"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	"github.com/crossplane/crossplane-runtime/pkg/test"
	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/crossplane/provider-aws/apis/network/v1alpha3"
	"github.com/crossplane/provider-aws/pkg/clients/zone"
	"github.com/crossplane/provider-aws/pkg/clients/zone/fake"
)

const (
	providerName = "aws-creds"
	testRegion   = "us-east-1"
)

var (
	unexpecedItem resource.Managed
	zoneName            = "crossplane.io"
	uuid                = "a96abeca-8da3-40fc-a2d5-08d72084eb65"
	errBoom             = errors.New("Some random error")
	id                  = "/hostedzone/XXXXXXXXXXXXXXXXXXX"
	location            = "https://route53.amazonaws.com/2013-04-01/hostedzone/XXXXXXXXXXXXXXXXXXX"
	rrCount       int64 = 2
	c                   = new(string)
	b                   = new(bool)
)

type zoneModifier func(*v1alpha3.Zone)

type args struct {
	kube    client.Client
	route53 zone.Client
	cr      resource.Managed
}

func withZoneName(s *string) zoneModifier {
	return func(r *v1alpha3.Zone) { meta.SetExternalName(r, *s) }
}

func withConditions(c ...runtimev1alpha1.Condition) zoneModifier {
	return func(r *v1alpha3.Zone) { r.Status.ConditionedStatus.Conditions = c }
}

func withStatus(id, location string, rr int64) zoneModifier {
	return func(r *v1alpha3.Zone) {
		r.Status.AtProvider.ID = id
		r.Status.AtProvider.Location = location
		r.Status.AtProvider.ResourceRecordCount = rr
	}
}

func withComment(c string) zoneModifier {
	return func(r *v1alpha3.Zone) { r.Spec.ForProvider.Comment = &c }
}

func zoneTester(m ...zoneModifier) *v1alpha3.Zone {
	cr := &v1alpha3.Zone{
		Spec: v1alpha3.ZoneSpec{
			ResourceSpec: runtimev1alpha1.ResourceSpec{
				ProviderReference: &corev1.ObjectReference{Name: providerName},
			},
			ForProvider: v1alpha3.ZoneParameters{
				Comment:         c,
				CallerReference: &uuid,
				Name:            &zoneName,
				PrivateZone:     b,
			},
		},
	}
	for _, f := range m {
		f(cr)
	}
	return cr
}

func TestConnect(t *testing.T) {

	type args struct {
		cr          resource.Managed
		newClientFn func(*aws.Config) zone.Client
		awsConfigFn func(context.Context, client.Reader, *corev1.ObjectReference) (*aws.Config, error)
	}
	type want struct {
		err error
	}

	cases := map[string]struct {
		args
		want
	}{
		"ValidInput": {
			args: args{
				newClientFn: func(config *aws.Config) zone.Client {
					if diff := cmp.Diff(testRegion, config.Region); diff != "" {
						t.Errorf("r: -want, +got:\n%s", diff)
					}
					return nil
				},
				awsConfigFn: func(_ context.Context, _ client.Reader, p *corev1.ObjectReference) (*aws.Config, error) {
					if diff := cmp.Diff(providerName, p.Name); diff != "" {
						t.Errorf("r: -want, +got:\n%s", diff)
					}
					return &aws.Config{Region: testRegion}, nil
				},
				cr: zoneTester(),
			},
			want: want{
				err: nil,
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			c := &connector{newClientFn: tc.newClientFn, awsConfigFn: tc.awsConfigFn}
			_, err := c.Connect(context.Background(), tc.args.cr)
			if diff := cmp.Diff(tc.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
		})
	}
}

func TestObserve(t *testing.T) {

	type want struct {
		cr     resource.Managed
		result managed.ExternalObservation
		err    error
	}

	cases := map[string]struct {
		args
		want
	}{
		"VaildInput": {
			args: args{
				kube: &test.MockClient{
					MockStatusUpdate: test.NewMockStatusUpdateFn(nil),
				},
				route53: &fake.MockZoneClient{
					MockGetZoneRequest: func(input *string) awsroute53.GetHostedZoneRequest {
						return awsroute53.GetHostedZoneRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{},
								Data: &awsroute53.GetHostedZoneOutput{
									DelegationSet: &awsroute53.DelegationSet{},
									HostedZone: &awsroute53.HostedZone{
										CallerReference:        &uuid,
										Id:                     &id,
										ResourceRecordSetCount: &rrCount,
										Config: &awsroute53.HostedZoneConfig{
											Comment:     c,
											PrivateZone: b,
										},
									},
									VPCs: make([]awsroute53.VPC, 0),
								}},
						}
					},
				},
				cr: zoneTester(
					withZoneName(&zoneName),
					withStatus(id, location, rrCount)),
			},
			want: want{
				cr: zoneTester(
					withZoneName(&zoneName),
					withStatus(id, location, rrCount),
					withConditions(runtimev1alpha1.Available())),
				result: managed.ExternalObservation{
					ResourceExists:    true,
					ResourceUpToDate:  true,
					ConnectionDetails: managed.ConnectionDetails{},
				},
			},
		},
		"InValidInput": {
			args: args{
				cr: unexpecedItem,
			},
			want: want{
				cr:  unexpecedItem,
				err: errors.New(errUnexpectedObject),
			},
		},
		"ResourceDoesNotExist": {
			args: args{
				route53: &fake.MockZoneClient{
					MockGetZoneRequest: func(input *string) awsroute53.GetHostedZoneRequest {
						return awsroute53.GetHostedZoneRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Error: awserr.New(awsroute53.ErrCodeNoSuchHostedZone, "", nil)},
						}
					},
				},
				cr: zoneTester(),
			},
			want: want{
				cr: zoneTester(),
				result: managed.ExternalObservation{
					ConnectionDetails: managed.ConnectionDetails{},
				},
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &external{kube: test.NewMockClient(), client: tc.route53}
			o, err := e.Observe(context.Background(), tc.args.cr)

			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
			if diff := cmp.Diff(tc.want.cr, tc.args.cr, test.EquateConditions()); diff != "" {
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
		cr     resource.Managed
		result managed.ExternalCreation
		err    error
	}

	cases := map[string]struct {
		args
		want
	}{
		"VaildInput": {
			args: args{
				route53: &fake.MockZoneClient{
					MockCreateZoneRequest: func(cr *v1alpha3.Zone) awsroute53.CreateHostedZoneRequest {
						return awsroute53.CreateHostedZoneRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{},
								Data: &awsroute53.CreateHostedZoneOutput{
									HostedZone: &awsroute53.HostedZone{
										CallerReference:        &uuid,
										Id:                     &id,
										ResourceRecordSetCount: &rrCount,
										Config: &awsroute53.HostedZoneConfig{
											Comment:     c,
											PrivateZone: b,
										},
									},
									Location: &location,
								},
							},
						}
					},
				},
				cr: zoneTester(withZoneName(&zoneName)),
			},
			want: want{
				cr: zoneTester(
					withZoneName(&zoneName),
					withStatus(id, location, rrCount),
					withConditions(runtimev1alpha1.Creating())),
			},
		},
		"InValidInput": {
			args: args{
				cr: unexpecedItem,
			},
			want: want{
				cr:  unexpecedItem,
				err: errors.New(errUnexpectedObject),
			},
		},
		"ClientError": {
			args: args{
				route53: &fake.MockZoneClient{
					MockCreateZoneRequest: func(cr *v1alpha3.Zone) awsroute53.CreateHostedZoneRequest {
						return awsroute53.CreateHostedZoneRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Error: errBoom},
						}
					},
				},
				cr: zoneTester(),
			},
			want: want{
				cr:  zoneTester(withConditions(runtimev1alpha1.Creating())),
				err: errors.Wrap(errBoom, errCreate),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &external{client: tc.route53}
			o, err := e.Create(context.Background(), tc.args.cr)

			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
			if diff := cmp.Diff(tc.want.cr, tc.args.cr, test.EquateConditions()); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
			if diff := cmp.Diff(tc.want.result, o); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
		})
	}
}

func TestUpdate(t *testing.T) {

	type want struct {
		cr     resource.Managed
		result managed.ExternalUpdate
		err    error
	}

	cases := map[string]struct {
		args
		want
	}{
		"VaildInput": {
			args: args{
				route53: &fake.MockZoneClient{
					MockUpdateZoneRequest: func(id, comment *string) awsroute53.UpdateHostedZoneCommentRequest {
						return awsroute53.UpdateHostedZoneCommentRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{},
								Data: &awsroute53.UpdateHostedZoneCommentOutput{
									HostedZone: &awsroute53.HostedZone{
										CallerReference:        &uuid,
										Id:                     id,
										ResourceRecordSetCount: &rrCount,
										Config: &awsroute53.HostedZoneConfig{
											Comment:     comment,
											PrivateZone: b,
										},
									},
								}},
						}
					},
				},
				cr: zoneTester(withZoneName(&zoneName),
					withComment("New Comment")),
			},
			want: want{
				cr: zoneTester(withZoneName(&zoneName),
					withComment("New Comment")),
			},
		},
		"InValidInput": {
			args: args{
				cr: unexpecedItem,
			},
			want: want{
				cr:  unexpecedItem,
				err: errors.New(errUnexpectedObject),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &external{client: tc.route53}
			o, err := e.Update(context.Background(), tc.args.cr)

			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
			if diff := cmp.Diff(tc.want.cr, tc.args.cr, test.EquateConditions()); diff != "" {
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
		cr  resource.Managed
		err error
	}

	cases := map[string]struct {
		args
		want
	}{
		"VaildInput": {
			args: args{
				route53: &fake.MockZoneClient{
					MockDeleteZoneRequest: func(id *string) awsroute53.DeleteHostedZoneRequest {
						return awsroute53.DeleteHostedZoneRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Data: &awsroute53.DeleteHostedZoneOutput{}},
						}
					},
				},
				cr: zoneTester(withZoneName(&zoneName)),
			},
			want: want{
				cr: zoneTester(withZoneName(&zoneName),
					withConditions(runtimev1alpha1.Deleting())),
			},
		},
		"InValidInput": {
			args: args{
				cr: unexpecedItem,
			},
			want: want{
				cr:  unexpecedItem,
				err: errors.New(errUnexpectedObject),
			},
		},
		"ClientError": {
			args: args{
				route53: &fake.MockZoneClient{
					MockDeleteZoneRequest: func(id *string) awsroute53.DeleteHostedZoneRequest {
						return awsroute53.DeleteHostedZoneRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Error: errBoom},
						}
					},
				},
				cr: zoneTester(),
			},
			want: want{
				cr:  zoneTester(withConditions(runtimev1alpha1.Deleting())),
				err: errors.Wrap(errBoom, errDelete),
			},
		},
		"ResourceDoesNotExist": {
			args: args{
				route53: &fake.MockZoneClient{
					MockDeleteZoneRequest: func(id *string) awsroute53.DeleteHostedZoneRequest {
						return awsroute53.DeleteHostedZoneRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Error: awserr.New(awsroute53.ErrCodeNoSuchHostedZone, "", nil)},
						}
					},
				},
				cr: zoneTester(),
			},
			want: want{
				cr: zoneTester(withConditions(runtimev1alpha1.Deleting())),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &external{client: tc.route53}
			err := e.Delete(context.Background(), tc.args.cr)

			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
			if diff := cmp.Diff(tc.want.cr, tc.args.cr, test.EquateConditions()); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
		})
	}
}
