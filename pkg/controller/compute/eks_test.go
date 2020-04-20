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

package compute

import (
	"context"
	"encoding/base64"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/cloudformation"
	"github.com/aws/aws-sdk-go-v2/service/eks"
	"github.com/ghodss/yaml"
	"github.com/google/go-cmp/cmp"
	. "github.com/onsi/gomega"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/clientcmd"
	. "sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/crossplane/provider-aws/apis"
	. "github.com/crossplane/provider-aws/apis/compute/v1alpha3"
	eksclients "github.com/crossplane/provider-aws/pkg/clients/eks"
	"github.com/crossplane/provider-aws/pkg/clients/eks/fake"

	runtimev1alpha1 "github.com/crossplane/crossplane-runtime/apis/core/v1alpha1"
	"github.com/crossplane/crossplane-runtime/pkg/logging"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	"github.com/crossplane/crossplane-runtime/pkg/test"
)

const (
	namespace    = "default"
	providerName = "test-provider"
	clusterName  = "test-cluster"
)

var (
	key = types.NamespacedName{
		Name: clusterName,
	}
	request = reconcile.Request{
		NamespacedName: key,
	}
)

func init() {
	_ = apis.AddToScheme(scheme.Scheme)
}

func testCluster() *EKSCluster {
	return &EKSCluster{
		ObjectMeta: metav1.ObjectMeta{
			Name: clusterName,
		},
		Spec: EKSClusterSpec{
			ResourceSpec: runtimev1alpha1.ResourceSpec{
				ProviderReference: &corev1.ObjectReference{
					Name: providerName,
				},
			},
		},
	}
}

// assertResource a helper function to check on cluster and its status
func assertResource(g *GomegaWithT, r *Reconciler, s runtimev1alpha1.ConditionedStatus) *EKSCluster {
	rc := &EKSCluster{}
	err := r.Get(ctx, key, rc)
	g.Expect(err).To(BeNil())
	g.Expect(cmp.Diff(s, rc.Status.ConditionedStatus, test.EquateConditions())).Should(BeZero())
	return rc
}

func TestGenerateEksAuth(t *testing.T) {
	g := NewGomegaWithT(t)
	arnName := "test-arn"
	var expectRoles []MapRole
	var expectUsers []MapUser

	defaultMapRole := MapRole{
		RoleARN:  arnName,
		Username: "system:node:{{EC2PrivateDNSName}}",
		Groups:   []string{"system:bootstrappers", "system:nodes"},
	}

	exampleMapRole := MapRole{
		RoleARN:  "arn:aws:iam::000000000000:role/KubernetesAdmin",
		Username: "kubernetes-admin",
		Groups:   []string{"system:masters"},
	}

	exampleMapUser := MapUser{
		UserARN:  "arn:aws:iam::000000000000:user/Alice",
		Username: "alice",
		Groups:   []string{"system:masters"},
	}

	expectRoles = append(expectRoles, exampleMapRole)
	expectUsers = append(expectUsers, exampleMapUser)

	cluster := testCluster()
	cluster.Spec.MapRoles = expectRoles
	cluster.Spec.MapUsers = expectUsers

	// Default is included by so we don't add it to spec
	expectRoles = append(expectRoles, defaultMapRole)

	cm, err := generateAWSAuthConfigMap(cluster, arnName)
	g.Expect(err).To(BeNil())

	g.Expect(cm.Name).To(Equal("aws-auth"))
	g.Expect(cm.Namespace).To(Equal("kube-system"))

	var outputRoles []MapRole
	val := cm.Data["mapRoles"]
	err = yaml.Unmarshal([]byte(val), &outputRoles)
	g.Expect(err).To(BeNil())

	var outputUsers []MapUser
	val = cm.Data["mapUsers"]
	err = yaml.Unmarshal([]byte(val), &outputUsers)
	g.Expect(err).To(BeNil())

	g.Expect(outputRoles).To(Equal(expectRoles))
	g.Expect(outputUsers).To(Equal(expectUsers))
}

