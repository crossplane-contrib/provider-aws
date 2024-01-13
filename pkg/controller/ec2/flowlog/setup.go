package flowlog

import (
	"context"

	svcsdk "github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/ec2/ec2iface"
	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/connection"
	"github.com/crossplane/crossplane-runtime/pkg/controller"
	"github.com/crossplane/crossplane-runtime/pkg/errors"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	cpresource "github.com/crossplane/crossplane-runtime/pkg/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"

	svcapitypes "github.com/crossplane-contrib/provider-aws/apis/ec2/v1alpha1"
	"github.com/crossplane-contrib/provider-aws/apis/v1alpha1"
	"github.com/crossplane-contrib/provider-aws/pkg/features"
	errorutils "github.com/crossplane-contrib/provider-aws/pkg/utils/errors"
	"github.com/crossplane-contrib/provider-aws/pkg/utils/pointer"
	custommanaged "github.com/crossplane-contrib/provider-aws/pkg/utils/reconciler/managed"
)

var (
	_ = &metav1.Time{}
)

const (
	errUpdateTags      = "cannot update tags"
	flowLogTagResource = "vpc-flow-log"
)

type updater struct {
	client ec2iface.EC2API
}
type deleter struct {
	client ec2iface.EC2API
}

// SetupFlowLog adds a controller that reconciles FlowLog
func SetupFlowLog(mgr ctrl.Manager, o controller.Options) error {
	name := managed.ControllerName(svcapitypes.FlowLogGroupKind)
	opts := []option{
		func(e *external) {
			e.preCreate = preCreate
			e.preObserve = preObserve
			e.filterList = filterList
			e.postObserve = postObserve
			e.postCreate = postCreate
			u := &updater{client: e.client}
			e.isUpToDate = u.isUpToDate
			e.update = u.update
			d := &deleter{client: e.client}
			e.delete = d.delete
		},
	}

	cps := []managed.ConnectionPublisher{managed.NewAPISecretPublisher(mgr.GetClient(), mgr.GetScheme())}
	if o.Features.Enabled(features.EnableAlphaExternalSecretStores) {
		cps = append(cps, connection.NewDetailsManager(mgr.GetClient(), v1alpha1.StoreConfigGroupVersionKind))
	}

	reconcilerOpts := []managed.ReconcilerOption{
		managed.WithCriticalAnnotationUpdater(custommanaged.NewRetryingCriticalAnnotationUpdater(mgr.GetClient())),
		managed.WithExternalConnecter(&connector{kube: mgr.GetClient(), opts: opts}),
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
		resource.ManagedKind(svcapitypes.FlowLogGroupVersionKind),
		reconcilerOpts...)

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(o.ForControllerRuntime()).
		WithEventFilter(resource.DesiredStateChanged()).
		For(&svcapitypes.FlowLog{}).
		Complete(r)

}

func determineResourceIdsAndType(cr *svcapitypes.FlowLog) ([]*string, *string) {
	vpcResourceType := svcsdk.FlowLogsResourceTypeVpc
	transitGatewayResourceType := "TransitGateway"
	transitGatewayAttachmentResourceType := "TransitGatewayAttachment"
	subnetResourceType := svcsdk.FlowLogsResourceTypeSubnet
	networkInterfaceResourceType := svcsdk.FlowLogsResourceTypeNetworkInterface
	if cr.Spec.ForProvider.VPCID != nil {
		return []*string{cr.Spec.ForProvider.VPCID}, &vpcResourceType
	}

	if cr.Spec.ForProvider.TransitGatewayID != nil {
		return []*string{cr.Spec.ForProvider.TransitGatewayID}, &transitGatewayResourceType
	}

	if cr.Spec.ForProvider.TransitGatewayAttachmentID != nil {
		return []*string{cr.Spec.ForProvider.TransitGatewayAttachmentID}, &transitGatewayAttachmentResourceType
	}

	if cr.Spec.ForProvider.SubnetID != nil {
		return []*string{cr.Spec.ForProvider.SubnetID}, &subnetResourceType
	}

	if cr.Spec.ForProvider.NetworkInterfaceID != nil {
		return []*string{cr.Spec.ForProvider.NetworkInterfaceID}, &networkInterfaceResourceType
	}

	return nil, nil
}

