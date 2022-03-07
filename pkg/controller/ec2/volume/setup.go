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

package volume

import (
	"context"
	"time"

	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"k8s.io/client-go/util/workqueue"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/controller"

	svcsdk "github.com/aws/aws-sdk-go/service/ec2"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/logging"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/ratelimiter"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"

	svcapitypes "github.com/crossplane/provider-aws/apis/ec2/v1alpha1"
	awsclients "github.com/crossplane/provider-aws/pkg/clients"
)

// SetupVolume adds a controller that reconciles Volume.
func SetupVolume(mgr ctrl.Manager, l logging.Logger, limiter workqueue.RateLimiter, poll time.Duration) error {
	name := managed.ControllerName(svcapitypes.VolumeGroupKind)
	opts := []option{
		func(e *external) {
			e.preCreate = preCreate
			e.postCreate = postCreate
			e.postObserve = postObserve
			e.filterList = filterList
		},
	}
	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(controller.Options{
			RateLimiter: ratelimiter.NewController(limiter),
		}).
		For(&svcapitypes.Volume{}).
		Complete(managed.NewReconciler(mgr,
			resource.ManagedKind(svcapitypes.VolumeGroupVersionKind),
			managed.WithInitializers(managed.NewDefaultProviderConfig(mgr.GetClient())),
			managed.WithExternalConnecter(&connector{kube: mgr.GetClient(), opts: opts}),
			managed.WithPollInterval(poll),
			managed.WithLogger(l.WithValues("controller", name)),
			managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name)))))
}

func filterList(cr *svcapitypes.Volume, obj *svcsdk.DescribeVolumesOutput) *svcsdk.DescribeVolumesOutput {
	volumeIdentifier := awsclients.String(meta.GetExternalName(cr))
	resp := &svcsdk.DescribeVolumesOutput{}
	for _, volume := range obj.Volumes {
		if awsclients.StringValue(volume.VolumeId) == awsclients.StringValue(volumeIdentifier) {
			resp.Volumes = append(resp.Volumes, volume)
			break
		}
	}
	return resp
}

func preCreate(_ context.Context, cr *svcapitypes.Volume, obj *svcsdk.CreateVolumeInput) error {
	obj.KmsKeyId = cr.Spec.ForProvider.KMSKeyID
	obj.ClientToken = awsclients.String(string(cr.UID))
	return nil
}

func postCreate(_ context.Context, cr *svcapitypes.Volume, obj *svcsdk.Volume, cre managed.ExternalCreation, err error) (managed.ExternalCreation, error) {
	if err != nil {
		return managed.ExternalCreation{}, err
	}
	meta.SetExternalName(cr, awsclients.StringValue(obj.VolumeId))
	return managed.ExternalCreation{ExternalNameAssigned: true}, nil
}

func postObserve(_ context.Context, cr *svcapitypes.Volume, obj *svcsdk.DescribeVolumesOutput, obs managed.ExternalObservation, err error) (managed.ExternalObservation, error) {
	if err != nil {
		return managed.ExternalObservation{}, err
	}

	switch awsclients.StringValue(obj.Volumes[0].State) {
	case string(svcapitypes.VolumeState_available):
		cr.SetConditions(xpv1.Available())
	case string(svcapitypes.VolumeState_creating):
		cr.SetConditions(xpv1.Creating())
	case string(svcapitypes.VolumeState_error):
		cr.SetConditions(xpv1.Unavailable())
	case string(svcapitypes.VolumeState_deleting):
		cr.SetConditions(xpv1.Deleting())
	}

	obs.ConnectionDetails = managed.ConnectionDetails{
		"volumeID": []byte(awsclients.StringValue(obj.Volumes[0].VolumeId)),
	}
	return obs, nil
}
