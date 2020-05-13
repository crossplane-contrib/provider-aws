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
	"context"
	"errors"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/google/go-cmp/cmp"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	. "sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/crossplane/provider-aws/apis"
	"github.com/crossplane/provider-aws/apis/storage/v1alpha3"
	. "github.com/crossplane/provider-aws/apis/storage/v1alpha3"
	aws "github.com/crossplane/provider-aws/pkg/clients"
	client "github.com/crossplane/provider-aws/pkg/clients/s3"
	. "github.com/crossplane/provider-aws/pkg/clients/s3/fake"

	runtimev1alpha1 "github.com/crossplane/crossplane-runtime/apis/core/v1alpha1"
	"github.com/crossplane/crossplane-runtime/pkg/logging"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	"github.com/crossplane/crossplane-runtime/pkg/test"
	storagev1alpha1 "github.com/crossplane/crossplane/apis/storage/v1alpha1"
)

const (
	namespace  = "default"
	bucketName = "test-bucket"
)

var (
	key = types.NamespacedName{
		Name: bucketName,
	}
	request = reconcile.Request{
		NamespacedName: key,
	}
)

func init() {
	if err := apis.AddToScheme(scheme.Scheme); err != nil {
		panic(err)
	}
}

func testResource() *S3Bucket {
	perm := storagev1alpha1.ReadOnlyPermission
	testIAMUsername := "test-username"
	return &S3Bucket{
		ObjectMeta: metav1.ObjectMeta{
			Name: bucketName,
		},
		Spec: S3BucketSpec{
			ResourceSpec:       runtimev1alpha1.ResourceSpec{ProviderReference: &corev1.ObjectReference{}},
			S3BucketParameters: v1alpha3.S3BucketParameters{LocalPermission: &perm, IAMUsername: testIAMUsername},
		},
	}
}

// assertResource a helper function to check on cluster and its status
func assertResource(g *GomegaWithT, r *Reconciler, s runtimev1alpha1.ConditionedStatus) *S3Bucket {
	resource := &S3Bucket{}
	err := r.Get(ctx, key, resource)
	g.Expect(err).To(BeNil())
	g.Expect(cmp.Diff(s, resource.Status.ConditionedStatus, test.EquateConditions())).Should(BeZero())
	return resource
}

func TestSyncBucketError(t *testing.T) {
	g := NewGomegaWithT(t)

	assert := func(instance *S3Bucket, client client.Service, expectedResult reconcile.Result, expectedStatus runtimev1alpha1.ConditionedStatus) {
		r := &Reconciler{
			Client: NewFakeClient(instance),
			log:    logging.NewNopLogger(),
		}

		rs, err := r._sync(instance, client)

		g.Expect(err).NotTo(HaveOccurred())
		g.Expect(rs).To(Equal(expectedResult))
		assertResource(g, r, expectedStatus)
	}

	// error iam username not set
	testError := errors.New("username not set, .Status.IAMUsername")
	cl := &MockS3Client{}

	expectedStatus := runtimev1alpha1.ConditionedStatus{}
	expectedStatus.SetConditions(runtimev1alpha1.ReconcileError(testError))
	noUserResource := testResource()
	noUserResource.Spec.IAMUsername = ""
	assert(noUserResource, cl, resultRequeue, expectedStatus)

	// error get bucket info
	testError = errors.New("mock get bucket info err")
	cl.MockGetBucketInfo = func(username string, bucket *S3Bucket) (*client.Bucket, error) {
		return nil, testError
	}
	expectedStatus = runtimev1alpha1.ConditionedStatus{}
	expectedStatus.SetConditions(runtimev1alpha1.ReconcileError(testError))
	assert(testResource(), cl, resultRequeue, expectedStatus)

	//update versioning error
	cl.MockGetBucketInfo = func(username string, bucket *S3Bucket) (*client.Bucket, error) {
		return &client.Bucket{Versioning: true, UserPolicyVersion: "v1"}, nil
	}

	testError = errors.New("bucket-versioning-update-error")
	cl.MockUpdateVersioning = func(bucket *S3Bucket) error {
		return testError
	}
	expectedStatus = runtimev1alpha1.ConditionedStatus{}
	expectedStatus.SetConditions(runtimev1alpha1.ReconcileError(testError))
	assert(testResource(), cl, resultRequeue, expectedStatus)

	// update bucket acl error
	cl.MockGetBucketInfo = func(username string, bucket *S3Bucket) (*client.Bucket, error) {
		return &client.Bucket{Versioning: false, UserPolicyVersion: "v1"}, nil
	}

	testError = errors.New("bucket-acl-update-error")
	cl.MockUpdateBucketACL = func(bucket *S3Bucket) error {
		return testError
	}

	expectedStatus = runtimev1alpha1.ConditionedStatus{}
	expectedStatus.SetConditions(runtimev1alpha1.ReconcileError(testError))
	assert(testResource(), cl, resultRequeue, expectedStatus)

	cl.MockUpdateBucketACL = func(bucket *S3Bucket) error {
		return nil
	}

	// Update policy error
	perm := storagev1alpha1.WriteOnlyPermission
	bucketWithPolicyChanges := testResource()
	bucketWithPolicyChanges.Spec.LocalPermission = &perm
	bucketWithPolicyChanges.Status.LastUserPolicyVersion = 1
	bucketWithPolicyChanges.Status.LastLocalPermission = storagev1alpha1.ReadOnlyPermission

	testError = errors.New("policy-update-err")
	cl.MockUpdatePolicyDocument = func(username string, bucket *S3Bucket) (string, error) {
		return "", testError
	}
	expectedStatus = runtimev1alpha1.ConditionedStatus{}
	expectedStatus.SetConditions(runtimev1alpha1.ReconcileError(testError))
	assert(testResource(), cl, resultRequeue, expectedStatus)
}

