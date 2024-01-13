/*
Copyright 2021 The Crossplane Authors.
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
	"errors"
	"sort"
	"time"

	svcsdk "github.com/aws/aws-sdk-go/service/glue"
	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/connection"
	"github.com/crossplane/crossplane-runtime/pkg/controller"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"

	svcapitypes "github.com/crossplane-contrib/provider-aws/apis/glue/v1alpha1"
	"github.com/crossplane-contrib/provider-aws/apis/v1alpha1"
	"github.com/crossplane-contrib/provider-aws/pkg/features"
	"github.com/crossplane-contrib/provider-aws/pkg/utils/pointer"
	custommanaged "github.com/crossplane-contrib/provider-aws/pkg/utils/reconciler/managed"
)

const (
	errUpdateNoPrincipal    = "cannot update Glue Database. Missing principal in createTableDefaultPermissions entry"
	errCreateSameIdentifier = "cannot create Glue Database. Combine permissions for same principals under one createTableDefaultPermissions entry"
)

// SetupDatabase adds a controller that reconciles Database.
func SetupDatabase(mgr ctrl.Manager, o controller.Options) error {
	name := managed.ControllerName(svcapitypes.DatabaseGroupKind)
	opts := []option{
		func(e *external) {
			e.preObserve = preObserve
			e.postObserve = postObserve
			e.preDelete = preDelete
			e.postCreate = postCreate
			e.preCreate = preCreate
			e.lateInitialize = lateInitialize
			e.isUpToDate = isUpToDate
			e.preUpdate = preUpdate
		},
	}

	cps := []managed.ConnectionPublisher{managed.NewAPISecretPublisher(mgr.GetClient(), mgr.GetScheme())}
	if o.Features.Enabled(features.EnableAlphaExternalSecretStores) {
		cps = append(cps, connection.NewDetailsManager(mgr.GetClient(), v1alpha1.StoreConfigGroupVersionKind))
	}

	reconcilerOpts := []managed.ReconcilerOption{
		managed.WithCriticalAnnotationUpdater(custommanaged.NewRetryingCriticalAnnotationUpdater(mgr.GetClient())),
		managed.WithExternalConnecter(&connector{kube: mgr.GetClient(), opts: opts}),
		managed.WithPollInterval(o.PollInterval),
		managed.WithLogger(o.Logger.WithValues("controller", name)),
		managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name))),
		managed.WithConnectionPublishers(cps...),
	}

	if o.Features.Enabled(features.EnableAlphaManagementPolicies) {
		reconcilerOpts = append(reconcilerOpts, managed.WithManagementPolicies())
	}

	r := managed.NewReconciler(mgr,
		resource.ManagedKind(svcapitypes.DatabaseGroupVersionKind),
		reconcilerOpts...)

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(o.ForControllerRuntime()).
		WithEventFilter(resource.DesiredStateChanged()).
		For(&svcapitypes.Database{}).
		Complete(r)
}

func preDelete(_ context.Context, cr *svcapitypes.Database, obj *svcsdk.DeleteDatabaseInput) (bool, error) {
	obj.Name = pointer.ToOrNilIfZeroValue(meta.GetExternalName(cr))
	return false, nil
}

func preObserve(_ context.Context, cr *svcapitypes.Database, obj *svcsdk.GetDatabaseInput) error {
	obj.Name = pointer.ToOrNilIfZeroValue(meta.GetExternalName(cr))
	return nil
}

func postObserve(_ context.Context, cr *svcapitypes.Database, obj *svcsdk.GetDatabaseOutput, obs managed.ExternalObservation, err error) (managed.ExternalObservation, error) {
	if err != nil {
		return managed.ExternalObservation{}, err
	}

	cr.Status.AtProvider.CreateTime = fromTimePtr(obj.Database.CreateTime)

	cr.SetConditions(xpv1.Available())
	return obs, nil
}

func lateInitialize(spec *svcapitypes.DatabaseParameters, resp *svcsdk.GetDatabaseOutput) error {

	spec.CatalogID = pointer.LateInitialize(spec.CatalogID, resp.Database.CatalogId)

	if spec.CustomDatabaseInput == nil {
		spec.CustomDatabaseInput = &svcapitypes.CustomDatabaseInput{}
	}
	if spec.CustomDatabaseInput.CreateTableDefaultPermissions == nil {

		spec.CustomDatabaseInput.CreateTableDefaultPermissions = []*svcapitypes.PrincipalPermissions{}
		for _, createPerms := range resp.Database.CreateTableDefaultPermissions {
			specPrins := &svcapitypes.PrincipalPermissions{
				Permissions: createPerms.Permissions,
				Principal:   &svcapitypes.DataLakePrincipal{DataLakePrincipalIdentifier: createPerms.Principal.DataLakePrincipalIdentifier},
			}
			spec.CustomDatabaseInput.CreateTableDefaultPermissions = append(spec.CustomDatabaseInput.CreateTableDefaultPermissions, specPrins)
		}
	}

	return nil
}

func isUpToDate(_ context.Context, cr *svcapitypes.Database, resp *svcsdk.GetDatabaseOutput) (bool, string, error) {

	currentParams := customGenerateDatabase(resp).Spec.ForProvider

	// checks for isuptodate-state of CustomDatabaseInput.CreateTableDefaultPermissions
	if cr.Spec.ForProvider.CustomDatabaseInput != nil && currentParams.CustomDatabaseInput.CreateTableDefaultPermissions != nil {

		if len(cr.Spec.ForProvider.CustomDatabaseInput.CreateTableDefaultPermissions) != len(currentParams.CustomDatabaseInput.CreateTableDefaultPermissions) {
			// will also partly catch edgecase when user made 2+ entries for same principal (which AWS combines to one entry)
			// this will ensure no panic error, but open an endless update-loop until user fixes the edgecase in specs
			// -> error message for user info is thrown in (pre)create and (pre)update
			return false, "", nil
		}

		// sorting both just to be safe
		sort.SliceStable(currentParams.CustomDatabaseInput.CreateTableDefaultPermissions, func(i, j int) bool {
			return pointer.StringValue(currentParams.CustomDatabaseInput.CreateTableDefaultPermissions[i].Principal.DataLakePrincipalIdentifier) > pointer.StringValue(currentParams.CustomDatabaseInput.CreateTableDefaultPermissions[j].Principal.DataLakePrincipalIdentifier)
		})

		sort.SliceStable(cr.Spec.ForProvider.CustomDatabaseInput.CreateTableDefaultPermissions, func(i, j int) bool {
			return pointer.StringValue(cr.Spec.ForProvider.CustomDatabaseInput.CreateTableDefaultPermissions[i].Principal.DataLakePrincipalIdentifier) > pointer.StringValue(cr.Spec.ForProvider.CustomDatabaseInput.CreateTableDefaultPermissions[j].Principal.DataLakePrincipalIdentifier)
		})

		// check all CreateTableDefaultPermissions entries if they are uptodate
		for i, prins := range cr.Spec.ForProvider.CustomDatabaseInput.CreateTableDefaultPermissions {
			currPrins := currentParams.CustomDatabaseInput.CreateTableDefaultPermissions[i]

			// to avoid panic
			if prins.Principal == nil {
				return false, "", errors.New(errUpdateNoPrincipal)
			}
			// check if this entry is uptodate
			if pointer.StringValue(prins.Principal.DataLakePrincipalIdentifier) != pointer.StringValue(currPrins.Principal.DataLakePrincipalIdentifier) {
				// both should be sorted the same way, so if we land here that would mean that
				// at least one entry has been added to spec and one has been removed from spec
				// or aka one entry was simply changed /"updated"
				return false, "", nil
			}

			sortOpts := cmpopts.SortSlices(func(a, b *string) bool {
				return pointer.StringValue(a) < pointer.StringValue(b)
			})

			// check if the permissions of this entry are uptodate
			if diff := cmp.Diff(prins.Permissions, currPrins.Permissions, sortOpts, cmpopts.EquateEmpty()); diff != "" {
				return false, diff, nil
			}
		}
	}

	diff := cmp.Diff(cr.Spec.ForProvider, currentParams,
		cmpopts.IgnoreTypes(&xpv1.Reference{}, &xpv1.Selector{}, []xpv1.Reference{}),
		cmpopts.IgnoreFields(svcapitypes.DatabaseParameters{}, "Region"),
		cmpopts.IgnoreFields(svcapitypes.CustomDatabaseInput{}, "CreateTableDefaultPermissions"),
		cmpopts.EquateEmpty())
	return diff == "", diff, nil
}

func preUpdate(_ context.Context, cr *svcapitypes.Database, obj *svcsdk.UpdateDatabaseInput) error {
	obj.Name = pointer.ToOrNilIfZeroValue(meta.GetExternalName(cr))

	obj.DatabaseInput = &svcsdk.DatabaseInput{
		Description: cr.Spec.ForProvider.CustomDatabaseInput.Description,
		LocationUri: cr.Spec.ForProvider.CustomDatabaseInput.LocationURI,
		Name:        pointer.ToOrNilIfZeroValue(meta.GetExternalName(cr)),
		Parameters:  cr.Spec.ForProvider.CustomDatabaseInput.Parameters,
	}

	if cr.Spec.ForProvider.CustomDatabaseInput.CreateTableDefaultPermissions != nil {
		obj.DatabaseInput.CreateTableDefaultPermissions = make([]*svcsdk.PrincipalPermissions, len(cr.Spec.ForProvider.CustomDatabaseInput.CreateTableDefaultPermissions))
		for i, prins := range cr.Spec.ForProvider.CustomDatabaseInput.CreateTableDefaultPermissions {
			sdkPrins := &svcsdk.PrincipalPermissions{
				Permissions: make([]*string, len(prins.Permissions)),
			}

			copy(sdkPrins.Permissions, prins.Permissions)

			if prins.Principal != nil {
				sdkPrins.Principal = &svcsdk.DataLakePrincipal{
					DataLakePrincipalIdentifier: prins.Principal.DataLakePrincipalIdentifier,
				}

				// handle case when user made 2+ createTableDefaultpermissions entries with same dataLakePrincipalIdentifier?
				// bc AWS will auto-combine these into 1 createTableDefaultpermissions entry

				// check for identical previous PrincipalIdentifiers
				for i3 := 0; i3 < i; i3++ {
					if obj.DatabaseInput.CreateTableDefaultPermissions[i3].Principal != nil &&
						pointer.StringValue(obj.DatabaseInput.CreateTableDefaultPermissions[i3].Principal.DataLakePrincipalIdentifier) == pointer.StringValue(sdkPrins.Principal.DataLakePrincipalIdentifier) {
						return errors.New(errCreateSameIdentifier)
					}
				}
			}

			obj.DatabaseInput.CreateTableDefaultPermissions[i] = sdkPrins
		}
	}

	if cr.Spec.ForProvider.CustomDatabaseInput.TargetDatabase != nil {
		obj.DatabaseInput.TargetDatabase = &svcsdk.DatabaseIdentifier{
			CatalogId:    cr.Spec.ForProvider.CustomDatabaseInput.TargetDatabase.CatalogID,
			DatabaseName: cr.Spec.ForProvider.CustomDatabaseInput.TargetDatabase.DatabaseName,
		}
	}

	return nil
}

func postCreate(_ context.Context, cr *svcapitypes.Database, obj *svcsdk.CreateDatabaseOutput, _ managed.ExternalCreation, err error) (managed.ExternalCreation, error) {
	if err != nil {
		return managed.ExternalCreation{}, err
	}
	meta.SetExternalName(cr, cr.Name)
	return managed.ExternalCreation{}, nil
}

func preCreate(_ context.Context, cr *svcapitypes.Database, obj *svcsdk.CreateDatabaseInput) error {

	if cr.Spec.ForProvider.CustomDatabaseInput == nil {
		obj.DatabaseInput = &svcsdk.DatabaseInput{
			Name: pointer.ToOrNilIfZeroValue(meta.GetExternalName(cr)),
		}
	} else {
		obj.DatabaseInput = &svcsdk.DatabaseInput{
			Name:        pointer.ToOrNilIfZeroValue(meta.GetExternalName(cr)),
			Description: cr.Spec.ForProvider.CustomDatabaseInput.Description,
			LocationUri: cr.Spec.ForProvider.CustomDatabaseInput.LocationURI,
			Parameters:  cr.Spec.ForProvider.CustomDatabaseInput.Parameters,
		}

		if cr.Spec.ForProvider.CustomDatabaseInput.CreateTableDefaultPermissions != nil {
			obj.DatabaseInput.CreateTableDefaultPermissions = make([]*svcsdk.PrincipalPermissions, len(cr.Spec.ForProvider.CustomDatabaseInput.CreateTableDefaultPermissions))
			for i, prins := range cr.Spec.ForProvider.CustomDatabaseInput.CreateTableDefaultPermissions {
				sdkPrins := &svcsdk.PrincipalPermissions{
					Permissions: make([]*string, len(prins.Permissions)),
				}

				copy(sdkPrins.Permissions, prins.Permissions)

				if prins.Principal != nil {
					sdkPrins.Principal = &svcsdk.DataLakePrincipal{
						DataLakePrincipalIdentifier: prins.Principal.DataLakePrincipalIdentifier,
					}

					// handle case when user made 2+ createTableDefaultpermissions entries with same dataLakePrincipalIdentifier?
					// bc AWS will auto-combine these into 1 createTableDefaultpermissions entry

					// check for identical previous PrincipalIdentifiers
					for i3 := 0; i3 < i; i3++ {
						if obj.DatabaseInput.CreateTableDefaultPermissions[i3].Principal != nil &&
							pointer.StringValue(obj.DatabaseInput.CreateTableDefaultPermissions[i3].Principal.DataLakePrincipalIdentifier) == pointer.StringValue(sdkPrins.Principal.DataLakePrincipalIdentifier) {
							return errors.New(errCreateSameIdentifier)
						}
					}
				}

				obj.DatabaseInput.CreateTableDefaultPermissions[i] = sdkPrins
			}
		}

		if cr.Spec.ForProvider.CustomDatabaseInput.TargetDatabase != nil {
			obj.DatabaseInput.TargetDatabase = &svcsdk.DatabaseIdentifier{
				CatalogId:    cr.Spec.ForProvider.CustomDatabaseInput.TargetDatabase.CatalogID,
				DatabaseName: cr.Spec.ForProvider.CustomDatabaseInput.TargetDatabase.DatabaseName,
			}
		}
	}

	return nil
}

// Custom GenerateDatabase for isuptodate (the generated one in zz_conversion.go is missing too much)
func customGenerateDatabase(resp *svcsdk.GetDatabaseOutput) *svcapitypes.Database {

	cr := &svcapitypes.Database{}
	cr.Spec.ForProvider.CustomDatabaseInput = &svcapitypes.CustomDatabaseInput{}

	if resp.Database.CatalogId != nil {
		cr.Spec.ForProvider.CatalogID = resp.Database.CatalogId
	} else {
		cr.Spec.ForProvider.CatalogID = nil
	}

	if resp.Database.CreateTableDefaultPermissions != nil {
		cr.Spec.ForProvider.CustomDatabaseInput.CreateTableDefaultPermissions = []*svcapitypes.PrincipalPermissions{}

		for _, createPerms := range resp.Database.CreateTableDefaultPermissions {

			specPrins := &svcapitypes.PrincipalPermissions{
				Permissions: createPerms.Permissions,
				Principal:   &svcapitypes.DataLakePrincipal{DataLakePrincipalIdentifier: createPerms.Principal.DataLakePrincipalIdentifier},
			}
			cr.Spec.ForProvider.CustomDatabaseInput.CreateTableDefaultPermissions = append(cr.Spec.ForProvider.CustomDatabaseInput.CreateTableDefaultPermissions, specPrins)
		}
	} else {
		cr.Spec.ForProvider.CustomDatabaseInput.CreateTableDefaultPermissions = nil
	}

	if resp.Database.Description != nil {
		cr.Spec.ForProvider.CustomDatabaseInput.Description = resp.Database.Description
	} else {
		cr.Spec.ForProvider.CustomDatabaseInput.Description = nil
	}

	if resp.Database.LocationUri != nil {
		cr.Spec.ForProvider.CustomDatabaseInput.LocationURI = resp.Database.LocationUri
	} else {
		cr.Spec.ForProvider.CustomDatabaseInput.LocationURI = nil
	}

	if resp.Database.Parameters != nil {
		cr.Spec.ForProvider.CustomDatabaseInput.Parameters = resp.Database.Parameters
	} else {
		cr.Spec.ForProvider.CustomDatabaseInput.Parameters = nil
	}

	if resp.Database.TargetDatabase != nil {
		cr.Spec.ForProvider.CustomDatabaseInput.TargetDatabase = &svcapitypes.DatabaseIdentifier{
			CatalogID:    resp.Database.TargetDatabase.CatalogId,
			DatabaseName: resp.Database.TargetDatabase.DatabaseName,
		}
	} else {
		cr.Spec.ForProvider.CustomDatabaseInput.TargetDatabase = nil
	}

	return cr
}

// fromTimePtr is a helper for converting a *time.Time to a *metav1.Time
func fromTimePtr(t *time.Time) *metav1.Time {
	if t != nil {
		m := metav1.NewTime(*t)
		return &m
	}
	return nil
}
