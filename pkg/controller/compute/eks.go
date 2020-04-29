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
	"fmt"
	"strings"
	"time"

	cf "github.com/aws/aws-sdk-go-v2/service/cloudformation"
	"github.com/ghodss/yaml"
	"github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	runtimev1alpha1 "github.com/crossplane/crossplane-runtime/apis/core/v1alpha1"
	"github.com/crossplane/crossplane-runtime/pkg/logging"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"

	awscomputev1alpha3 "github.com/crossplane/provider-aws/apis/compute/v1alpha3"
	cloudformationclient "github.com/crossplane/provider-aws/pkg/clients/cloudformation"
	"github.com/crossplane/provider-aws/pkg/clients/eks"
	"github.com/crossplane/provider-aws/pkg/controller/utils"
)

const (
	controllerName = "eks.compute.aws.crossplane.io"
	finalizer      = "finalizer." + controllerName

	eksAuthConfigMapName = "aws-auth"
	eksAuthMapRolesKey   = "mapRoles"
	eksAuthMapUsersKey   = "mapUsers"
)

var (
	ctx = context.Background()
)

// Amounts of time we wait before requeuing a reconcile.
const (
	aShortWait = 30 * time.Second
	aLongWait  = 60 * time.Second
)

// Error strings
const (
	errUpdateCustomResource = "cannot update ekscluster custom resource"
)

// CloudFormation States that are non-transitory
var (
	completedCFState = map[cf.StackStatus]bool{
		cf.StackStatusCreateComplete:   true,
		cf.StackStatusUpdateComplete:   true,
		cf.StackStatusRollbackComplete: true,
	}

	failedCFState = map[cf.StackStatus]bool{
		cf.StackStatusCreateFailed:   true,
		cf.StackStatusRollbackFailed: true,
		cf.StackStatusDeleteComplete: true,
		cf.StackStatusDeleteFailed:   true,
	}
)

// Reconciler reconciles a Provider object
type Reconciler struct {
	client.Client
	publisher managed.ConnectionPublisher
	managed.ReferenceResolver
	initializer managed.Initializer

	connect func(*awscomputev1alpha3.EKSCluster) (eks.Client, error)
	create  func(*awscomputev1alpha3.EKSCluster, eks.Client) (reconcile.Result, error)
	sync    func(*awscomputev1alpha3.EKSCluster, *eks.Cluster, eks.Client) (reconcile.Result, error)
	delete  func(*awscomputev1alpha3.EKSCluster, eks.Client) (reconcile.Result, error)
	secret  func(*eks.Cluster, *awscomputev1alpha3.EKSCluster, eks.Client) error
	awsauth func(*eks.Cluster, *awscomputev1alpha3.EKSCluster, eks.Client, string) error

	log logging.Logger
}

// SetupEKSCluster adds a controller that reconciles EKSClusters.
func SetupEKSCluster(mgr ctrl.Manager, l logging.Logger) error {
	name := managed.ControllerName(awscomputev1alpha3.EKSClusterGroupKind)

	r := &Reconciler{
		Client:            mgr.GetClient(),
		publisher:         managed.NewAPISecretPublisher(mgr.GetClient(), mgr.GetScheme()),
		ReferenceResolver: managed.NewAPISimpleReferenceResolver(mgr.GetClient()),
		log:               l.WithValues("controller", name),
		initializer:       managed.NewNameAsExternalName(mgr.GetClient()),
	}
	r.connect = r._connect
	r.create = r._create
	r.sync = r._sync
	r.delete = r._delete
	r.secret = r._secret
	r.awsauth = r._awsauth

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		For(&awscomputev1alpha3.EKSCluster{}).
		Complete(r)
}

