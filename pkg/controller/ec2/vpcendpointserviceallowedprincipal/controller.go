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

package vpcendpointserviceallowedprincipal

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsec2 "github.com/aws/aws-sdk-go-v2/service/ec2"
	awsec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/connection"
	"github.com/crossplane/crossplane-runtime/pkg/controller"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	"github.com/pkg/errors"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/crossplane-contrib/provider-aws/apis/ec2/manualv1alpha1"
	"github.com/crossplane-contrib/provider-aws/apis/v1alpha1"
	"github.com/crossplane-contrib/provider-aws/pkg/features"
	connectaws "github.com/crossplane-contrib/provider-aws/pkg/utils/connect/aws"
	errorutils "github.com/crossplane-contrib/provider-aws/pkg/utils/errors"
	custommanaged "github.com/crossplane-contrib/provider-aws/pkg/utils/reconciler/managed"
)

const (
	errUnexpectedObject   = "The managed resource is not a VPCEndpointServiceAllowedPrincipal resource"
	errDescribe           = "failed to describe VPC endpoint service permissions"
	errModifyPermissions  = "failed to modify VPC endpoint service permissions"
	errCreateExternalName = "failed to create external name"
	errParseExternalName  = "failed to parse external name"
	errPrincipalNotFound  = "principal not found in VPC endpoint service permissions"
)

// SetupVPCEndpointServiceAllowedPrincipal adds a controller that reconciles VPCEndpointServiceAllowedPrincipals.
func SetupVPCEndpointServiceAllowedPrincipal(mgr ctrl.Manager, o controller.Options) error {
	name := managed.ControllerName(manualv1alpha1.VPCEndpointServiceAllowedPrincipalGroupKind)

	cps := []managed.ConnectionPublisher{managed.NewAPISecretPublisher(mgr.GetClient(), mgr.GetScheme())}
	if o.Features.Enabled(features.EnableAlphaExternalSecretStores) {
		cps = append(cps, connection.NewDetailsManager(mgr.GetClient(), v1alpha1.StoreConfigGroupVersionKind))
	}

	reconcilerOpts := []managed.ReconcilerOption{
		managed.WithCriticalAnnotationUpdater(custommanaged.NewRetryingCriticalAnnotationUpdater(mgr.GetClient())),
		managed.WithExternalConnecter(&connector{kube: mgr.GetClient(), newClientFn: newEC2Client}),
		managed.WithCreationGracePeriod(3 * time.Minute),
		managed.WithReferenceResolver(managed.NewAPISimpleReferenceResolver(mgr.GetClient())),
		managed.WithConnectionPublishers(),
		managed.WithInitializers(),
		managed.WithPollInterval(o.PollInterval),
		managed.WithLogger(o.Logger.WithValues("controller", name)),
		managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name))),
		managed.WithConnectionPublishers(cps...),
	}

	if o.Features.Enabled(features.EnableAlphaManagementPolicies) {
		reconcilerOpts = append(reconcilerOpts, managed.WithManagementPolicies())
	}

	r := managed.NewReconciler(mgr,
		resource.ManagedKind(manualv1alpha1.VPCEndpointServiceAllowedPrincipalGroupVersionKind),
		reconcilerOpts...)

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(o.ForControllerRuntime()).
		WithEventFilter(resource.DesiredStateChanged()).
		For(&manualv1alpha1.VPCEndpointServiceAllowedPrincipal{}).
		Complete(r)
}

// VPCEndpointServiceAllowedPrincipalClient interface for AWS EC2 client
type VPCEndpointServiceAllowedPrincipalClient interface {
	DescribeVpcEndpointServicePermissions(ctx context.Context, input *awsec2.DescribeVpcEndpointServicePermissionsInput, opts ...func(*awsec2.Options)) (*awsec2.DescribeVpcEndpointServicePermissionsOutput, error)
	ModifyVpcEndpointServicePermissions(ctx context.Context, input *awsec2.ModifyVpcEndpointServicePermissionsInput, opts ...func(*awsec2.Options)) (*awsec2.ModifyVpcEndpointServicePermissionsOutput, error)
}

func newEC2Client(config aws.Config) VPCEndpointServiceAllowedPrincipalClient {
	return awsec2.NewFromConfig(config)
}

type connector struct {
	kube        client.Client
	newClientFn func(config aws.Config) VPCEndpointServiceAllowedPrincipalClient
}

func (c *connector) Connect(ctx context.Context, mg resource.Managed) (managed.ExternalClient, error) {
	cr, ok := mg.(*manualv1alpha1.VPCEndpointServiceAllowedPrincipal)
	if !ok {
		return nil, errors.New(errUnexpectedObject)
	}
	cfg, err := connectaws.GetConfig(ctx, c.kube, mg, cr.Spec.ForProvider.Region)
	if err != nil {
		return nil, err
	}
	return &external{client: c.newClientFn(*cfg), kube: c.kube}, nil
}

type external struct {
	kube   client.Client
	client VPCEndpointServiceAllowedPrincipalClient
}

// generateExternalName creates a unique external name by combining service ID and principal ARN
func generateExternalName(serviceID, principalARN string) string {
	return fmt.Sprintf("%s:%s", serviceID, principalARN)
}

// parseExternalName parses the external name to extract service ID and principal ARN
func parseExternalName(externalName string) (serviceID, principalARN string, err error) {
	parts := strings.SplitN(externalName, ":", 2)
	if len(parts) != 2 {
		return "", "", errors.New("external name must be in format 'serviceID:principalARN'")
	}
	return parts[0], parts[1], nil
}

