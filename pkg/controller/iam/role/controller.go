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

package role

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsiam "github.com/aws/aws-sdk-go-v2/service/iam"
	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/connection"
	"github.com/crossplane/crossplane-runtime/pkg/controller"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/crossplane-contrib/provider-aws/apis/iam/v1beta1"
	"github.com/crossplane-contrib/provider-aws/apis/v1alpha1"
	"github.com/crossplane-contrib/provider-aws/pkg/clients/iam"
	"github.com/crossplane-contrib/provider-aws/pkg/features"
	connectaws "github.com/crossplane-contrib/provider-aws/pkg/utils/connect/aws"
	errorutils "github.com/crossplane-contrib/provider-aws/pkg/utils/errors"
	custommanaged "github.com/crossplane-contrib/provider-aws/pkg/utils/reconciler/managed"
)

const (
	errUnexpectedObject = "The managed resource is not an Role resource"
	errGet              = "failed to get Role with name"
	errCreate           = "failed to create the Role resource"
	errDelete           = "failed to delete the Role resource"
	errUpdate           = "failed to update the Role resource"
	errSDK              = "empty Role received from IAM API"
	errCreatePatch      = "failed to create patch object for comparison"

	errKubeUpdateFailed = "cannot late initialize Role"
	errUpToDateFailed   = "cannot check whether object is up-to-date"
)

// SetupRole adds a controller that reconciles Roles.
func SetupRole(mgr ctrl.Manager, o controller.Options) error {
	name := managed.ControllerName(v1beta1.RoleGroupKind)

	cps := []managed.ConnectionPublisher{managed.NewAPISecretPublisher(mgr.GetClient(), mgr.GetScheme())}
	if o.Features.Enabled(features.EnableAlphaExternalSecretStores) {
		cps = append(cps, connection.NewDetailsManager(mgr.GetClient(), v1alpha1.StoreConfigGroupVersionKind))
	}

	reconcilerOpts := []managed.ReconcilerOption{
		managed.WithCriticalAnnotationUpdater(custommanaged.NewRetryingCriticalAnnotationUpdater(mgr.GetClient())),
		managed.WithExternalConnecter(&connector{kube: mgr.GetClient(), newClientFn: iam.NewRoleClient}),
		managed.WithReferenceResolver(managed.NewAPISimpleReferenceResolver(mgr.GetClient())),
		managed.WithConnectionPublishers(),
		managed.WithPollInterval(o.PollInterval),
		managed.WithLogger(o.Logger.WithValues("controller", name)),
		managed.WithInitializers(managed.NewNameAsExternalName(mgr.GetClient())),
		managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name))),
		managed.WithConnectionPublishers(cps...),
	}

	if o.Features.Enabled(features.EnableAlphaManagementPolicies) {
		reconcilerOpts = append(reconcilerOpts, managed.WithManagementPolicies())
	}

	r := managed.NewReconciler(mgr,
		resource.ManagedKind(v1beta1.RoleGroupVersionKind),
		reconcilerOpts...)

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(o.ForControllerRuntime()).
		WithEventFilter(resource.DesiredStateChanged()).
		For(&v1beta1.Role{}).
		Complete(r)
}

type connector struct {
	kube        client.Client
	newClientFn func(config aws.Config) iam.RoleClient
}

func (c *connector) Connect(ctx context.Context, mg resource.Managed) (managed.ExternalClient, error) {
	cfg, err := connectaws.GetConfig(ctx, c.kube, mg, connectaws.GlobalRegion)
	if err != nil {
		return nil, err
	}
	return &external{client: c.newClientFn(*cfg), kube: c.kube}, nil
}

type external struct {
	client iam.RoleClient
	kube   client.Client
}

func (e *external) Observe(ctx context.Context, mgd resource.Managed) (managed.ExternalObservation, error) {
	cr, ok := mgd.(*v1beta1.Role)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errUnexpectedObject)
	}

	observed, err := e.client.GetRole(ctx, &awsiam.GetRoleInput{
		RoleName: aws.String(meta.GetExternalName(cr)),
	})

	if err != nil {
		return managed.ExternalObservation{}, errorutils.Wrap(resource.Ignore(iam.IsErrorNotFound, err), errGet)
	}

	if observed.Role == nil {
		return managed.ExternalObservation{}, errors.New(errSDK)
	}

	role := *observed.Role
	current := cr.Spec.ForProvider.DeepCopy()
	iam.LateInitializeRole(&cr.Spec.ForProvider, &role)
	if !cmp.Equal(current, &cr.Spec.ForProvider) {
		if err := e.kube.Update(ctx, cr); err != nil {
			return managed.ExternalObservation{}, errors.Wrap(err, errKubeUpdateFailed)
		}
	}

	cr.SetConditions(xpv1.Available())

	cr.Status.AtProvider = iam.GenerateRoleObservation(*observed.Role)

	upToDate, diff, err := iam.IsRoleUpToDate(cr.Spec.ForProvider, role)
	if err != nil {
		return managed.ExternalObservation{}, errors.Wrap(err, errUpToDateFailed)
	}

	return managed.ExternalObservation{
		ResourceExists:   true,
		ResourceUpToDate: upToDate,
		ConnectionDetails: managed.ConnectionDetails{
			"arn": []byte(cr.Status.AtProvider.ARN),
		},
		Diff: diff,
	}, nil
}