// fail - helper function to set fail condition with reason and message
func (r *Reconciler) fail(instance *awscomputev1alpha3.EKSCluster, err error) (reconcile.Result, error) {
	instance.Status.SetConditions(runtimev1alpha1.ReconcileError(err))

	// If this is the first time we've encountered this error we'll be requeued
	// implicitly due to the status update. Otherwise we requeue after a short
	// wait in case the error condition was resolved.
	return reconcile.Result{RequeueAfter: aShortWait}, r.Update(ctx, instance)
}

func (r *Reconciler) _connect(instance *awscomputev1alpha3.EKSCluster) (eks.Client, error) {
	config, err := utils.RetrieveAwsConfigFromProvider(ctx, r, instance.Spec.ProviderReference)
	if err != nil {
		return nil, err
	}

	// NOTE(negz): EKS clusters must specify a region for creation. They never
	// use the provider's region. This should be addressed per the below issue.
	// https://github.com/crossplane/provider-aws/issues/38
	config.Region = instance.Spec.Region

	// Create new EKS Client
	return eks.NewClient(config), nil
}

func (r *Reconciler) _create(instance *awscomputev1alpha3.EKSCluster, client eks.Client) (reconcile.Result, error) {
	instance.Status.SetConditions(runtimev1alpha1.Creating())

	// Create Master
	_, err := client.Create(meta.GetExternalName(instance), instance.Spec)
	if err != nil && !eks.IsErrorAlreadyExists(err) {
		if eks.IsErrorBadRequest(err) {
			// If this was the first time we encountered this error we'll be
			// requeued implicitly. Otherwise there's no point requeuing, since
			// the error indicates our spec is bad and needs updating before
			// anything will work.
			instance.Status.SetConditions(runtimev1alpha1.ReconcileError(err))
			return reconcile.Result{}, r.Update(ctx, instance)
		}
		return r.fail(instance, err)
	}
	// Update status
	instance.Status.State = awscomputev1alpha3.ClusterStatusCreating
	instance.Status.SetConditions(runtimev1alpha1.ReconcileSuccess())

	// We'll be requeued immediately the first time we update our status
	// condition. Otherwise we want to requeue after a short wait in order to
	// determine whether the cluster is ready.
	return reconcile.Result{RequeueAfter: aShortWait}, r.Update(ctx, instance)
}

// generateAWSAuthConfigMap generates the configmap for configure auth
func generateAWSAuthConfigMap(instance *awscomputev1alpha3.EKSCluster, workerARN string) (*v1.ConfigMap, error) {
	data := map[string]string{}
	defaultRole := awscomputev1alpha3.MapRole{
		RoleARN:  workerARN,
		Username: "system:node:{{EC2PrivateDNSName}}",
		Groups:   []string{"system:bootstrappers", "system:nodes"},
	}

	// Serialize mapRoles
	roles := make([]awscomputev1alpha3.MapRole, len(instance.Spec.MapRoles))
	copy(roles, instance.Spec.MapRoles)
	roles = append(roles, defaultRole)

	rolesMarshalled, err := yaml.Marshal(roles)
	if err != nil {
		return nil, err
	}

	data[eksAuthMapRolesKey] = string(rolesMarshalled)

	// Serialize mapUsers
	if len(instance.Spec.MapUsers) > 0 {
		usersMarshalled, err := yaml.Marshal(instance.Spec.MapUsers)
		if err != nil {
			return nil, err
		}
		data[eksAuthMapUsersKey] = string(usersMarshalled)
	}

	name := eksAuthConfigMapName
	namespace := "kube-system"
	cm := v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Data: data,
	}

	return &cm, nil
}

