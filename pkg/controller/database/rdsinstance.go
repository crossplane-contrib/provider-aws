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

package database

import (
	"context"
	"reflect"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsrds "github.com/aws/aws-sdk-go-v2/service/rds"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	runtimev1alpha1 "github.com/crossplane/crossplane-runtime/apis/core/v1alpha1"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/logging"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/password"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"

	"github.com/crossplane/provider-aws/apis/database/v1beta1"
	awsv1alpha3 "github.com/crossplane/provider-aws/apis/v1alpha3"
	awsclients "github.com/crossplane/provider-aws/pkg/clients"
	"github.com/crossplane/provider-aws/pkg/clients/rds"
)

const (
	errNotRDSInstance   = "managed resource is not an RDS instance custom resource"
	errKubeUpdateFailed = "cannot update RDS instance custom resource"

	errCreateRDSClient   = "cannot create RDS client"
	errGetProvider       = "cannot get provider"
	errGetProviderSecret = "cannot get provider secret"

	errCreateFailed        = "cannot create RDS instance"
	errModifyFailed        = "cannot modify RDS instance"
	errAddTagsFailed       = "cannot add tags to RDS instance"
	errDeleteFailed        = "cannot delete RDS instance"
	errDescribeFailed      = "cannot describe RDS instance"
	errPatchCreationFailed = "cannot create a patch object"
	errUpToDateFailed      = "cannot check whether object is up-to-date"
)

// SetupRDSInstance adds a controller that reconciles RDSInstances.
func SetupRDSInstance(mgr ctrl.Manager, l logging.Logger) error {
	name := managed.ControllerName(v1beta1.RDSInstanceGroupKind)

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		For(&v1beta1.RDSInstance{}).
		Complete(managed.NewReconciler(mgr,
			resource.ManagedKind(v1beta1.RDSInstanceGroupVersionKind),
			managed.WithExternalConnecter(&connector{kube: mgr.GetClient(), newClientFn: rds.NewClient}),
			managed.WithInitializers(managed.NewNameAsExternalName(mgr.GetClient()), &tagger{kube: mgr.GetClient()}),
			managed.WithReferenceResolver(managed.NewAPISimpleReferenceResolver(mgr.GetClient())),
			managed.WithLogger(l.WithValues("controller", name)),
			managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name)))))
}

type connector struct {
	kube        client.Client
	newClientFn func(ctx context.Context, credentials []byte, region string, auth awsclients.AuthMethod) (rds.Client, error)
}

func (c *connector) Connect(ctx context.Context, mg resource.Managed) (managed.ExternalClient, error) {
	cr, ok := mg.(*v1beta1.RDSInstance)
	if !ok {
		return nil, errors.New(errNotRDSInstance)
	}

	p := &awsv1alpha3.Provider{}
	if err := c.kube.Get(ctx, meta.NamespacedNameOf(cr.Spec.ProviderReference), p); err != nil {
		return nil, errors.Wrap(err, errGetProvider)
	}

	if aws.BoolValue(p.Spec.UseServiceAccount) {
		rdsClient, err := c.newClientFn(ctx, []byte{}, p.Spec.Region, awsclients.UsePodServiceAccount)
		return &external{client: rdsClient, kube: c.kube}, errors.Wrap(err, errCreateRDSClient)
	}

	if p.GetCredentialsSecretReference() == nil {
		return nil, errors.New(errGetProviderSecret)
	}

	s := &corev1.Secret{}
	n := types.NamespacedName{Namespace: p.Spec.CredentialsSecretRef.Namespace, Name: p.Spec.CredentialsSecretRef.Name}
	if err := c.kube.Get(ctx, n, s); err != nil {
		return nil, errors.Wrap(err, errGetProviderSecret)
	}

	rdsClient, err := c.newClientFn(ctx, s.Data[p.Spec.CredentialsSecretRef.Key], p.Spec.Region, awsclients.UseProviderSecret)
	return &external{client: rdsClient, kube: c.kube}, errors.Wrap(err, errCreateRDSClient)
}

type external struct {
	client rds.Client
	kube   client.Client
}

