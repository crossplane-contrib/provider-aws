package provisionedproduct

import (
	"context"

	"github.com/aws/aws-sdk-go/aws/awserr"
	svcsdk "github.com/aws/aws-sdk-go/service/servicecatalog"
	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/pointer"

	svcapitypes "github.com/crossplane-contrib/provider-aws/apis/servicecatalog/v1alpha1"
	aws "github.com/crossplane-contrib/provider-aws/pkg/clients"
	awsclient "github.com/crossplane-contrib/provider-aws/pkg/clients"
)

const (
	msgProvisionedProductStatusSdkTainted        = "provisioned product has status TAINTED"
	msgProvisionedProductStatusSdkUnderChange    = "provisioned product is updating, availability depends on product"
	msgProvisionedProductStatusSdkPlanInProgress = "provisioned product is awaiting plan approval"
	msgProvisionedProductStatusSdkError          = "provisioned product has status ERROR"

	errUpdatePending = "Provisioned product is already under change, not updating"
)

func (c *custom) lateInitialize(spec *svcapitypes.ProvisionedProductParameters, _ *svcsdk.DescribeProvisionedProductOutput) error {
	acceptLanguageEnglish := acceptLanguageEnglish
	spec.AcceptLanguage = awsclient.LateInitializeStringPtr(spec.AcceptLanguage, &acceptLanguageEnglish)
	return nil
}

func (c *custom) isUpToDate(_ context.Context, ds *svcapitypes.ProvisionedProduct, resp *svcsdk.DescribeProvisionedProductOutput) (bool, string, error) {
	// If the product is undergoing change, we want to assume that it is not up-to-date. This will force this resource
	// to be queued for an update (which will be skipped due to UNDER_CHANGE), and once that update fails, we will
	// recheck the status again. This will allow us to quickly transition from UNDER_CHANGE to AVAILABLE without having
	// to wait for the entire polling interval to pass before re-checking the status.
	if pointer.StringDeref(ds.Status.AtProvider.Status, "") == string(svcapitypes.ProvisionedProductStatus_SDK_UNDER_CHANGE) {
		return true, "", nil
	}

	getPPOutputInput := &svcsdk.GetProvisionedProductOutputsInput{ProvisionedProductId: resp.ProvisionedProductDetail.Id}
	getPPOutput, err := c.client.GetProvisionedProductOutputs(getPPOutputInput)
	if err != nil {
		// We want to specifically handle this exception, since it will occur when something
		// is wrong with the provisioned product (error on creation, tainted, etc)
		// We will be able to handle those specific cases in postObserve
		var aerr awserr.Error
		if ok := errors.As(err, &aerr); ok {
			if aerr.Code() == svcsdk.ErrCodeInvalidParametersException {
				if aerr.Message() == "Last Successful Provisioning Record doesn't exist." {
					return false, "", nil
				}
			}
		}

		return false, "", errors.Wrap(err, "could not get provisioned product outputs")
	}
	c.cache.getProvisionedProductOutputs = getPPOutput.Outputs
	cfStackParameters, err := c.client.GetCloudformationStackParameters(getPPOutput.Outputs)
	if err != nil {
		return false, "", errors.Wrap(err, "could not get cloudformation stack parameters")
	}

	productOrArtifactAreChanged, err := c.productOrArtifactAreChanged(&ds.Spec.ForProvider, resp.ProvisionedProductDetail)
	if err != nil {
		return false, "", errors.Wrap(err, "could not discover if product or artifact ids have changed")
	}

	if productOrArtifactAreChanged || provisioningParamsAreChanged(cfStackParameters, ds.Spec.ForProvider.ProvisioningParameters) {
		return false, "", nil
	}
	return true, "", nil
}