func preObserve(ctx context.Context, cr *svcapitypes.FlowLog, obj *svcsdk.DescribeFlowLogsInput) error {
	externalName := meta.GetExternalName(cr)
	obj.FlowLogIds = []*string{&externalName}
	return nil
}

func filterList(cr *svcapitypes.FlowLog, list *svcsdk.DescribeFlowLogsOutput) *svcsdk.DescribeFlowLogsOutput {
	if len(list.FlowLogs) == 0 {
		return list
	}
	flowLogs := []*svcsdk.FlowLog{}
	for _, f := range list.FlowLogs {
		if pointer.StringValue(f.FlowLogId) == meta.GetExternalName(cr) {
			flowLogs = append(flowLogs, f)
		}
	}
	list.FlowLogs = flowLogs
	return list
}

func postObserve(ctx context.Context, cr *svcapitypes.FlowLog, obj *svcsdk.DescribeFlowLogsOutput, obs managed.ExternalObservation, err error) (managed.ExternalObservation, error) {
	cr.SetConditions(xpv1.Available())
	return obs, err
}

func preCreate(_ context.Context, cr *svcapitypes.FlowLog, obj *svcsdk.CreateFlowLogsInput) error {

	if cr.Spec.ForProvider.S3BucketLogDestination != nil {
		obj.LogDestination = cr.Spec.ForProvider.S3BucketLogDestination
		if cr.Spec.ForProvider.S3BucketSubfolder != nil {
			// If a subfolder is given, we append it to the ARN managed by crossplane
			destination := *obj.LogDestination + "/" + *cr.Spec.ForProvider.S3BucketSubfolder + "/"
			obj.LogDestination = &destination
		}
	}

	if cr.Spec.ForProvider.CloudWatchLogDestination != nil {
		obj.LogDestination = cr.Spec.ForProvider.CloudWatchLogDestination
	}

	if cr.Spec.ForProvider.DeliverLogsPermissionARN != nil {
		obj.DeliverLogsPermissionArn = cr.Spec.ForProvider.DeliverLogsPermissionARN
	}

	if cr.Spec.ForProvider.Tags != nil {

		obj.SetTagSpecifications(generateTagSpecifications(cr))
	}

	obj.ResourceIds, obj.ResourceType = determineResourceIdsAndType(cr)

	return nil
}

func generateTagSpecifications(cr *svcapitypes.FlowLog) []*svcsdk.TagSpecification {
	tagSpecification := &svcsdk.TagSpecification{}
	tagSpecification.SetResourceType(flowLogTagResource)
	tags := []*svcsdk.Tag{}

	for _, cTag := range cr.Spec.ForProvider.Tags {
		tag := &svcsdk.Tag{}

		if cTag.Key != nil {
			tag.SetKey(*cTag.Key)
		}
		if cTag.Value != nil {
			tag.SetValue(*cTag.Value)
		}
		tags = append(tags, tag)
	}

	tagSpecification.SetTags(tags)
	tagSpecifications := []*svcsdk.TagSpecification{tagSpecification}
	return tagSpecifications
}

func postCreate(ctx context.Context, cr *svcapitypes.FlowLog, obj *svcsdk.CreateFlowLogsOutput, cre managed.ExternalCreation, err error) (managed.ExternalCreation, error) {
	if len(obj.FlowLogIds) > 0 {
		meta.SetExternalName(cr, pointer.StringValue(obj.FlowLogIds[0]))
	}
	return cre, nil
}

func (u *updater) isUpToDate(ctx context.Context, cr *svcapitypes.FlowLog, obj *svcsdk.DescribeFlowLogsOutput) (bool, string, error) {

	input := GenerateDescribeFlowLogsInput(cr)
	resp, err := u.client.DescribeFlowLogsWithContext(ctx, input)
	if err != nil {
		return false, "", errors.Wrap(err, errDescribe)
	}

	resp = filterList(cr, resp)

	if len(resp.FlowLogs) == 0 {
		return false, "", errors.New(errDescribe)
	}

	tags := resp.FlowLogs[0].Tags

	add, remove := DiffTags(cr.Spec.ForProvider.Tags, tags)

	return len(add) == 0 && len(remove) == 0, "", nil
}