func (e *external) Observe(ctx context.Context, mg resource.Managed) (managed.ExternalObservation, error) {
	cr, ok := mg.(*v1beta1.RDSInstance)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errNotRDSInstance)
	}
	// TODO(muvaf): There are some parameters that require a specific call
	// for retrieval. For example, DescribeDBInstancesOutput does not expose
	// the tags map of the RDS instance, you have to make ListTagsForResourceRequest
	req := e.client.DescribeDBInstancesRequest(&awsrds.DescribeDBInstancesInput{DBInstanceIdentifier: aws.String(meta.GetExternalName(cr))})
	rsp, err := req.Send(ctx)
	if err != nil {
		return managed.ExternalObservation{}, errors.Wrap(resource.Ignore(rds.IsErrorNotFound, err), errDescribeFailed)
	}

	// Describe requests can be used with filters, which then returns a list.
	// But we use an explicit identifier, so, if there is no error, there should
	// be only 1 element in the list.
	instance := rsp.DBInstances[0]
	current := cr.Spec.ForProvider.DeepCopy()
	rds.LateInitialize(&cr.Spec.ForProvider, &instance)
	if !reflect.DeepEqual(current, &cr.Spec.ForProvider) {
		if err := e.kube.Update(ctx, cr); err != nil {
			return managed.ExternalObservation{}, errors.Wrap(err, errKubeUpdateFailed)
		}
	}
	cr.Status.AtProvider = rds.GenerateObservation(instance)

	switch cr.Status.AtProvider.DBInstanceStatus {
	case v1beta1.RDSInstanceStateAvailable:
		cr.Status.SetConditions(runtimev1alpha1.Available())
		resource.SetBindable(cr)
	case v1beta1.RDSInstanceStateCreating:
		cr.Status.SetConditions(runtimev1alpha1.Creating())
	case v1beta1.RDSInstanceStateDeleting:
		cr.Status.SetConditions(runtimev1alpha1.Deleting())
	default:
		cr.Status.SetConditions(runtimev1alpha1.Unavailable())
	}
	upToDate, err := rds.IsUpToDate(cr.Spec.ForProvider, instance)
	if err != nil {
		return managed.ExternalObservation{}, errors.Wrap(err, errUpToDateFailed)
	}

	return managed.ExternalObservation{
		ResourceExists:    true,
		ResourceUpToDate:  upToDate,
		ConnectionDetails: rds.GetConnectionDetails(*cr),
	}, nil
}

func (e *external) Create(ctx context.Context, mg resource.Managed) (managed.ExternalCreation, error) {
	cr, ok := mg.(*v1beta1.RDSInstance)
	if !ok {
		return managed.ExternalCreation{}, errors.New(errNotRDSInstance)
	}
	cr.SetConditions(runtimev1alpha1.Creating())
	if cr.Status.AtProvider.DBInstanceStatus == v1beta1.RDSInstanceStateCreating {
		return managed.ExternalCreation{}, nil
	}
	pw, err := password.Generate()
	if err != nil {
		return managed.ExternalCreation{}, err
	}
	req := e.client.CreateDBInstanceRequest(rds.GenerateCreateDBInstanceInput(meta.GetExternalName(cr), pw, &cr.Spec.ForProvider))
	_, err = req.Send(ctx)
	if err != nil {
		return managed.ExternalCreation{}, errors.Wrap(err, errCreateFailed)
	}
	conn := managed.ConnectionDetails{
		runtimev1alpha1.ResourceCredentialsSecretPasswordKey: []byte(pw),
	}
	if cr.Spec.ForProvider.MasterUsername != nil {
		conn[runtimev1alpha1.ResourceCredentialsSecretUserKey] = []byte(aws.StringValue(cr.Spec.ForProvider.MasterUsername))
	}
	return managed.ExternalCreation{ConnectionDetails: conn}, nil
}

