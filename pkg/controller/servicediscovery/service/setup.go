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

package service

import (
	"context"
	"strings"

	ctrl "sigs.k8s.io/controller-runtime"

	"github.com/crossplane/crossplane-runtime/pkg/connection"
	"github.com/crossplane/crossplane-runtime/pkg/controller"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	"github.com/pkg/errors"

	"github.com/aws/aws-sdk-go/aws"
	svcsdk "github.com/aws/aws-sdk-go/service/servicediscovery"
	svcapitypes "github.com/crossplane-contrib/provider-aws/apis/servicediscovery/v1alpha1"
	"github.com/crossplane-contrib/provider-aws/apis/v1alpha1"
	"github.com/crossplane-contrib/provider-aws/pkg/features"
	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
)

// SetupService adds a controller that reconciles Service.
func SetupService(mgr ctrl.Manager, o controller.Options) error {
	name := managed.ControllerName(svcapitypes.ServiceGroupKind)
	opts := []option{
		func(e *external) {
			e.preObserve = preObserve
			e.postObserve = postObserve
			e.postCreate = postCreate
			e.preUpdate = preUpdate
			e.isUpToDate = isUpToDate
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
		For(&svcapitypes.Service{}).
		Complete(managed.NewReconciler(mgr,
			resource.ManagedKind(svcapitypes.ServiceGroupVersionKind),
			managed.WithExternalConnecter(&connector{kube: mgr.GetClient(), opts: opts}),
			managed.WithInitializers(),
			managed.WithPollInterval(o.PollInterval),
			managed.WithLogger(o.Logger.WithValues("controller", name)),
			managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name))),
			managed.WithConnectionPublishers(cps...)))
}

func getIDFromCR(cr *svcapitypes.Service) (*string, error) {
	arn := strings.SplitN(meta.GetExternalName(cr), "/", 2)
	if len(arn) != 2 {
		return nil, errors.New("external name has to be in the ARN format")
	}

	return aws.String(arn[1]), nil
}

func preObserve(_ context.Context, cr *svcapitypes.Service, obj *svcsdk.GetServiceInput) error {
	id, err := getIDFromCR(cr)
	if err != nil {
		return err
	}
	obj.Id = id
	return nil
}

func preUpdate(_ context.Context, cr *svcapitypes.Service, obj *svcsdk.UpdateServiceInput) error {

	if obj.Service == nil {
		obj.Service = &svcsdk.ServiceChange{}
	}
	obj.Service.Description = cr.Spec.ForProvider.Description

	if cr.Spec.ForProvider.DNSConfig != nil {
		obj.Service.DnsConfig = &svcsdk.DnsConfigChange{}
		newDNSConfig := []*svcsdk.DnsRecord{}
		for _, specDNSRecord := range cr.Spec.ForProvider.DNSConfig.DNSRecords {
			newDNSConfig = append(newDNSConfig, &svcsdk.DnsRecord{TTL: specDNSRecord.TTL, Type: specDNSRecord.Type})
		}
		obj.Service.DnsConfig.DnsRecords = newDNSConfig
	} else {
		cr.Spec.ForProvider.DNSConfig = nil
	}

	if cr.Spec.ForProvider.HealthCheckConfig != nil {
		if obj.Service.HealthCheckConfig == nil {
			obj.Service.HealthCheckConfig = &svcsdk.HealthCheckConfig{}
		}
		obj.Service.HealthCheckConfig.Type = cr.Spec.ForProvider.HealthCheckConfig.Type
		obj.Service.HealthCheckConfig.ResourcePath = cr.Spec.ForProvider.HealthCheckConfig.ResourcePath
		obj.Service.HealthCheckConfig.FailureThreshold = cr.Spec.ForProvider.HealthCheckConfig.FailureThreshold
	} else {
		obj.Service.HealthCheckConfig = nil
	}

	return nil
}

func postCreate(_ context.Context, cr *svcapitypes.Service, obj *svcsdk.CreateServiceOutput, cre managed.ExternalCreation, err error) (managed.ExternalCreation, error) {
	if err != nil {
		return managed.ExternalCreation{}, err
	}
	meta.SetExternalName(cr, aws.StringValue(obj.Service.Arn))
	return cre, nil
}

func postObserve(_ context.Context, cr *svcapitypes.Service, resp *svcsdk.GetServiceOutput, obs managed.ExternalObservation, err error) (managed.ExternalObservation, error) {
	if err != nil {
		return managed.ExternalObservation{}, err
	}
	cr.SetConditions(xpv1.Available())

	return obs, nil
}

func isUpToDate(cr *svcapitypes.Service, resp *svcsdk.GetServiceOutput) (bool, error) {
	if *resp.Service.Description != *cr.Spec.ForProvider.Description {
		return false, nil
	}

	if !isEqualDNSRecords(resp.Service.DnsConfig.DnsRecords, cr.Spec.ForProvider.DNSConfig.DNSRecords) {
		return false, nil
	}

	if !isEqualHealthCheckConfig(resp.Service.HealthCheckConfig, cr.Spec.ForProvider.HealthCheckConfig) {
		return false, nil
	}

	return true, nil
}

func isEqualHealthCheckConfig(outHealthCheck *svcsdk.HealthCheckConfig, crHealthCheck *svcapitypes.HealthCheckConfig) bool {
	if outHealthCheck == nil && crHealthCheck == nil {
		return true
	}

	if (outHealthCheck == nil && crHealthCheck != nil) || (crHealthCheck == nil && outHealthCheck != nil) {
		return false
	}

	if *outHealthCheck.Type != *crHealthCheck.Type || *outHealthCheck.ResourcePath != *crHealthCheck.ResourcePath || *outHealthCheck.FailureThreshold != *crHealthCheck.FailureThreshold {
		return false
	}

	return true
}

func isEqualDNSRecords(outDNSRecords []*svcsdk.DnsRecord, crDNSRecords []*svcapitypes.DNSRecord) bool {

	if len(outDNSRecords) != len(crDNSRecords) {
		return false
	}

	equals := false
	for _, outDNSRecord := range outDNSRecords {
		for _, crDNSRecord := range crDNSRecords {
			if *outDNSRecord.TTL == *crDNSRecord.TTL && *outDNSRecord.Type == *crDNSRecord.Type {
				equals = true
				break
			}
		}
		if !equals {
			return false
		}
		equals = false
	}

	return true
}