func TestSyncBucket(t *testing.T) {
	g := NewGomegaWithT(t)
	tr := testResource()
	tr.Status.LastUserPolicyVersion = 1
	tr.Status.LastLocalPermission = storagev1alpha1.ReadOnlyPermission

	r := &Reconciler{
		Client: NewFakeClient(tr),
		log:    logging.NewNopLogger(),
	}
	//
	updateBucketACLCalled := false
	getBucketInfoCalled := false
	cl := &MockS3Client{
		MockUpdateBucketACL: func(bucket *S3Bucket) error {
			updateBucketACLCalled = true
			return nil
		},
		MockGetBucketInfo: func(username string, bucket *S3Bucket) (*client.Bucket, error) {
			getBucketInfoCalled = true
			return &client.Bucket{Versioning: false, UserPolicyVersion: "v1"}, nil
		},
	}

	expectedStatus := runtimev1alpha1.ConditionedStatus{}
	expectedStatus.SetConditions(runtimev1alpha1.ReconcileSuccess())
	rs, err := r._sync(tr, cl)
	g.Expect(rs).To(Equal(result))
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(updateBucketACLCalled).To(BeTrue())
	g.Expect(getBucketInfoCalled).To(BeTrue())
	assertResource(g, r, expectedStatus)
}

func TestDelete(t *testing.T) {
	g := NewGomegaWithT(t)

	tr := testResource()

	r := &Reconciler{
		Client: NewFakeClient(tr),
		log:    logging.NewNopLogger(),
	}

	cl := &MockS3Client{}

	// test delete w/ reclaim policy
	tr.Spec.ReclaimPolicy = runtimev1alpha1.ReclaimRetain
	expectedStatus := runtimev1alpha1.ConditionedStatus{}
	expectedStatus.SetConditions(runtimev1alpha1.Deleting(), runtimev1alpha1.ReconcileSuccess())

	rs, err := r._delete(tr, cl)
	g.Expect(rs).To(Equal(result))
	g.Expect(err).NotTo(HaveOccurred())
	assertResource(g, r, expectedStatus)

	// test delete w/ delete policy
	tr.Spec.ReclaimPolicy = runtimev1alpha1.ReclaimDelete
	called := false
	cl.MockDelete = func(bucket *S3Bucket) error {
		called = true
		return nil
	}

	rs, err = r._delete(tr, cl)
	g.Expect(rs).To(Equal(result))
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(called).To(BeTrue())
	assertResource(g, r, expectedStatus)

	// test delete w/ delete policy error
	testError := errors.New("test-delete-error")
	called = false
	cl.MockDelete = func(bucket *S3Bucket) error {
		called = true
		return testError
	}
	expectedStatus = runtimev1alpha1.ConditionedStatus{}
	expectedStatus.SetConditions(runtimev1alpha1.Deleting(), runtimev1alpha1.ReconcileError(testError))

	rs, err = r._delete(tr, cl)
	g.Expect(rs).To(Equal(resultRequeue))
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(called).To(BeTrue())
	assertResource(g, r, expectedStatus)
}

func TestCreate(t *testing.T) {
	g := NewGomegaWithT(t)

	tr := testResource()

	r := &Reconciler{
		Client:              NewFakeClient(tr),
		ConnectionPublisher: managed.PublisherChain{}, // A no-op publisher.
		log:                 logging.NewNopLogger(),
	}

	createOrUpdateBucketCalled := false
	createUserCalled := false
	cl := &MockS3Client{
		MockCreateUser: func(username string, bucket *S3Bucket) (*iam.AccessKey, string, error) {
			createUserCalled = true
			fakeKey := &iam.AccessKey{
				AccessKeyId:     aws.String("fake-string"),
				SecretAccessKey: aws.String(""),
			}
			return fakeKey, "v2", nil
		},
		MockCreateOrUpdateBucket: func(bucket *S3Bucket) error {
			createOrUpdateBucketCalled = true
			return nil
		},
	}

	expectedStatus := runtimev1alpha1.ConditionedStatus{}
	expectedStatus.SetConditions(runtimev1alpha1.Available(), runtimev1alpha1.ReconcileSuccess())

	resource := testResource()
	rs, err := r._create(resource, cl)
	g.Expect(rs).To(Equal(result))
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(createOrUpdateBucketCalled).To(BeTrue())
	g.Expect(createUserCalled).To(BeTrue())
	assertResource(g, r, expectedStatus)
	g.Expect(resource.Status.LastUserPolicyVersion).To(Equal(2))
}