func (e *external) Update(ctx context.Context, mg resource.Managed) (managed.ExternalUpdate, error) {
	cr, ok := mg.(*v1beta1.RDSInstance)
	if !ok {
		return managed.ExternalUpdate{}, errors.New(errNotRDSInstance)
	}
	switch cr.Status.AtProvider.DBInstanceStatus {
	case v1beta1.RDSInstanceStateModifying, v1beta1.RDSInstanceStateCreating:
		return managed.ExternalUpdate{}, nil
	}
	// AWS rejects modification requests if you send fields whose value is same
	// as the current one. So, we have to create a patch out of the desired state
	// and the current state. Since the DBInstance is not fully mirrored in status,
	// we lose the current state after a change is made to spec, which forces us
	// to make a DescribeDBInstancesRequest to get the current state.
	describe := e.client.DescribeDBInstancesRequest(&awsrds.DescribeDBInstancesInput{DBInstanceIdentifier: aws.String(meta.GetExternalName(cr))})
	rsp, err := describe.Send(ctx)
	if err != nil {
		return managed.ExternalUpdate{}, errors.Wrap(err, errDescribeFailed)
	}
	patch, err := rds.CreatePatch(&rsp.DBInstances[0], &cr.Spec.ForProvider)
	if err != nil {
		return managed.ExternalUpdate{}, errors.Wrap(err, errPatchCreationFailed)
	}
	_, err = e.client.ModifyDBInstanceRequest(rds.GenerateModifyDBInstanceInput(meta.GetExternalName(cr), patch)).Send(ctx)
	if err != nil {
		return managed.ExternalUpdate{}, errors.Wrap(err, errModifyFailed)
	}
	if len(patch.Tags) > 0 {
		tags := make([]awsrds.Tag, len(patch.Tags))
		for i, t := range patch.Tags {
			tags[i] = awsrds.Tag{Key: aws.String(t.Key), Value: aws.String(t.Value)}
		}
		_, err = e.client.AddTagsToResourceRequest(&awsrds.AddTagsToResourceInput{
			ResourceName: aws.String(cr.Status.AtProvider.DBInstanceArn),
			Tags:         tags,
		}).Send(ctx)
		if err != nil {
			return managed.ExternalUpdate{}, errors.Wrap(err, errAddTagsFailed)
		}
	}
	return managed.ExternalUpdate{}, errors.Wrap(err, errModifyFailed)
}

func (e *external) Delete(ctx context.Context, mg resource.Managed) error {
	cr, ok := mg.(*v1beta1.RDSInstance)
	if !ok {
		return errors.New(errNotRDSInstance)
	}
	cr.SetConditions(runtimev1alpha1.Deleting())
	if cr.Status.AtProvider.DBInstanceStatus == v1beta1.RDSInstanceStateDeleting {
		return nil
	}
	// TODO(muvaf): There are cases where deletion results in an error that can
	// be solved only by a config change. But to do that, reconciler has to call
	// Update before Delete, which is not the case currently. In RDS, deletion
	// protection is an example for that and it's pretty common to use it. So,
	// until managed reconciler does Update before Delete, we do it here manually.
	// Update here is a best effort and deletion should not stop if it fails since
	// user may want to delete a resource whose fields are causing error.
	_, err := e.Update(ctx, cr)
	if rds.IsErrorNotFound(err) {
		return nil
	}

	input := awsrds.DeleteDBInstanceInput{
		DBInstanceIdentifier: aws.String(meta.GetExternalName(cr)),
		SkipFinalSnapshot:    cr.Spec.ForProvider.SkipFinalSnapshotBeforeDeletion,
	}
	_, err = e.client.DeleteDBInstanceRequest(&input).Send(ctx)
	return errors.Wrap(resource.Ignore(rds.IsErrorNotFound, err), errDeleteFailed)
}

type tagger struct {
	kube client.Client
}

func (t *tagger) Initialize(ctx context.Context, mg resource.Managed) error {
	cr, ok := mg.(*v1beta1.RDSInstance)
	if !ok {
		return errors.New(errNotRDSInstance)
	}
	tagMap := map[string]string{}
	for _, t := range cr.Spec.ForProvider.Tags {
		tagMap[t.Key] = t.Value
	}
	for k, v := range resource.GetExternalTags(mg) {
		tagMap[k] = v
	}
	cr.Spec.ForProvider.Tags = make([]v1beta1.Tag, len(tagMap))
	i := 0
	for k, v := range tagMap {
		cr.Spec.ForProvider.Tags[i] = v1beta1.Tag{Key: k, Value: v}
		i++
	}
	return errors.Wrap(t.kube.Update(ctx, cr), errKubeUpdateFailed)
}