func TestCreate(t *testing.T) {
	g := NewGomegaWithT(t)

	test := func(cluster *EKSCluster, client eksclients.Client, expectedResult reconcile.Result, expectedStatus runtimev1alpha1.ConditionedStatus) *EKSCluster {
		r := &Reconciler{
			Client:            NewFakeClient(cluster),
			ReferenceResolver: managed.NewAPISimpleReferenceResolver(NewFakeClient(cluster)),
		}

		rs, err := r._create(cluster, client)
		g.Expect(rs).To(Equal(expectedResult))
		g.Expect(err).To(BeNil())
		return assertResource(g, r, expectedStatus)
	}

	// new cluster
	cluster := testCluster()

	client := &fake.MockEKSClient{
		MockCreate: func(string, EKSClusterSpec) (*eksclients.Cluster, error) { return &eksclients.Cluster{}, nil },
	}

	expectedStatus := runtimev1alpha1.ConditionedStatus{}
	expectedStatus.SetConditions(runtimev1alpha1.Creating(), runtimev1alpha1.ReconcileSuccess())

	reconciledCluster := test(cluster, client, reconcile.Result{RequeueAfter: aShortWait}, expectedStatus)
	g.Expect(reconciledCluster.Status.State).To(Equal(ClusterStatusCreating))

	// cluster create error - bad request
	cluster = testCluster()
	errorBadRequest := errors.New("InvalidParameterException")
	client.MockCreate = func(string, EKSClusterSpec) (*eksclients.Cluster, error) {
		return &eksclients.Cluster{}, errorBadRequest
	}
	expectedStatus = runtimev1alpha1.ConditionedStatus{}
	expectedStatus.SetConditions(runtimev1alpha1.Creating(), runtimev1alpha1.ReconcileError(errorBadRequest))

	reconciledCluster = test(cluster, client, reconcile.Result{}, expectedStatus)
	g.Expect(reconciledCluster.Finalizers).To(BeEmpty())
	g.Expect(reconciledCluster.Status.State).To(BeEmpty())
	g.Expect(reconciledCluster.Spec.CloudFormationStackID).To(BeEmpty())

	// cluster create error - other
	cluster = testCluster()
	cluster.ObjectMeta.UID = types.UID("test-uid")
	errorOther := errors.New("other")
	client.MockCreate = func(string, EKSClusterSpec) (*eksclients.Cluster, error) {
		return &eksclients.Cluster{}, errorOther
	}
	expectedStatus = runtimev1alpha1.ConditionedStatus{}
	expectedStatus.SetConditions(runtimev1alpha1.Creating(), runtimev1alpha1.ReconcileError(errorOther))

	reconciledCluster = test(cluster, client, reconcile.Result{RequeueAfter: aShortWait}, expectedStatus)
	g.Expect(reconciledCluster.Finalizers).To(BeEmpty())
	g.Expect(reconciledCluster.Status.State).To(BeEmpty())
	g.Expect(reconciledCluster.Spec.CloudFormationStackID).To(BeEmpty())
}

