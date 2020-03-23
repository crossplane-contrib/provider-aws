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
	"encoding/json"
	"fmt"
	"strings"
	"time"

	cf "github.com/aws/aws-sdk-go-v2/service/cloudformation"
	"github.com/ghodss/yaml"
	"github.com/pkg/errors"

	v1 "k8s.io/api/core/v1"

	"k8s.io/apimachinery/pkg/runtime/schema"

	apiextensionsv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	apiextension "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/types"

	"k8s.io/client-go/dynamic"
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
	controllerName    = "eks.compute.aws.crossplane.io"
	finalizer         = "finalizer." + controllerName
	clusterNamePrefix = "eks-"

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
	errUpdateManagedStatus = "cannot update managed resource status"
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

	connect       func(*awscomputev1alpha3.EKSCluster) (eks.Client, error)
	create        func(*awscomputev1alpha3.EKSCluster, eks.Client) (reconcile.Result, error)
	sync          func(*awscomputev1alpha3.EKSCluster, eks.Client) (reconcile.Result, error)
	delete        func(*awscomputev1alpha3.EKSCluster, eks.Client) (reconcile.Result, error)
	secret        func(*eks.Cluster, *awscomputev1alpha3.EKSCluster, eks.Client) error
	awsauth       func(*eks.Cluster, *awscomputev1alpha3.EKSCluster, eks.Client, string) error
	customnetwork func(*eks.Cluster, *awscomputev1alpha3.EKSCluster, eks.Client) error

	log logging.Logger
}

// SetupEKSCluster adds a controller that reconciles EKSClusters.
func SetupEKSCluster(mgr ctrl.Manager, l logging.Logger) error {
	name := managed.ControllerName(awscomputev1alpha3.EKSClusterGroupKind)

	r := &Reconciler{
		Client:            mgr.GetClient(),
		publisher:         managed.NewAPISecretPublisher(mgr.GetClient(), mgr.GetScheme()),
		ReferenceResolver: managed.NewAPIReferenceResolver(mgr.GetClient()),
		log:               l.WithValues("controller", name),
	}
	r.connect = r._connect
	r.create = r._create
	r.sync = r._sync
	r.delete = r._delete
	r.secret = r._secret
	r.awsauth = r._awsauth
	r.customnetwork = r._customnetwork

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
	clusterName := fmt.Sprintf("%s%s", clusterNamePrefix, instance.UID)

	// Create Master
	createdCluster, err := client.Create(clusterName, instance.Spec)
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

	// we will need to set State.ClusterVersion it. this is needed to retrieve
	// the right ami image for the worker nodes
	instance.Status.ClusterVersion = createdCluster.Version

	// Update status
	instance.Status.State = awscomputev1alpha3.ClusterStatusCreating
	instance.Status.ClusterName = clusterName

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
	token, err := client.ConnectionToken(instance.Status.ClusterName)
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
	_, err = clientset.CoreV1().ConfigMaps(cm.Namespace).Create(cm)
	if err != nil {
		if apierrors.IsAlreadyExists(err) {
			_, err = clientset.CoreV1().ConfigMaps(cm.Namespace).Update(cm)
		}
	}

	return err
}

// setup a custom network (secondary cidr ranges) on EKS
// per https://docs.aws.amazon.com/eks/latest/userguide/cni-custom-network.html
// and https://aws.amazon.com/de/premiumsupport/knowledge-center/eks-multiple-cidr-ranges/
func (r *Reconciler) _customnetwork(cluster *eks.Cluster, instance *awscomputev1alpha3.EKSCluster, client eks.Client) error {

	// noop if custom network feature isn't set
	if len(instance.Spec.CustomSubnetIDs) == 0 {
		return nil
	}

	// Sync aws-node to remote eks cluster to configure its network.
	token, err := client.ConnectionToken(instance.Status.ClusterName)
	if err != nil {
		return err
	}

	// Client to eks cluster
	caData, err := base64.StdEncoding.DecodeString(cluster.CA)
	if err != nil {
		return err
	}

	config := rest.Config{
		Host: cluster.Endpoint,
		TLSClientConfig: rest.TLSClientConfig{
			CAData: caData,
		},
		BearerToken: token,
	}

	// Patch aws-node daemonset
	err = r.patchAWSNodeDaemonSet(config)
	if err != nil {
		return err
	}

	// Create ENIConfig CRD & CRs
	err = r.createENIConfigCRD(config)
	if err != nil {
		return err
	}
	err = r.createENIConfigCRs(instance, client, config)
	if err != nil {
		return err
	}

	return err
}