func (e *external) Observe(ctx context.Context, mgd resource.Managed) (managed.ExternalObservation, error) {
	cr, ok := mgd.(*manualv1alpha1.VPCEndpointServiceAllowedPrincipal)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errUnexpectedObject)
	}

	// If no external name, resource doesn't exist yet
	externalName := meta.GetExternalName(cr)
	if externalName == "" {
		return managed.ExternalObservation{
			ResourceExists: false,
		}, nil
	}

	// Parse the external name to get service ID and principal ARN
	_, principalARN, err := parseExternalName(externalName)
	if err != nil {
		return managed.ExternalObservation{}, errorutils.Wrap(err, errParseExternalName)
	}

	// Describe VPC endpoint service permissions
	response, err := e.client.DescribeVpcEndpointServicePermissions(ctx, &awsec2.DescribeVpcEndpointServicePermissionsInput{
		ServiceId: aws.String(cr.Spec.ForProvider.VPCEndpointServiceID),
		Filters: []awsec2types.Filter{
			{
				Name:   aws.String("principal"),
				Values: []string{principalARN},
			},
		},
	})

	if err != nil {
		return managed.ExternalObservation{
			ResourceExists: false,
		}, errorutils.Wrap(resource.Ignore(isVPCEndpointServiceNotFoundErr, err), errDescribe)
	}

	// Check if the principal is found in the allowed principals
	var foundPrincipal *awsec2types.AllowedPrincipal
	for _, principal := range response.AllowedPrincipals {
		if aws.ToString(principal.Principal) == principalARN {
			foundPrincipal = &principal
			break
		}
	}

	if foundPrincipal == nil {
		return managed.ExternalObservation{
			ResourceExists: false,
		}, nil
	}

	// Update status with observed values
	cr.Status.AtProvider = manualv1alpha1.VPCEndpointServiceAllowedPrincipalObservation{
		Principal:           foundPrincipal.Principal,
		PrincipalType:       aws.String(string(foundPrincipal.PrincipalType)),
		ServicePermissionID: foundPrincipal.ServicePermissionId,
		ServiceID:           aws.String(cr.Spec.ForProvider.VPCEndpointServiceID),
	}

	cr.SetConditions(xpv1.Available())

	return managed.ExternalObservation{
		ResourceExists:   true,
		ResourceUpToDate: true,
	}, nil
}

func (e *external) Create(ctx context.Context, mgd resource.Managed) (managed.ExternalCreation, error) {
	cr, ok := mgd.(*manualv1alpha1.VPCEndpointServiceAllowedPrincipal)
	if !ok {
		return managed.ExternalCreation{}, errors.New(errUnexpectedObject)
	}

	// Add the principal to the VPC endpoint service
	_, err := e.client.ModifyVpcEndpointServicePermissions(ctx, &awsec2.ModifyVpcEndpointServicePermissionsInput{
		ServiceId:            aws.String(cr.Spec.ForProvider.VPCEndpointServiceID),
		AddAllowedPrincipals: []string{cr.Spec.ForProvider.PrincipalARN},
	})
	if err != nil {
		return managed.ExternalCreation{}, errorutils.Wrap(err, errModifyPermissions)
	}

	// Set external name
	externalName := generateExternalName(cr.Spec.ForProvider.VPCEndpointServiceID, cr.Spec.ForProvider.PrincipalARN)
	meta.SetExternalName(cr, externalName)

	return managed.ExternalCreation{}, nil
}

func (e *external) Update(_ context.Context, _ resource.Managed) (managed.ExternalUpdate, error) {
	// VPC endpoint service allowed principals are immutable - they can only be added or removed
	return managed.ExternalUpdate{}, nil
}

func (e *external) Delete(ctx context.Context, mgd resource.Managed) (managed.ExternalDelete, error) {
	cr, ok := mgd.(*manualv1alpha1.VPCEndpointServiceAllowedPrincipal)
	if !ok {
		return managed.ExternalDelete{}, errors.New(errUnexpectedObject)
	}

	// Remove the principal from the VPC endpoint service
	_, err := e.client.ModifyVpcEndpointServicePermissions(ctx, &awsec2.ModifyVpcEndpointServicePermissionsInput{
		ServiceId:               aws.String(cr.Spec.ForProvider.VPCEndpointServiceID),
		RemoveAllowedPrincipals: []string{cr.Spec.ForProvider.PrincipalARN},
	})

	return managed.ExternalDelete{}, errorutils.Wrap(resource.Ignore(isVPCEndpointServiceNotFoundErr, err), errModifyPermissions)
}

func (e *external) Disconnect(_ context.Context) error {
	// Unimplemented, required by newer versions of crossplane-runtime
	return nil
}

// isVPCEndpointServiceNotFoundErr returns true if the error indicates that the VPC endpoint service was not found
func isVPCEndpointServiceNotFoundErr(err error) bool {
	// Check for AWS error codes that indicate the VPC endpoint service doesn't exist
	if err == nil {
		return false
	}

	// Common AWS error patterns for VPC endpoint service not found
	return strings.Contains(err.Error(), "InvalidVpcEndpointServiceId") ||
		strings.Contains(err.Error(), "VpcEndpointServiceNotFound") ||
		strings.Contains(err.Error(), "does not exist")
}
