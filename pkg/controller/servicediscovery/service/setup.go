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

	// svcsdk "github.com/aws/aws-sdk-go-v2/service/servicediscovery"
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
			e.postDelete = postDelete
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

func getIdFromCR(cr *svcapitypes.Service) (*string, error) {
	arn := strings.SplitN(meta.GetExternalName(cr), "/", 2)
	if len(arn) != 2 {
		return nil, errors.New("external name has to be in the ARN format")
	}

	return aws.String(arn[1]), nil
}

func preObserve(_ context.Context, cr *svcapitypes.Service, obj *svcsdk.GetServiceInput) error {
	id, err := getIdFromCR(cr)
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
	obj.Service.DnsConfig = &svcsdk.DnsConfigChange{}
	newDnsConfig := []*svcsdk.DnsRecord{}
	for _, specDnsRecord := range cr.Spec.ForProvider.DNSConfig.DNSRecords {
		newDnsConfig = append(newDnsConfig, &svcsdk.DnsRecord{TTL: specDnsRecord.TTL, Type: specDnsRecord.Type})
	}
	obj.Service.DnsConfig.DnsRecords = newDnsConfig
	return nil
}

func postDelete(_ context.Context, cr *svcapitypes.Service, resp *svcsdk.DeleteServiceOutput, err error) error {

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

	if !isEqualDnsRecords(resp.Service.DnsConfig.DnsRecords, cr.Spec.ForProvider.DNSConfig.DNSRecords) {
		return false, nil
	}

	return true, nil
}

func isEqualDnsRecords(p1 []*svcsdk.DnsRecord, p2 []*svcapitypes.DNSRecord) bool {

	if len(p1) != len(p2) {
		return false
	}

	equals := false
	for _, outDnsRecord := range p1 {
		for _, crDnsRecord := range p2 {
			if *outDnsRecord.TTL == *crDnsRecord.TTL && *outDnsRecord.Type == *crDnsRecord.Type {
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
