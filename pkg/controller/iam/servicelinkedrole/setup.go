/*
Copyright 2022 The Crossplane Authors.
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

package servicelinkedrole

import (
	"context"
	"strings"

	svcsdk "github.com/aws/aws-sdk-go/service/iam"
	svcsdkapi "github.com/aws/aws-sdk-go/service/iam/iamiface"
	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/connection"
	"github.com/crossplane/crossplane-runtime/pkg/controller"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	"github.com/pkg/errors"
	ctrl "sigs.k8s.io/controller-runtime"

	svcapitypes "github.com/crossplane-contrib/provider-aws/apis/iam/v1alpha1"
	"github.com/crossplane-contrib/provider-aws/apis/v1alpha1"
	"github.com/crossplane-contrib/provider-aws/pkg/features"
	"github.com/crossplane-contrib/provider-aws/pkg/utils/arn"
	"github.com/crossplane-contrib/provider-aws/pkg/utils/pointer"
	custommanaged "github.com/crossplane-contrib/provider-aws/pkg/utils/reconciler/managed"
)

const (
	errNotServiceLinkedRole = "role is not a service-linked-role"
	errGetRole              = "cannot get role"
)

// SetupServiceLinkedRole adds a controller that reconciles ServiceLinkedRole.
func SetupServiceLinkedRole(mgr ctrl.Manager, o controller.Options) error {
	name := managed.ControllerName(svcapitypes.ServiceLinkedRoleGroupKind)
	opts := []option{
		func(e *external) {
			h := hooks{client: e.client}
			e.postCreate = postCreate
			e.preDelete = preDelete
			e.observe = h.observe
		},
	}

	cps := []managed.ConnectionPublisher{managed.NewAPISecretPublisher(mgr.GetClient(), mgr.GetScheme())}
	if o.Features.Enabled(features.EnableAlphaExternalSecretStores) {
		cps = append(cps, connection.NewDetailsManager(mgr.GetClient(), v1alpha1.StoreConfigGroupVersionKind))
	}

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(o.ForControllerRuntime()).
		WithEventFilter(resource.DesiredStateChanged()).
		For(&svcapitypes.ServiceLinkedRole{}).
		Complete(managed.NewReconciler(mgr,
			resource.ManagedKind(svcapitypes.ServiceLinkedRoleGroupVersionKind),
			managed.WithCriticalAnnotationUpdater(custommanaged.NewRetryingCriticalAnnotationUpdater(mgr.GetClient())),
			managed.WithExternalConnecter(&connector{kube: mgr.GetClient(), opts: opts}),
			managed.WithConnectionPublishers(),
			managed.WithPollInterval(o.PollInterval),
			managed.WithLogger(o.Logger.WithValues("controller", name)),
			managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name))),
			managed.WithConnectionPublishers(cps...)))
}

type hooks struct {
	client svcsdkapi.IAMAPI
}

func (e *hooks) observe(ctx context.Context, mg resource.Managed) (managed.ExternalObservation, error) {
	cr, ok := mg.(*svcapitypes.ServiceLinkedRole)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errUnexpectedObject)
	}

	if meta.GetExternalName(cr) == "" {
		return managed.ExternalObservation{}, nil
	}

	res, err := e.client.GetRoleWithContext(ctx, &svcsdk.GetRoleInput{
		RoleName: pointer.ToOrNilIfZeroValue(meta.GetExternalName(cr)),
	})
	if err != nil {
		return managed.ExternalObservation{}, errors.Wrap(resource.Ignore(IsNotFound, err), errGetRole)
	}

	cr.Status.AtProvider = generateServiceLinkedRoleObservation(res.Role)

	if err := isServiceLinkedRole(res.Role); err != nil {
		return managed.ExternalObservation{
			ResourceExists: true,
		}, err
	}

	cr.SetConditions(xpv1.Available())
	return managed.ExternalObservation{
		ResourceExists:   true,
		ResourceUpToDate: true,
	}, nil
}

func isServiceLinkedRole(role *svcsdk.Role) error {
	arn, err := arn.ParseARN(pointer.StringValue(role.Arn))
	if err != nil {
		return err
	}
	if !strings.HasPrefix(arn.Resource, "role/aws-service-role/") {
		return errors.New(errNotServiceLinkedRole)
	}
	return nil
}

func generateServiceLinkedRoleObservation(role *svcsdk.Role) svcapitypes.ServiceLinkedRoleObservation {
	o := svcapitypes.ServiceLinkedRoleObservation{
		ARN:                      role.Arn,
		AssumeRolePolicyDocument: role.AssumeRolePolicyDocument,
		CreateDate:               pointer.TimeToMetaTime(role.CreateDate),
		MaxSessionDuration:       role.MaxSessionDuration,
		Path:                     role.Path,
		RoleID:                   role.RoleId,
		RoleName:                 role.RoleName,
	}
	if role.PermissionsBoundary != nil {
		o.PermissionsBoundary = &svcapitypes.AttachedPermissionsBoundary{
			PermissionsBoundaryARN:  role.PermissionsBoundary.PermissionsBoundaryArn,
			PermissionsBoundaryType: role.PermissionsBoundary.PermissionsBoundaryType,
		}
	}
	if role.RoleLastUsed != nil {
		o.RoleLastUsed = &svcapitypes.RoleLastUsed{
			LastUsedDate: pointer.TimeToMetaTime(role.RoleLastUsed.LastUsedDate),
			Region:       role.RoleLastUsed.Region,
		}
	}
	o.Tags = make([]*svcapitypes.Tag, len(role.Tags))
	for i, t := range role.Tags {
		o.Tags[i] = &svcapitypes.Tag{
			Key:   t.Key,
			Value: t.Value,
		}
	}
	return o
}

func postCreate(ctx context.Context, cr *svcapitypes.ServiceLinkedRole, obj *svcsdk.CreateServiceLinkedRoleOutput, cre managed.ExternalCreation, err error) (managed.ExternalCreation, error) {
	if err != nil {
		return managed.ExternalCreation{}, err
	}

	nn := strings.Split(pointer.StringValue(obj.Role.Arn), "/")
	meta.SetExternalName(cr, nn[3])
	return managed.ExternalCreation{}, nil
}

func preDelete(ctx context.Context, cr *svcapitypes.ServiceLinkedRole, obj *svcsdk.DeleteServiceLinkedRoleInput) (bool, error) {
	obj.RoleName = pointer.ToOrNilIfZeroValue(meta.GetExternalName(cr))
	return false, nil
}