func TestSync(t *testing.T) {
	g := NewGomegaWithT(t)
	fakeStackID := "fake-stack-id"

	test := func(tc *EKSCluster, cluster *eksclients.Cluster, client *fake.MockEKSClient, sec func(*eksclients.Cluster, *EKSCluster, eksclients.Client) error, auth func(*eksclients.Cluster, *EKSCluster, eksclients.Client, string) error,
		rslt reconcile.Result, exp runtimev1alpha1.ConditionedStatus) *EKSCluster {
		r := &Reconciler{
			Client:            NewFakeClient(tc),
			secret:            sec,
			awsauth:           auth,
			ReferenceResolver: managed.NewAPISimpleReferenceResolver(NewFakeClient()),
		}

		rs, err := r._sync(tc, cluster, client)
		g.Expect(rs).To(Equal(rslt))
		g.Expect(err).NotTo(HaveOccurred())
		return assertResource(g, r, exp)
	}

	fakeWorkerARN := "fake-worker-arn"
	mockClusterWorker := eksclients.ClusterWorkers{
		WorkerStackID: fakeStackID,
		WorkerARN:     fakeWorkerARN,
	}

	// cluster is not ready
	cluster := &eksclients.Cluster{
		Status: ClusterStatusCreating,
	}
	cl := &fake.MockEKSClient{
		MockCreateWorkerNodes: func(string, string, EKSClusterSpec) (*eksclients.ClusterWorkers, error) {
			return &mockClusterWorker, nil
		},
		MockGetWorkerNodes: func(string) (*eksclients.ClusterWorkers, error) {
			return &eksclients.ClusterWorkers{
				WorkersStatus: cloudformation.StackStatusCreateInProgress,
				WorkerReason:  "",
				WorkerStackID: fakeStackID}, nil
		},
	}
	expectedStatus := runtimev1alpha1.ConditionedStatus{}
	tc := testCluster()
	test(tc, cluster, cl, nil, nil, reconcile.Result{RequeueAfter: aShortWait}, expectedStatus)

	// cluster is ready, but lets create workers that error
	cluster = &eksclients.Cluster{
		Status: ClusterStatusActive,
	}

	errorCreateNodes := errors.New("create nodes")
	cl.MockCreateWorkerNodes = func(string, string, EKSClusterSpec) (*eksclients.ClusterWorkers, error) {
		return nil, errorCreateNodes
	}

	expectedStatus = runtimev1alpha1.ConditionedStatus{}
	expectedStatus.SetConditions(runtimev1alpha1.ReconcileError(errorCreateNodes))
	tc = testCluster()
	reconciledCluster := test(tc, cluster, cl, nil, nil, reconcile.Result{RequeueAfter: aShortWait}, expectedStatus)
	g.Expect(reconciledCluster.Spec.CloudFormationStackID).To(BeEmpty())

	// cluster is ready, lets create workers
	cluster = &eksclients.Cluster{
		Status: ClusterStatusActive,
	}

	cl.MockCreateWorkerNodes = func(string, string, EKSClusterSpec) (*eksclients.ClusterWorkers, error) {
		return &eksclients.ClusterWorkers{WorkerStackID: fakeStackID}, nil
	}

	expectedStatus = runtimev1alpha1.ConditionedStatus{}
	expectedStatus.SetConditions(runtimev1alpha1.ReconcileSuccess())
	tc = testCluster()
	reconciledCluster = test(tc, cluster, cl, nil, nil, reconcile.Result{RequeueAfter: aShortWait}, expectedStatus)
	g.Expect(reconciledCluster.Spec.CloudFormationStackID).To(Equal(fakeStackID))

	// cluster is ready, but auth sync failed
	cl.MockGetWorkerNodes = func(string) (*eksclients.ClusterWorkers, error) {
		return &eksclients.ClusterWorkers{
			WorkersStatus: cloudformation.StackStatusCreateComplete,
			WorkerReason:  "",
			WorkerStackID: fakeStackID,
			WorkerARN:     fakeWorkerARN,
		}, nil
	}

	errorAuth := errors.New("auth")
	expectedStatus = runtimev1alpha1.ConditionedStatus{}
	expectedStatus.SetConditions(runtimev1alpha1.ReconcileError(errors.Wrap(errorAuth, "failed to set auth map on eks")))
	tc = testCluster()
	tc.Spec.CloudFormationStackID = fakeStackID
	auth := func(*eksclients.Cluster, *EKSCluster, eksclients.Client, string) error {
		return errorAuth
	}
	test(tc, cluster, cl, nil, auth, reconcile.Result{RequeueAfter: aShortWait}, expectedStatus)

	// cluster is ready, but secret failed
	cl.MockGetWorkerNodes = func(string) (*eksclients.ClusterWorkers, error) {
		return &eksclients.ClusterWorkers{
			WorkersStatus: cloudformation.StackStatusCreateComplete,
			WorkerReason:  "",
			WorkerStackID: fakeStackID,
			WorkerARN:     fakeWorkerARN,
		}, nil
	}

	auth = func(*eksclients.Cluster, *EKSCluster, eksclients.Client, string) error {
		return nil
	}

	errorSecret := errors.New("secret")
	fSec := func(*eksclients.Cluster, *EKSCluster, eksclients.Client) error {
		return errorSecret
	}
	expectedStatus = runtimev1alpha1.ConditionedStatus{}
	expectedStatus.SetConditions(runtimev1alpha1.ReconcileError(errorSecret))
	tc = testCluster()
	tc.Spec.CloudFormationStackID = fakeStackID
	test(tc, cluster, cl, fSec, auth, reconcile.Result{RequeueAfter: aShortWait}, expectedStatus)

	// cluster is ready
	fSec = func(*eksclients.Cluster, *EKSCluster, eksclients.Client) error {
		return nil
	}
	expectedStatus = runtimev1alpha1.ConditionedStatus{}
	expectedStatus.SetConditions(runtimev1alpha1.Available(), runtimev1alpha1.ReconcileSuccess())
	tc = testCluster()
	tc.Spec.CloudFormationStackID = fakeStackID
	test(tc, cluster, cl, fSec, auth, reconcile.Result{RequeueAfter: aLongWait}, expectedStatus)
}