func (u *updater) update(ctx context.Context, mg cpresource.Managed) (managed.ExternalUpdate, error) {
	cr, ok := mg.(*svcapitypes.FlowLog)
	if !ok {
		return managed.ExternalUpdate{}, errors.New(errUnexpectedObject)
	}
	input := GenerateDescribeFlowLogsInput(cr)
	resp, err := u.client.DescribeFlowLogs(input)
	if err != nil {
		return managed.ExternalUpdate{}, errors.Wrap(err, errDescribe)
	}

	resp = filterList(cr, resp)

	if len(resp.FlowLogs) == 0 {
		return managed.ExternalUpdate{}, errors.New(errDescribe)
	}

	tags := resp.FlowLogs[0].Tags

	add, remove := DiffTags(cr.Spec.ForProvider.Tags, tags)
	err = u.updateTags(ctx, cr, add, remove)
	if err != nil {
		return managed.ExternalUpdate{}, err
	}

	return managed.ExternalUpdate{}, nil
}

// DiffTags returns tags that should be added or removed.
func DiffTags(spec []svcapitypes.Tag, current []*svcsdk.Tag) (addTags []*svcsdk.Tag, remove []*svcsdk.Tag) {
	addMap := make(map[string]string, len(spec))
	for _, t := range spec {
		addMap[pointer.StringValue(t.Key)] = pointer.StringValue(t.Value)
	}
	removeMap := make(map[string]string, len(spec))
	for _, t := range current {
		if addMap[pointer.StringValue(t.Key)] == pointer.StringValue(t.Value) {
			delete(addMap, pointer.StringValue(t.Key))
			continue
		}
		removeMap[pointer.StringValue(t.Key)] = pointer.StringValue(t.Value)
	}
	for k, v := range addMap {
		addTags = append(addTags, &svcsdk.Tag{Key: pointer.ToOrNilIfZeroValue(k), Value: pointer.ToOrNilIfZeroValue(v)})
	}
	for k, v := range removeMap {
		remove = append(remove, &svcsdk.Tag{Key: pointer.ToOrNilIfZeroValue(k), Value: pointer.ToOrNilIfZeroValue(v)})
	}
	return
}

func (u *updater) updateTags(ctx context.Context, cr *svcapitypes.FlowLog, addTags []*svcsdk.Tag, removeTags []*svcsdk.Tag) error {

	if len(removeTags) > 0 {
		inputR := &svcsdk.DeleteTagsInput{
			Resources: pointer.SliceValueToPtr([]string{meta.GetExternalName(cr)}),
			Tags:      removeTags,
		}

		_, err := u.client.DeleteTagsWithContext(ctx, inputR)
		if err != nil {
			return errors.New(errUpdateTags)
		}
	}
	if len(addTags) > 0 {
		inputC := &svcsdk.CreateTagsInput{
			Resources: pointer.SliceValueToPtr([]string{meta.GetExternalName(cr)}),
			Tags:      addTags,
		}

		_, err := u.client.CreateTagsWithContext(ctx, inputC)
		if err != nil {
			return errors.New(errUpdateTags)
		}

	}
	return nil

}

// GenerateDeleteFlowLogsInput returns a deletion input.
func GenerateDeleteFlowLogsInput(cr *svcapitypes.FlowLog) *svcsdk.DeleteFlowLogsInput {
	res := &svcsdk.DeleteFlowLogsInput{}

	externalName := meta.GetExternalName(cr)
	res.SetFlowLogIds([]*string{&externalName})
	return res
}

func (d *deleter) delete(ctx context.Context, mg cpresource.Managed) error {
	cr, ok := mg.(*svcapitypes.FlowLog)
	if !ok {
		return errors.New(errUnexpectedObject)
	}
	cr.Status.SetConditions(xpv1.Deleting())
	if meta.GetExternalName(cr) == "" {
		return nil
	}
	input := GenerateDeleteFlowLogsInput(cr)
	_, err := d.client.DeleteFlowLogsWithContext(ctx, input)
	return errorutils.Wrap(cpresource.Ignore(IsNotFound, err), errDelete)
}