// Patch aws-node daemonset on eks cluster
// AWS_VPC_K8S_CNI_CUSTOM_NETWORK_CFG=true
// ENI_CONFIG_LABEL_DEF=failure-domain.beta.kubernetes.io/zone
func (r *Reconciler) patchAWSNodeDaemonSet(config rest.Config) error {

	clientset, err := kubernetes.NewForConfig(&config)
	if err != nil {
		return err
	}

	// NOT WORKING:
	// failed to set custom network on eks: DaemonSet.apps "aws-node" is invalid: [spec.template.metadata.labels: Invalid value: map[string]string{"k8s-app":"aws-node"}: `selector` does not match template `labels`, spec.selector: Invalid value: "null": field is immutable]
	// patch := &apps.DaemonSet{
	// 	Spec: apps.DaemonSetSpec{
	// 		Template: v1.PodTemplateSpec{
	// 			Spec: v1.PodSpec{
	// 				Containers: []v1.Container{
	// 					{
	// 						Name: "aws-node",
	// 						Env: []v1.EnvVar{
	// 							{
	// 								Name:  "AWS_VPC_K8S_CNI_CUSTOM_NETWORK_CFG",
	// 								Value: "true",
	// 							},
	// 							{
	// 								Name:  "ENI_CONFIG_LABEL_DEF",
	// 								Value: "failure-domain.beta.kubernetes.io/zone",
	// 							},
	// 						},
	// 					},
	// 				},
	// 			},
	// 		},
	// 	},
	// }
	// byteBuffer := new(bytes.Buffer)
	// if err := json.NewEncoder(byteBuffer).Encode(patch); err != nil {
	// 	return err
	// }

	// patch := byteBuffer.Bytes()
	patch, _ := json.Marshal(map[string]interface{}{
		"spec": map[string]interface{}{
			"template": map[string]interface{}{
				"spec": map[string]interface{}{
					"containers": []interface{}{
						map[string]interface{}{
							"name": "aws-node",
							"env": []interface{}{
								map[string]interface{}{"name": "AWS_VPC_K8S_CNI_CUSTOM_NETWORK_CFG", "value": "true"},
								map[string]interface{}{"name": "ENI_CONFIG_LABEL_DEF", "value": "failure-domain.beta.kubernetes.io/zone"},
							},
						}}}}},
	})

	_, err = clientset.AppsV1().DaemonSets("kube-system").Patch("aws-node", types.StrategicMergePatchType, patch)

	return err
}

// create the ENIConfig CRD
func (r *Reconciler) createENIConfigCRD(config rest.Config) error {

	clientset, err := apiextension.NewForConfig(&config)
	if err != nil {
		return err
	}

	// create the ENIConfig Custom Resource Definition
	eniConfigCRD := &apiextensionsv1beta1.CustomResourceDefinition{
		ObjectMeta: metav1.ObjectMeta{
			Name: "eniconfigs.crd.k8s.amazonaws.com",
		},
		Spec: apiextensionsv1beta1.CustomResourceDefinitionSpec{
			Group:   "crd.k8s.amazonaws.com",
			Version: "v1alpha1",
			Scope:   apiextensionsv1beta1.ClusterScoped,
			Names: apiextensionsv1beta1.CustomResourceDefinitionNames{
				Plural:   "eniconfigs",
				Singular: "eniconfig",
				Kind:     "ENIConfig",
			},
		},
	}

	_, err = clientset.ApiextensionsV1beta1().CustomResourceDefinitions().Create(eniConfigCRD)
	if err != nil {
		if apierrors.IsAlreadyExists(err) {
			r.log.Debug("clientset", "IsAlreadyExists", err)
			err = nil
			// _, err = clientset.ApiextensionsV1beta1().CustomResourceDefinitions().Update(crd)
			// if err != nil {
			// 	r.log.Debug("Kein", "Update", err)
			// 	return err
			// }
		}
	}
	return err
}