func TestSecret(t *testing.T) {
	clusterCA := []byte("test-ca")
	token := "test-token"
	cluster := &eksclients.Cluster{
		Status:   ClusterStatusActive,
		Endpoint: "test-ep",
		CA:       base64.StdEncoding.EncodeToString(clusterCA),
	}
	config, _ := eksclients.GenerateClientConfig(cluster, token)
	rawConfig, _ := clientcmd.Write(config)
	r := &Reconciler{
		publisher: managed.ConnectionPublisherFns{
			PublishConnectionFn: func(_ context.Context, _ resource.Managed, got managed.ConnectionDetails) error {
				want := managed.ConnectionDetails{
					runtimev1alpha1.ResourceCredentialsSecretEndpointKey:   []byte(cluster.Endpoint),
					runtimev1alpha1.ResourceCredentialsSecretCAKey:         clusterCA,
					runtimev1alpha1.ResourceCredentialsSecretTokenKey:      []byte(token),
					runtimev1alpha1.ResourceCredentialsSecretKubeconfigKey: rawConfig,
				}

				if diff := cmp.Diff(want, got); diff != "" {
					t.Errorf("-want, +got\n%s", diff)
				}

				return nil
			},
		},
		ReferenceResolver: managed.NewAPISimpleReferenceResolver(NewFakeClient()),
	}

	tc := testCluster()
	client := &fake.MockEKSClient{}

	// Ensure we return an error when we can't get a new token.
	testError := "test-connection-token-error"
	client.MockConnectionToken = func(string) (string, error) { return "", errors.New(testError) }
	want := errors.New(testError)
	got := r._secret(cluster, tc, client)
	if diff := cmp.Diff(want, got, test.EquateErrors()); diff != "" {
		t.Errorf("r._secret(...): -want error, +got error:\n%s", diff)
	}

	// Ensure we don't return an error when we can get a new token.
	client.MockConnectionToken = func(string) (string, error) { return "test-token", nil }
	if err := r._secret(cluster, tc, client); err != nil {
		t.Errorf("r._secret(...): %s", err)
	}
}