// _awsauth generates an aws-auth configmap and pushes it to the remote eks cluster to configure auth
func (r *Reconciler) _awsauth(cluster *eks.Cluster, instance *awscomputev1alpha3.EKSCluster, client eks.Client, workerARN string) error {
	cm, err := generateAWSAuthConfigMap(instance, workerARN)
	if err != nil {
		return err
	}

	// Sync aws-auth to remote eks cluster to configure it's auth.
	token, err := client.ConnectionToken(meta.GetExternalName(instance))
	if err != nil {
		return err
	}

	// Client to eks cluster
	caData, err := base64.StdEncoding.DecodeString(cluster.CA)
	if err != nil {
		return err
	}

	c := rest.Config{
		Host: cluster.Endpoint,
		TLSClientConfig: rest.TLSClientConfig{
			CAData: caData,
		},
		BearerToken: token,
	}

	clientset, err := kubernetes.NewForConfig(&c)
	if err != nil {
		return err
	}

	// Create or update aws-auth configmap on eks cluster
	_, err = clientset.CoreV1().ConfigMaps(cm.Namespace).Create(ctx, cm, metav1.CreateOptions{})
	if apierrors.IsAlreadyExists(err) {
		_, err = clientset.CoreV1().ConfigMaps(cm.Namespace).Update(ctx, cm, metav1.UpdateOptions{})
	}

	return err
}

func (r *Reconciler) _sync(instance *awscomputev1alpha3.EKSCluster, cluster *eks.Cluster, client eks.Client) (reconcile.Result, error) {
	if cluster.Status != awscomputev1alpha3.ClusterStatusActive {
		instance.Status.SetConditions(runtimev1alpha1.ReconcileSuccess())
		// Requeue after a short wait to see if the cluster has become ready.
		return reconcile.Result{RequeueAfter: aShortWait}, nil
	}

	// Create workers
	if instance.Spec.CloudFormationStackID == "" {
		clusterWorkers, err := client.CreateWorkerNodes(meta.GetExternalName(instance), cluster.Version, instance.Spec)
		if err != nil {
			return r.fail(instance, err)
		}
		instance.Spec.CloudFormationStackID = clusterWorkers.WorkerStackID
		instance.Status.SetConditions(runtimev1alpha1.ReconcileSuccess())

		// We'll likely be requeued implicitly due to the status update, but
		// otherwise we want to requeue a reconcile after a short wait to check
		// on the worker node creation.
		return reconcile.Result{RequeueAfter: aShortWait}, r.Update(ctx, instance)
	}

	clusterWorker, err := client.GetWorkerNodes(instance.Spec.CloudFormationStackID)
	if err != nil {
		return r.fail(instance, err)
	}

	if failedCFState[clusterWorker.WorkersStatus] {
		return r.fail(instance, fmt.Errorf("clusterworker stack failed with status %q and reason %q", clusterWorker.WorkersStatus, clusterWorker.WorkerReason))
	}

	if !completedCFState[clusterWorker.WorkersStatus] {
		instance.Status.SetConditions(runtimev1alpha1.ReconcileSuccess())

		return reconcile.Result{RequeueAfter: aShortWait}, r.Update(ctx, instance)
	}

	if err := r.awsauth(cluster, instance, client, clusterWorker.WorkerARN); err != nil {
		return r.fail(instance, errors.Wrap(err, "failed to set auth map on eks"))
	}

	if err := r.secret(cluster, instance, client); err != nil {
		return r.fail(instance, err)
	}

	// update resource status
	instance.Status.Endpoint = cluster.Endpoint
	instance.Status.State = awscomputev1alpha3.ClusterStatusActive
	instance.Status.ClusterVersion = cluster.Version
	instance.Status.SetConditions(runtimev1alpha1.Available(), runtimev1alpha1.ReconcileSuccess())
	resource.SetBindable(instance)

	// Our cluster is available. Requeue speculatively after a long wait in case
	// the cluster has changed.
	return reconcile.Result{RequeueAfter: aLongWait}, r.Update(ctx, instance)
}

