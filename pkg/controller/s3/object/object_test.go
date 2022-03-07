package object

import (
	"bytes"
	"context"
	"io/ioutil"
	"testing"

	awsclient "github.com/crossplane/provider-aws/pkg/clients"
	"github.com/crossplane/provider-aws/pkg/clients/s3"
	"github.com/crossplane/provider-aws/pkg/clients/s3/fake"
	s3Testing "github.com/crossplane/provider-aws/pkg/controller/s3/testing"

	awss3 "github.com/aws/aws-sdk-go-v2/service/s3"
	s3types "github.com/aws/aws-sdk-go-v2/service/s3/types"
	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	"github.com/crossplane/crossplane-runtime/pkg/test"
	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var (
	// an arbitrary managed resource
	unexpectedItem resource.Managed
	errBoom        = errors.New("boom")
)

type args struct {
	kube   client.Client
	client s3.ObjectClient
	cr     resource.Managed
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
		"InValidInput": {
			args: args{
				cr: unexpectedItem,
			},
			want: want{
				cr:  unexpectedItem,
				err: errors.New(errUnexpectedObject),
			},
		},
		"ClientError": {
			args: args{
				client: &fake.MockObjectClient{
					MockGetObject: func(ctx context.Context, input *awss3.GetObjectInput, opts []func(*awss3.Options)) (*awss3.GetObjectOutput, error) {
						return nil, errBoom
					},
				},
				cr: s3Testing.Object("key", "", xpv1.ConditionReason("UnKnow")),
			},
			want: want{
				cr:  s3Testing.Object("key", "", xpv1.ConditionReason("UnKnow")),
				err: awsclient.Wrap(errBoom, errGet),
			},
		},
		"BucketDoesNotExist": {
			args: args{
				client: &fake.MockObjectClient{
					MockGetObject: func(ctx context.Context, input *awss3.GetObjectInput, opts []func(*awss3.Options)) (*awss3.GetObjectOutput, error) {
						return nil, &s3types.NoSuchBucket{}
					},
				},
				cr: s3Testing.Object("key", "", xpv1.ConditionReason("UnKnow")),
			},
			want: want{
				cr: s3Testing.Object("key", "", xpv1.ConditionReason("UnKnow")),
			},
		},
		"ObjectDoesNotExist": {
			args: args{
				client: &fake.MockObjectClient{
					MockGetObject: func(ctx context.Context, input *awss3.GetObjectInput, opts []func(*awss3.Options)) (*awss3.GetObjectOutput, error) {
						return nil, &s3types.NoSuchKey{}
					},
				},
				cr: s3Testing.Object("key", "", xpv1.ConditionReason("UnKnow")),
			},
			want: want{
				cr: s3Testing.Object("key", "", xpv1.ConditionReason("UnKnow")),
			},
		},
		"ObjectCreateSuccess": {
			// this case is the same as needing an update, we should not late init here.
			args: args{
				client: &fake.MockObjectClient{
					MockGetObject: func(ctx context.Context, input *awss3.GetObjectInput, opts []func(*awss3.Options)) (*awss3.GetObjectOutput, error) {
						return &awss3.GetObjectOutput{
							Body: ioutil.NopCloser(bytes.NewReader([]byte("hello world"))),
						}, nil
					},
				},
				cr: s3Testing.Object("key", "hello world", xpv1.ConditionReason("UnKnow")),
			},
			want: want{
				cr: s3Testing.Object("key", "hello world", xpv1.ReasonAvailable),
				result: managed.ExternalObservation{
					ResourceExists:   true,
					ResourceUpToDate: true,
				},
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &external{client: tc.client, kube: tc.kube}
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
				kube: &test.MockClient{
					MockUpdate: test.NewMockUpdateFn(nil),
				},
				client: &fake.MockObjectClient{
					MockPutObject: func(ctx context.Context, input *awss3.PutObjectInput, opts []func(*awss3.Options)) (*awss3.PutObjectOutput, error) {
						return &awss3.PutObjectOutput{}, nil
					},
				},
				cr: s3Testing.Object("key", "", xpv1.ReasonCreating),
			},
			want: want{
				cr: s3Testing.Object("key", "", xpv1.ReasonCreating),
			},
		},
		"InValidInput": {
			args: args{
				kube: &test.MockClient{
					MockUpdate: test.NewMockUpdateFn(nil),
				},
				cr: unexpectedItem,
			},
			want: want{
				cr:  unexpectedItem,
				err: errors.New(errUnexpectedObject),
			},
		},
		"ClientError": {
			args: args{
				kube: &test.MockClient{
					MockUpdate: test.NewMockUpdateFn(nil),
				},
				client: &fake.MockObjectClient{
					MockPutObject: func(ctx context.Context, input *awss3.PutObjectInput, opts []func(*awss3.Options)) (*awss3.PutObjectOutput, error) {
						return &awss3.PutObjectOutput{}, errBoom
					},
				},
				cr: s3Testing.Object("key", "", xpv1.ReasonCreating),
			},
			want: want{
				cr:  s3Testing.Object("key", "", xpv1.ReasonCreating),
				err: awsclient.Wrap(errBoom, errCreate),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &external{client: tc.client, kube: tc.kube}
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
		"InValidInput": {
			args: args{
				cr: unexpectedItem,
			},
			want: want{
				cr:  unexpectedItem,
				err: errors.New(errUnexpectedObject),
			},
		},
		"ValidInput": {
			args: args{
				kube: &test.MockClient{
					MockUpdate: test.NewMockUpdateFn(nil),
				},
				client: &fake.MockObjectClient{
					MockPutObject: func(ctx context.Context, input *awss3.PutObjectInput, opts []func(*awss3.Options)) (*awss3.PutObjectOutput, error) {
						return &awss3.PutObjectOutput{}, errBoom
					},
				},
				cr: s3Testing.Object("key", "", xpv1.ReasonCreating),
			},
			want: want{
				cr:  s3Testing.Object("key", "", xpv1.ReasonCreating),
				err: awsclient.Wrap(errBoom, errUpdate),
			},
		},
		"ClientError": {
			args: args{
				kube: &test.MockClient{
					MockUpdate: test.NewMockUpdateFn(nil),
				},
				client: &fake.MockObjectClient{
					MockPutObject: func(ctx context.Context, input *awss3.PutObjectInput, opts []func(*awss3.Options)) (*awss3.PutObjectOutput, error) {
						return &awss3.PutObjectOutput{}, errBoom
					},
				},
				cr: s3Testing.Object("key", "", xpv1.ReasonCreating),
			},
			want: want{
				cr:  s3Testing.Object("key", "", xpv1.ReasonCreating),
				err: awsclient.Wrap(errBoom, errUpdate),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &external{client: tc.client, kube: tc.kube}
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
				client: &fake.MockObjectClient{
					MockDeleteObjects: func(ctx context.Context, input *awss3.DeleteObjectsInput, opts []func(*awss3.Options)) (*awss3.DeleteObjectsOutput, error) {
						return &awss3.DeleteObjectsOutput{}, nil
					},
					MockListObjectVersions: func(ctx context.Context, input *awss3.ListObjectVersionsInput, opts []func(*awss3.Options)) (*awss3.ListObjectVersionsOutput, error) {
						return &awss3.ListObjectVersionsOutput{}, nil
					},
				},
				cr: s3Testing.Object("key", "", xpv1.ConditionReason("UnKnow")),
			},
			want: want{
				cr: s3Testing.Object("key", "", xpv1.ReasonDeleting),
			},
		},
		"InValidInput": {
			args: args{
				cr: unexpectedItem,
			},
			want: want{
				cr:  unexpectedItem,
				err: errors.New(errUnexpectedObject),
			},
		},
		"ClientErrorWhenDelete": {
			args: args{
				client: &fake.MockObjectClient{
					MockDeleteObjects: func(ctx context.Context, input *awss3.DeleteObjectsInput, opts []func(*awss3.Options)) (*awss3.DeleteObjectsOutput, error) {
						return nil, errBoom
					},
					MockListObjectVersions: func(ctx context.Context, input *awss3.ListObjectVersionsInput, opts []func(*awss3.Options)) (*awss3.ListObjectVersionsOutput, error) {
						return &awss3.ListObjectVersionsOutput{}, nil
					},
				},
				cr: s3Testing.Object("key", "", xpv1.ConditionReason("UnKnow")),
			},
			want: want{
				cr:  s3Testing.Object("key", "", xpv1.ReasonDeleting),
				err: awsclient.Wrap(errBoom, errDelete),
			},
		},
		"ClientErrorWhenList": {
			args: args{
				client: &fake.MockObjectClient{
					MockDeleteObjects: func(ctx context.Context, input *awss3.DeleteObjectsInput, opts []func(*awss3.Options)) (*awss3.DeleteObjectsOutput, error) {
						return &awss3.DeleteObjectsOutput{}, nil
					},
					MockListObjectVersions: func(ctx context.Context, input *awss3.ListObjectVersionsInput, opts []func(*awss3.Options)) (*awss3.ListObjectVersionsOutput, error) {
						return nil, errBoom
					},
				},
				cr: s3Testing.Object("key", "", xpv1.ConditionReason("UnKnow")),
			},
			want: want{
				cr:  s3Testing.Object("key", "", xpv1.ReasonDeleting),
				err: awsclient.Wrap(errBoom, errDelete),
			},
		},
		"BucketDoesNotExist": {
			args: args{
				client: &fake.MockObjectClient{
					MockListObjectVersions: func(ctx context.Context, input *awss3.ListObjectVersionsInput, opts []func(*awss3.Options)) (*awss3.ListObjectVersionsOutput, error) {
						return nil, &s3types.NoSuchBucket{}
					},
				},
				cr: s3Testing.Object("key", "", xpv1.ConditionReason("UnKnow")),
			},
			want: want{
				cr:  s3Testing.Object("key", "", xpv1.ReasonDeleting),
				err: awsclient.Wrap(nil, errDelete),
			},
		},
		"ObjectDoesNotExist": {
			args: args{
				client: &fake.MockObjectClient{
					MockListObjectVersions: func(ctx context.Context, input *awss3.ListObjectVersionsInput, opts []func(*awss3.Options)) (*awss3.ListObjectVersionsOutput, error) {
						return nil, &s3types.NoSuchKey{}
					},
				},
				cr: s3Testing.Object("key", "", xpv1.ConditionReason("UnKnow")),
			},
			want: want{
				cr:  s3Testing.Object("key", "", xpv1.ReasonDeleting),
				err: awsclient.Wrap(nil, errDelete),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &external{client: tc.client, kube: tc.kube}
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