func TestDelete(t *testing.T) {
	g := NewGomegaWithT(t)

	test := func(cluster *EKSCluster, client eksclients.Client, expectedResult reconcile.Result, expectedStatus runtimev1alpha1.ConditionedStatus) *EKSCluster {
		r := &Reconciler{
			Client:            NewFakeClient(cluster),
			ReferenceResolver: managed.NewAPISimpleReferenceResolver(NewFakeClient(cluster)),
		}

		rs, err := r._delete(cluster, client)
		g.Expect(rs).To(Equal(expectedResult))
		g.Expect(err).To(BeNil())
		return assertResource(g, r, expectedStatus)
	}

	// reclaim - delete
	cluster := testCluster()
	cluster.Finalizers = []string{finalizer}
	cluster.Spec.ReclaimPolicy = runtimev1alpha1.ReclaimDelete
	cluster.Spec.CloudFormationStackID = "fake-stack-id"
	cluster.Status.SetConditions(runtimev1alpha1.Available())

	client := &fake.MockEKSClient{}
	client.MockDelete = func(string) error { return nil }
	client.MockDeleteWorkerNodes = func(string) error { return nil }

	expectedStatus := runtimev1alpha1.ConditionedStatus{}
	expectedStatus.SetConditions(runtimev1alpha1.Deleting(), runtimev1alpha1.ReconcileSuccess())

	reconciledCluster := test(cluster, client, reconcile.Result{}, expectedStatus)
	g.Expect(reconciledCluster.Finalizers).To(BeEmpty())

	// reclaim - retain
	cluster.Spec.ReclaimPolicy = runtimev1alpha1.ReclaimRetain
	cluster.Status.SetConditions(runtimev1alpha1.Available())
	cluster.Finalizers = []string{finalizer}
	client.MockDelete = nil // should not be called

	reconciledCluster = test(cluster, client, reconcile.Result{}, expectedStatus)
	g.Expect(reconciledCluster.Finalizers).To(BeEmpty())

	// reclaim - delete, delete error
	cluster.Spec.ReclaimPolicy = runtimev1alpha1.ReclaimDelete
	cluster.Status.SetConditions(runtimev1alpha1.Available())
	cluster.Finalizers = []string{finalizer}
	errorDelete := errors.New("test-delete-error")
	client.MockDelete = func(string) error { return errorDelete }
	expectedStatus = runtimev1alpha1.ConditionedStatus{}
	expectedStatus.SetConditions(
		runtimev1alpha1.Deleting(),
		runtimev1alpha1.ReconcileError(errors.Wrap(errorDelete, "Master Delete Error")),
	)

	reconciledCluster = test(cluster, client, reconcile.Result{RequeueAfter: aShortWait}, expectedStatus)
	g.Expect(reconciledCluster.Finalizers).To(ContainElement(finalizer))

	// reclaim - delete, delete error with worker
	cluster.Spec.ReclaimPolicy = runtimev1alpha1.ReclaimDelete
	cluster.Status.SetConditions(runtimev1alpha1.Available())
	cluster.Finalizers = []string{finalizer}
	testErrorWorker := errors.New("test-delete-error-worker")
	client.MockDelete = func(string) error { return nil }
	client.MockDeleteWorkerNodes = func(string) error { return testErrorWorker }
	expectedStatus = runtimev1alpha1.ConditionedStatus{}
	expectedStatus.SetConditions(
		runtimev1alpha1.Deleting(),
		runtimev1alpha1.ReconcileError(errors.Wrap(testErrorWorker, "Worker Delete Error")),
	)

	reconciledCluster = test(cluster, client, reconcile.Result{RequeueAfter: aShortWait}, expectedStatus)
	g.Expect(reconciledCluster.Finalizers).To(ContainElement(finalizer))

	// reclaim - delete, delete error in cluster and cluster workers
	cluster.Spec.ReclaimPolicy = runtimev1alpha1.ReclaimDelete
	cluster.Status.SetConditions(runtimev1alpha1.Available())
	cluster.Finalizers = []string{finalizer}
	client.MockDelete = func(string) error { return nil }
	client.MockDelete = func(string) error { return errorDelete }
	client.MockDeleteWorkerNodes = func(string) error { return testErrorWorker }
	expectedStatus = runtimev1alpha1.ConditionedStatus{}
	expectedStatus.SetConditions(
		runtimev1alpha1.Deleting(),
		runtimev1alpha1.ReconcileError(errors.New("Master Delete Error: test-delete-error, Worker Delete Error: test-delete-error-worker")),
	)

	reconciledCluster = test(cluster, client, reconcile.Result{RequeueAfter: aShortWait}, expectedStatus)
	g.Expect(reconciledCluster.Finalizers).To(ContainElement(finalizer))
}

func TestReconcileObjectNotFound(t *testing.T) {
	g := NewGomegaWithT(t)

	r := &Reconciler{
		Client:            NewFakeClient(),
		ReferenceResolver: managed.NewAPISimpleReferenceResolver(NewFakeClient()),
		log:               logging.NewNopLogger(),
	}
	rs, err := r.Reconcile(request)
	g.Expect(rs).To(Equal(reconcile.Result{}))
	g.Expect(err).To(BeNil())
}