func (r *Reconciler) _secret(cluster *eks.Cluster, instance *awscomputev1alpha3.EKSCluster, client eks.Client) error {
	token, err := client.ConnectionToken(meta.GetExternalName(instance))
	if err != nil {
		return err
	}

	config, err := eks.GenerateClientConfig(cluster, token)
	if err != nil {
		return err
	}
	rawConfig, err := clientcmd.Write(config)
	if err != nil {
		return err
	}

	return r.publisher.PublishConnection(ctx, instance, managed.ConnectionDetails{
		runtimev1alpha1.ResourceCredentialsSecretEndpointKey:   []byte(config.Clusters[cluster.Name].Server),
		runtimev1alpha1.ResourceCredentialsSecretCAKey:         config.Clusters[cluster.Name].CertificateAuthorityData,
		runtimev1alpha1.ResourceCredentialsSecretTokenKey:      []byte(config.AuthInfos[cluster.Name].Token),
		runtimev1alpha1.ResourceCredentialsSecretKubeconfigKey: rawConfig,
	})
}

// _delete check reclaim policy and if needed delete the eks cluster resource
func (r *Reconciler) _delete(instance *awscomputev1alpha3.EKSCluster, client eks.Client) (reconcile.Result, error) {
	instance.Status.SetConditions(runtimev1alpha1.Deleting())
	if instance.Spec.ReclaimPolicy == runtimev1alpha1.ReclaimDelete {
		var deleteErrors []string
		if err := client.Delete(meta.GetExternalName(instance)); err != nil && !eks.IsErrorNotFound(err) {
			deleteErrors = append(deleteErrors, fmt.Sprintf("Master Delete Error: %s", err.Error()))
		}

		if instance.Spec.CloudFormationStackID != "" {
			if err := client.DeleteWorkerNodes(instance.Spec.CloudFormationStackID); err != nil && !cloudformationclient.IsErrorNotFound(err) {
				deleteErrors = append(deleteErrors, fmt.Sprintf("Worker Delete Error: %s", err.Error()))
			}
		}

		if len(deleteErrors) > 0 {
			return r.fail(instance, errors.New(strings.Join(deleteErrors, ", ")))
		}
	}

	meta.RemoveFinalizer(instance, finalizer)
	instance.Status.SetConditions(runtimev1alpha1.ReconcileSuccess())

	// No need to requeue a reconcile if we've successfully asked for the
	// cluster to be deleted.
	return reconcile.Result{Requeue: false}, r.Update(ctx, instance)
}

// Reconcile reads that state of the cluster for a Provider object and makes changes based on the state read
// and what is in the Provider.Spec
func (r *Reconciler) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	r.log.Debug("Reconciling", "request", request)
	// Fetch the Provider instance
	instance := &awscomputev1alpha3.EKSCluster{}
	err := r.Get(ctx, request.NamespacedName, instance)
	if err != nil {
		// No need to requeue if the resource no longer exists, otherwise we'll
		// be requeued because we return an error.
		return reconcile.Result{}, resource.IgnoreNotFound(err)
	}
	if err := r.initializer.Initialize(ctx, instance); err != nil {
		return reconcile.Result{}, err
	}

	// Create EKS Client
	eksClient, err := r.connect(instance)
	if err != nil {
		return r.fail(instance, err)
	}

	if err := r.ResolveReferences(ctx, instance); err != nil {
		instance.Status.SetConditions(runtimev1alpha1.ReconcileError(err))
		return reconcile.Result{RequeueAfter: aLongWait}, errors.Wrap(r.Update(ctx, instance), errUpdateCustomResource)
	}

	// Add finalizer
	meta.AddFinalizer(instance, finalizer)

	// Check for deletion
	if instance.DeletionTimestamp != nil {
		return r.delete(instance, eksClient)
	}

	cluster, err := eksClient.Get(meta.GetExternalName(instance))
	switch {
	case eks.IsErrorNotFound(err):
		return r.create(instance, eksClient)
	case err != nil:
		return r.fail(instance, err)
	}
	// Sync cluster instance status with cluster status
	return r.sync(instance, cluster, eksClient)
}