func (c *custom) postObserve(_ context.Context, cr *svcapitypes.ProvisionedProduct, resp *svcsdk.DescribeProvisionedProductOutput, obs managed.ExternalObservation, err error) (managed.ExternalObservation, error) { // nolint:gocyclo
	if err != nil {
		return managed.ExternalObservation{}, err
	}

	describeRecordInput := svcsdk.DescribeRecordInput{Id: resp.ProvisionedProductDetail.LastRecordId}
	describeRecordOutput, err := c.client.DescribeRecord(&describeRecordInput)
	if err != nil {
		return managed.ExternalObservation{}, errors.Wrap(err, "could not describe record")
	}

	recordType := ""
	if describeRecordOutput != nil && describeRecordOutput.RecordDetail != nil {
		recordType = pointer.StringDeref(describeRecordOutput.RecordDetail.RecordType, "UPDATE_PROVISIONED_PRODUCT")
	}

	ppStatus := aws.StringValue(resp.ProvisionedProductDetail.Status)
	switch {
	case ppStatus == string(svcapitypes.ProvisionedProductStatus_SDK_AVAILABLE):
		cr.SetConditions(xpv1.Available())
	case ppStatus == string(svcapitypes.ProvisionedProductStatus_SDK_UNDER_CHANGE) && recordType == "PROVISION_PRODUCT":
		cr.SetConditions(xpv1.Creating())
	case ppStatus == string(svcapitypes.ProvisionedProductStatus_SDK_UNDER_CHANGE) && recordType == "UPDATE_PROVISIONED_PRODUCT":
		cr.SetConditions(xpv1.Unavailable().WithMessage(msgProvisionedProductStatusSdkUnderChange))
	case ppStatus == string(svcapitypes.ProvisionedProductStatus_SDK_UNDER_CHANGE) && recordType == "TERMINATE_PROVISIONED_PRODUCT":
		cr.SetConditions(xpv1.Deleting())
	case ppStatus == string(svcapitypes.ProvisionedProductStatus_SDK_PLAN_IN_PROGRESS):
		cr.SetConditions(xpv1.Available().WithMessage(msgProvisionedProductStatusSdkPlanInProgress))
	case ppStatus == string(svcapitypes.ProvisionedProductStatus_SDK_ERROR):
		cr.SetConditions(xpv1.Unavailable().WithMessage(msgProvisionedProductStatusSdkError))
	case ppStatus == string(svcapitypes.ProvisionedProductStatus_SDK_TAINTED):
		cr.SetConditions(xpv1.Unavailable().WithMessage(msgProvisionedProductStatusSdkTainted))
	}

	var outputs = make(map[string]*svcapitypes.RecordOutput)
	for _, v := range c.cache.getProvisionedProductOutputs {
		outputs[*v.OutputKey] = &svcapitypes.RecordOutput{
			Description: v.Description,
			OutputValue: v.OutputValue}
	}

	cr.Status.AtProvider.Outputs = outputs
	cr.Status.AtProvider.ARN = resp.ProvisionedProductDetail.Arn
	cr.Status.AtProvider.CreatedTime = &metav1.Time{Time: *resp.ProvisionedProductDetail.CreatedTime}
	cr.Status.AtProvider.LastProvisioningRecordID = resp.ProvisionedProductDetail.LastProvisioningRecordId
	cr.Status.AtProvider.LaunchRoleARN = resp.ProvisionedProductDetail.LaunchRoleArn
	cr.Status.AtProvider.Status = resp.ProvisionedProductDetail.Status
	cr.Status.AtProvider.StatusMessage = resp.ProvisionedProductDetail.StatusMessage
	cr.Status.AtProvider.ProvisionedProductType = resp.ProvisionedProductDetail.Type
	cr.Status.AtProvider.RecordType = describeRecordOutput.RecordDetail.RecordType
	cr.Status.AtProvider.LastPathID = describeRecordOutput.RecordDetail.PathId
	cr.Status.AtProvider.LastProductID = describeRecordOutput.RecordDetail.ProductId
	cr.Status.AtProvider.LastProvisioningArtifactID = describeRecordOutput.RecordDetail.ProvisioningArtifactId

	return obs, nil
}

func (c *custom) preCreate(_ context.Context, cr *svcapitypes.ProvisionedProduct, obj *svcsdk.ProvisionProductInput) error {
	obj.ProvisionToken = aws.String(genIdempotencyToken())

	// We want to specifically set this to match the forProvider name, as that
	// is what we use to track the provisioned product
	if cr.Spec.ForProvider.Name != nil {
		meta.SetExternalName(cr, *cr.Spec.ForProvider.Name)
	}

	return nil
}

func (c *custom) postCreate(_ context.Context, cr *svcapitypes.ProvisionedProduct, obj *svcsdk.ProvisionProductOutput, cre managed.ExternalCreation, err error) (managed.ExternalCreation, error) {
	if err != nil {
		return cre, err
	}

	// We are expected to set the external-name annotation upon creation since
	// it can differ from the metadata.name for the ProvisionedProduct
	if obj.RecordDetail != nil && obj.RecordDetail.ProvisionedProductName != nil {
		meta.SetExternalName(cr, *obj.RecordDetail.ProvisionedProductName)
	}

	return cre, nil
}

// Update replaces the ExternalClient.Update function so we can perform some custom actions before delegating back to it
func (c *custom) Update(ctx context.Context, mg resource.Managed) (managed.ExternalUpdate, error) {
	cr, ok := mg.(*svcapitypes.ProvisionedProduct)
	if !ok {
		return managed.ExternalUpdate{}, errors.New(errUnexpectedObject)
	}

	// We want to check if we are in a state where we can actually update it
	ppStatus := pointer.StringDeref(cr.Status.AtProvider.Status, "")
	if ppStatus == "" || ppStatus == string(svcapitypes.ProvisionedProductStatus_SDK_UNDER_CHANGE) {
		// If we do not yet have a status, or it is still under change, it has
		// not yet been processed by the Service Catalog API, and requesting an
		// update will cause a 400 error code to be returned prematurely.
		return managed.ExternalUpdate{}, errors.New(errUpdatePending)
	}

	return c.external.Update(ctx, mg)
}

func (c *custom) preUpdate(_ context.Context, cr *svcapitypes.ProvisionedProduct, input *svcsdk.UpdateProvisionedProductInput) error {
	input.UpdateToken = aws.String(genIdempotencyToken())
	input.ProvisionedProductName = aws.String(meta.GetExternalName(cr))
	return nil
}

func (c *custom) preDelete(_ context.Context, cr *svcapitypes.ProvisionedProduct, obj *svcsdk.TerminateProvisionedProductInput) (bool, error) {
	if pointer.StringDeref(cr.Status.AtProvider.Status, "") == string(svcapitypes.ProvisionedProductStatus_SDK_UNDER_CHANGE) {
		return true, nil
	}
	obj.TerminateToken = aws.String(genIdempotencyToken())
	obj.ProvisionedProductName = aws.String(meta.GetExternalName(cr))
	return false, nil
}