func TestReconcileClientError(t *testing.T) {
	g := NewGomegaWithT(t)

	testError := errors.New("test-client-error")

	called := false
	kube := NewFakeClient(testCluster())
	r := &Reconciler{
		Client: kube,
		connect: func(*EKSCluster) (eksclients.Client, error) {
			called = true
			return nil, testError
		},
		ReferenceResolver: managed.NewAPISimpleReferenceResolver(kube),
		log:               logging.NewNopLogger(),
		initializer:       managed.NewNameAsExternalName(kube),
	}

	// expected to have a failed condition
	expectedStatus := runtimev1alpha1.ConditionedStatus{}
	expectedStatus.SetConditions(runtimev1alpha1.ReconcileError(testError))

	rs, err := r.Reconcile(request)
	g.Expect(rs).To(Equal(reconcile.Result{RequeueAfter: aShortWait}))
	g.Expect(err).To(BeNil())
	g.Expect(called).To(BeTrue())

	assertResource(g, r, expectedStatus)
}

func TestReconcileDelete(t *testing.T) {
	g := NewGomegaWithT(t)

	// test objects
	tc := testCluster()
	dt := metav1.Now()
	tc.DeletionTimestamp = &dt

	called := false
	kube := NewFakeClient(tc)
	r := &Reconciler{
		Client: kube,
		connect: func(*EKSCluster) (eksclients.Client, error) {
			return nil, nil
		},
		delete: func(*EKSCluster, eksclients.Client) (reconcile.Result, error) {
			called = true
			return reconcile.Result{}, nil
		},
		ReferenceResolver: managed.NewAPISimpleReferenceResolver(kube),
		log:               logging.NewNopLogger(),
		initializer:       managed.NewNameAsExternalName(kube),
	}

	rs, err := r.Reconcile(request)
	g.Expect(rs).To(Equal(reconcile.Result{}))
	g.Expect(err).To(BeNil())
	g.Expect(called).To(BeTrue())
	assertResource(g, r, runtimev1alpha1.ConditionedStatus{})
}

func TestReconcileCreate(t *testing.T) {
	g := NewGomegaWithT(t)

	called := false
	kube := NewFakeClient(testCluster())
	r := &Reconciler{
		Client: kube,
		connect: func(*EKSCluster) (eksclients.Client, error) {
			return &fake.MockEKSClient{MockGet: func(_ string) (*eksclients.Cluster, error) {
				return nil, errors.New(eks.ErrCodeResourceNotFoundException)
			}}, nil
		},
		create: func(*EKSCluster, eksclients.Client) (reconcile.Result, error) {
			called = true
			return reconcile.Result{RequeueAfter: aShortWait}, nil
		},
		ReferenceResolver: managed.NewAPISimpleReferenceResolver(kube),
		log:               logging.NewNopLogger(),
		initializer:       managed.NewNameAsExternalName(kube),
	}

	rs, err := r.Reconcile(request)
	g.Expect(rs).To(Equal(reconcile.Result{RequeueAfter: aShortWait}))
	g.Expect(err).To(BeNil())
	g.Expect(called).To(BeTrue())

	assertResource(g, r, runtimev1alpha1.ConditionedStatus{})
}

func TestReconcileSync(t *testing.T) {
	g := NewGomegaWithT(t)

	called := false

	tc := testCluster()
	tc.Finalizers = []string{finalizer}
	kube := NewFakeClient(tc)
	r := &Reconciler{
		Client: kube,
		connect: func(*EKSCluster) (eksclients.Client, error) {
			return &fake.MockEKSClient{MockGet: func(_ string) (*eksclients.Cluster, error) {
				return &eksclients.Cluster{}, nil
			}}, nil
		},
		sync: func(*EKSCluster, *eksclients.Cluster, eksclients.Client) (reconcile.Result, error) {
			called = true
			return reconcile.Result{RequeueAfter: aShortWait}, nil
		},
		ReferenceResolver: managed.NewAPISimpleReferenceResolver(kube),
		log:               logging.NewNopLogger(),
		initializer:       managed.NewNameAsExternalName(kube),
	}

	rs, err := r.Reconcile(request)
	g.Expect(rs).To(Equal(reconcile.Result{RequeueAfter: aShortWait}))
	g.Expect(err).To(BeNil())
	g.Expect(called).To(BeTrue())

	rc := assertResource(g, r, runtimev1alpha1.ConditionedStatus{})
	g.Expect(rc.Finalizers).To(HaveLen(1))
	g.Expect(rc.Finalizers).To(ContainElement(finalizer))
}