func TestCreateFail(t *testing.T) {
	g := NewGomegaWithT(t)
	tr := testResource()
	cl := &MockS3Client{
		MockCreateUser: func(username string, bucket *S3Bucket) (*iam.AccessKey, string, error) {
			fakeKey := &iam.AccessKey{
				AccessKeyId:     aws.String("fake-string"),
				SecretAccessKey: aws.String(""),
			}
			return fakeKey, "v2", nil
		},
		MockCreateOrUpdateBucket: func(bucket *S3Bucket) error {
			return nil
		},
	}

	testError := errors.New("test-publish-secret-error")
	r := &Reconciler{
		Client: NewFakeClient(tr),
		ConnectionPublisher: managed.ConnectionPublisherFns{
			PublishConnectionFn: func(_ context.Context, _ resource.Managed, _ managed.ConnectionDetails) error {
				return testError
			},
		},
		log: logging.NewNopLogger(),
	}

	expectedStatus := runtimev1alpha1.ConditionedStatus{}
	expectedStatus.SetConditions(runtimev1alpha1.Creating(), runtimev1alpha1.ReconcileError(testError))

	rs, err := r._create(tr, cl)
	g.Expect(rs).To(Equal(resultRequeue))
	g.Expect(err).NotTo(HaveOccurred())
	assertResource(g, r, expectedStatus)

	// test create resource error
	tr = testResource()
	testError = errors.New("test-create-user--error")
	called := false
	cl.MockCreateUser = func(username string, bucket *S3Bucket) (*iam.AccessKey, string, error) {
		called = true
		return nil, "", testError
	}

	expectedStatus = runtimev1alpha1.ConditionedStatus{}
	expectedStatus.SetConditions(runtimev1alpha1.Creating(), runtimev1alpha1.ReconcileError(testError))

	rs, err = r._create(tr, cl)
	g.Expect(rs).To(Equal(resultRequeue))
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(called).To(BeTrue())
	assertResource(g, r, expectedStatus)

	// test create bucket error
	cl.MockCreateUser = func(username string, bucket *S3Bucket) (*iam.AccessKey, string, error) {
		fakeKey := &iam.AccessKey{
			AccessKeyId:     aws.String("fake-string"),
			SecretAccessKey: aws.String(""),
		}
		return fakeKey, "v2", nil
	}

	tr = testResource()
	testError = errors.New("test-create-bucket--error")
	called = false
	cl.MockCreateOrUpdateBucket = func(bucket *S3Bucket) error {
		called = true
		return testError
	}

	expectedStatus = runtimev1alpha1.ConditionedStatus{}
	expectedStatus.SetConditions(runtimev1alpha1.Creating(), runtimev1alpha1.ReconcileError(testError))

	rs, err = r._create(tr, cl)
	g.Expect(rs).To(Equal(resultRequeue))
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(called).To(BeTrue())
	assertResource(g, r, expectedStatus)
}

func TestReconcile(t *testing.T) {
	g := NewGomegaWithT(t)

	tr := testResource()
	tr.Spec.IAMUsername = ""
	kube := NewFakeClient(tr)
	r := &Reconciler{
		Client:      kube,
		log:         logging.NewNopLogger(),
		initializer: managed.NewNameAsExternalName(kube),
	}

	// test connect error
	called := false
	testError := errors.New("test-connect-error")
	r.connect = func(*S3Bucket) (client.Service, error) {
		called = true
		return nil, testError
	}

	expectedStatus := runtimev1alpha1.ConditionedStatus{}
	expectedStatus.SetConditions(runtimev1alpha1.ReconcileError(testError))

	rs, err := r.Reconcile(request)
	g.Expect(rs).To(Equal(resultRequeue))
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(called).To(BeTrue())
	assertResource(g, r, expectedStatus)

	// test delete
	r.connect = func(instance *S3Bucket) (client client.Service, e error) {
		t := metav1.Now()
		instance.DeletionTimestamp = &t
		return nil, nil
	}
	called = false
	r.delete = func(instance *S3Bucket, client client.Service) (i reconcile.Result, e error) {
		called = true
		return result, nil
	}
	r.Reconcile(request)
	g.Expect(called).To(BeTrue())

	// test create
	r.connect = func(instance *S3Bucket) (client client.Service, e error) {
		return nil, nil
	}
	called = false
	r.delete = r._delete
	r.create = func(instance *S3Bucket, client client.Service) (i reconcile.Result, e error) {
		called = true
		return result, nil
	}
	r.Reconcile(request)
	g.Expect(called).To(BeTrue())

	// test sync
	r.connect = func(instance *S3Bucket) (client client.Service, e error) {
		instance.Spec.IAMUsername = "foo-user"
		return nil, nil
	}
	called = false
	r.create = r._create
	r.sync = func(instance *S3Bucket, client client.Service) (i reconcile.Result, e error) {
		called = true
		return result, nil
	}
	r.Reconcile(request)
	g.Expect(called).To(BeTrue())
}