func (e *external) Create(ctx context.Context, mgd resource.Managed) (managed.ExternalCreation, error) {
	cr, ok := mgd.(*v1beta1.Role)
	if !ok {
		return managed.ExternalCreation{}, errors.New(errUnexpectedObject)
	}

	cr.Status.SetConditions(xpv1.Creating())

	_, err := e.client.CreateRole(ctx, iam.GenerateCreateRoleInput(meta.GetExternalName(cr), &cr.Spec.ForProvider))
	return managed.ExternalCreation{}, errorutils.Wrap(err, errCreate)
}

func (e *external) Update(ctx context.Context, mgd resource.Managed) (managed.ExternalUpdate, error) { //nolint:gocyclo
	cr, ok := mgd.(*v1beta1.Role)
	if !ok {
		return managed.ExternalUpdate{}, errors.New(errUnexpectedObject)
	}

	observed, err := e.client.GetRole(ctx, &awsiam.GetRoleInput{
		RoleName: aws.String(meta.GetExternalName(cr)),
	})

	if err != nil {
		return managed.ExternalUpdate{}, errorutils.Wrap(resource.Ignore(iam.IsErrorNotFound, err), errGet)
	}
	if observed.Role == nil {
		return managed.ExternalUpdate{}, errors.New(errSDK)
	}

	crTagMap := make(map[string]string, len(cr.Spec.ForProvider.Tags))
	for _, v := range cr.Spec.ForProvider.Tags {
		crTagMap[v.Key] = v.Value
	}

	add, remove, _ := iam.DiffIAMTags(crTagMap, observed.Role.Tags)
	if len(remove) != 0 {
		if _, err := e.client.UntagRole(ctx, &awsiam.UntagRoleInput{
			RoleName: aws.String(meta.GetExternalName(cr)),
			TagKeys:  remove,
		}); err != nil {
			return managed.ExternalUpdate{}, errorutils.Wrap(err, "cannot untag")
		}
	}
	if len(add) != 0 {
		if _, err := e.client.TagRole(ctx, &awsiam.TagRoleInput{
			RoleName: aws.String(meta.GetExternalName(cr)),
			Tags:     add,
		}); err != nil {
			return managed.ExternalUpdate{}, errorutils.Wrap(err, "cannot tag")
		}
	}

	patch, err := iam.CreatePatch(observed.Role, &cr.Spec.ForProvider)
	if err != nil {
		return managed.ExternalUpdate{}, errors.Wrap(err, errCreatePatch)
	}

	if patch.Description != nil || patch.MaxSessionDuration != nil {
		_, err = e.client.UpdateRole(ctx, &awsiam.UpdateRoleInput{
			RoleName:           aws.String(meta.GetExternalName(cr)),
			Description:        cr.Spec.ForProvider.Description,
			MaxSessionDuration: cr.Spec.ForProvider.MaxSessionDuration,
		})

		if err != nil {
			return managed.ExternalUpdate{}, errorutils.Wrap(err, errUpdate)
		}
	}

	boundaryArn := ""
	if observed.Role.PermissionsBoundary != nil {
		boundaryArn = *observed.Role.PermissionsBoundary.PermissionsBoundaryArn
	}
	if aws.ToString(&boundaryArn) != aws.ToString(cr.Spec.ForProvider.PermissionsBoundary) {
		if aws.ToString(cr.Spec.ForProvider.PermissionsBoundary) == "" {
			_, err = e.client.DeleteRolePermissionsBoundary(ctx, &awsiam.DeleteRolePermissionsBoundaryInput{
				RoleName: aws.String(meta.GetExternalName(cr)),
			})
		} else {
			_, err = e.client.PutRolePermissionsBoundary(ctx, &awsiam.PutRolePermissionsBoundaryInput{
				PermissionsBoundary: cr.Spec.ForProvider.PermissionsBoundary,
				RoleName:            aws.String(meta.GetExternalName(cr)),
			})
		}

		if err != nil {
			return managed.ExternalUpdate{}, errorutils.Wrap(err, errUpdate)
		}
	}

	if patch.AssumeRolePolicyDocument != "" {
		_, err = e.client.UpdateAssumeRolePolicy(ctx, &awsiam.UpdateAssumeRolePolicyInput{
			PolicyDocument: &cr.Spec.ForProvider.AssumeRolePolicyDocument,
			RoleName:       aws.String(meta.GetExternalName(cr)),
		})
		if err != nil {
			return managed.ExternalUpdate{}, errorutils.Wrap(err, errUpdate)
		}
	}
	return managed.ExternalUpdate{}, nil
}

func (e *external) Delete(ctx context.Context, mgd resource.Managed) error {
	cr, ok := mgd.(*v1beta1.Role)
	if !ok {
		return errors.New(errUnexpectedObject)
	}

	cr.Status.SetConditions(xpv1.Deleting())

	_, err := e.client.DeleteRole(ctx, &awsiam.DeleteRoleInput{
		RoleName: aws.String(meta.GetExternalName(cr)),
	})

	return errorutils.Wrap(resource.Ignore(iam.IsErrorNotFound, err), errDelete)
}