// create a dynamic client to apply the ENIConfig Custom Resources
func (r *Reconciler) createENIConfigCRs(instance *awscomputev1alpha3.EKSCluster, client eks.Client, config rest.Config) error {
	dynClient, err := dynamic.NewForConfig(&config)
	if err != nil {
		return err
	}

	var (
		eniConfigCRs = []unstructured.Unstructured{}
		eniConfigGVR = schema.GroupVersionResource{
			Group:    "crd.k8s.amazonaws.com",
			Version:  "v1alpha1",
			Resource: "eniconfigs",
		}
	)

	// get Subnet/AvailabilityZone tupel for use in ENIConfig CR
	subnetAZs, err := client.GetSubnetZone(instance.Spec.CustomSubnetIDs)
	if err != nil {
		return err
	}

	// get workernode security group for use in ENIConfig CR
	var workerSecurityGroup string
	if instance.Status.CloudFormationStackID != "" {
		clusterWorker, err := client.GetWorkerNodes(instance.Status.CloudFormationStackID)
		if err != nil {
			return err
		}
		workerSecurityGroup = clusterWorker.WorkerSecurityGroup
	}

	// create unstructured ENIConfig object
	// aws-vpc-cni doesn't export a clientset yet
	// https://github.com/aws/amazon-vpc-cni-k8s/pull/227
	for availabilityZone, subnet := range subnetAZs {
		eniConfigCR := unstructured.Unstructured{
			Object: map[string]interface{}{
				"apiVersion": "crd.k8s.amazonaws.com/v1alpha1",
				"kind":       "ENIConfig",
				"metadata": map[string]interface{}{
					"name":      availabilityZone,
					"namespace": "default",
				},
				"spec": map[string]interface{}{
					"securityGroups": []string{workerSecurityGroup},
					"subnet":         subnet,
				},
			},
		}
		eniConfigCRs = append(eniConfigCRs, eniConfigCR)
	}

	// // create ENIConfig CRs
	// // this doesn't work with the dynamic client?
	// cr1 := &vpccni.ENIConfig{
	// 	TypeMeta: metav1.TypeMeta{
	// 		APIVersion: "crd.k8s.amazonaws.com/v1alpha1",
	// 		Kind:       "ENIConfig",
	// 	},
	// 	ObjectMeta: metav1.ObjectMeta{
	// 		Name: "eu-west-1a",
	// 	    Namespace: "default",
	// 	},
	// 	Spec: vpccni.ENIConfigSpec{
	// 		SecurityGroups: []string{},
	// 		Subnet:         "subnet-026c046bb4fbd1468",
	// 	},
	// }

	crdClient := dynClient.Resource(eniConfigGVR)
	for _, eniConfig := range eniConfigCRs {
		eniConfig := eniConfig
		_, err = crdClient.Create(&eniConfig, metav1.CreateOptions{})
		if err != nil {
			if apierrors.IsAlreadyExists(err) {
				r.log.Debug("ENIConfig", "IsAlreadyExists", eniConfig.GetName)
				// TODO
				// get object and patch or overwrite
				// _, err = crdClient.Update(&cr, metav1.UpdateOptions{})
				err = nil
			}
		}
	}
	return err
}

func (r *Reconciler) _sync(instance *awscomputev1alpha3.EKSCluster, client eks.Client) (reconcile.Result, error) { // nolint:gocyclo
	cluster, err := client.Get(instance.Status.ClusterName)
	if err != nil {
		return r.fail(instance, err)
	}

	if cluster.Status != awscomputev1alpha3.ClusterStatusActive {
		instance.Status.SetConditions(runtimev1alpha1.ReconcileSuccess())

		// Requeue after a short wait to see if the cluster has become ready.
		return reconcile.Result{RequeueAfter: aShortWait}, nil
	}

	// Create workers
	if instance.Status.CloudFormationStackID == "" {
		clusterWorkers, err := client.CreateWorkerNodes(instance.Status.ClusterName, instance.Status.ClusterVersion, instance.Spec)
		if err != nil {
			return r.fail(instance, err)
		}
		instance.Status.CloudFormationStackID = clusterWorkers.WorkerStackID
		instance.Status.SetConditions(runtimev1alpha1.ReconcileSuccess())

		// We'll likely be requeued implicitly due to the status update, but
		// otherwise we want to requeue a reconcile after a short wait to check
		// on the worker node creation.
		return reconcile.Result{RequeueAfter: aShortWait}, r.Update(ctx, instance)
	}

	clusterWorker, err := client.GetWorkerNodes(instance.Status.CloudFormationStackID)
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

	if err := r.customnetwork(cluster, instance, client); err != nil {
		return r.fail(instance, errors.Wrap(err, "failed to set custom network on eks"))
	}

	if err := r.secret(cluster, instance, client); err != nil {
		return r.fail(instance, err)
	}

	// update resource status
	instance.Status.Endpoint = cluster.Endpoint
	instance.Status.State = awscomputev1alpha3.ClusterStatusActive
	instance.Status.SetConditions(runtimev1alpha1.Available(), runtimev1alpha1.ReconcileSuccess())
	resource.SetBindable(instance)

	// Our cluster is available. Requeue speculatively after a long wait in case
	// the cluster has changed.
	return reconcile.Result{RequeueAfter: aLongWait}, r.Update(ctx, instance)
}

func (r *Reconciler) _secret(cluster *eks.Cluster, instance *awscomputev1alpha3.EKSCluster, client eks.Client) error {
	token, err := client.ConnectionToken(instance.Status.ClusterName)
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
		if err := client.Delete(instance.Status.ClusterName); err != nil && !eks.IsErrorNotFound(err) {
			deleteErrors = append(deleteErrors, fmt.Sprintf("Master Delete Error: %s", err.Error()))
		}

		if instance.Status.CloudFormationStackID != "" {
			if err := client.DeleteWorkerNodes(instance.Status.CloudFormationStackID); err != nil && !cloudformationclient.IsErrorNotFound(err) {
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

	// Create EKS Client
	eksClient, err := r.connect(instance)
	if err != nil {
		return r.fail(instance, err)
	}

	if !resource.IsConditionTrue(instance.GetCondition(runtimev1alpha1.TypeReferencesResolved)) {
		if err := r.ResolveReferences(ctx, instance); err != nil {
			condition := runtimev1alpha1.ReconcileError(err)
			if managed.IsReferencesAccessError(err) {
				condition = runtimev1alpha1.ReferenceResolutionBlocked(err)
			}

			instance.Status.SetConditions(condition)
			return reconcile.Result{RequeueAfter: aLongWait}, errors.Wrap(r.Update(ctx, instance), errUpdateManagedStatus)
		}

		// Add ReferenceResolutionSuccess to the conditions
		instance.Status.SetConditions(runtimev1alpha1.ReferenceResolutionSuccess())
	}

	// Add finalizer
	meta.AddFinalizer(instance, finalizer)

	// Check for deletion
	if instance.DeletionTimestamp != nil {
		return r.delete(instance, eksClient)
	}

	// Create cluster instance
	if instance.Status.ClusterName == "" {
		return r.create(instance, eksClient)
	}

	// Sync cluster instance status with cluster status
	return r.sync(instance, eksClient)
}
