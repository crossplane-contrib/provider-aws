/*
Copyright 2023 The Crossplane Authors.

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

package provisionedproduct

import (
	"context"
	"fmt"
	"strings"

	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	cfsdkv2 "github.com/aws/aws-sdk-go-v2/service/cloudformation"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	svcsdk "github.com/aws/aws-sdk-go/service/servicecatalog"
	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/connection"
	"github.com/crossplane/crossplane-runtime/pkg/controller"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
	ctrl "sigs.k8s.io/controller-runtime"

	svcapitypes "github.com/crossplane-contrib/provider-aws/apis/servicecatalog/v1alpha1"
	"github.com/crossplane-contrib/provider-aws/apis/v1alpha1"
	clientset "github.com/crossplane-contrib/provider-aws/pkg/clients/servicecatalog"
	"github.com/crossplane-contrib/provider-aws/pkg/features"
	"github.com/crossplane-contrib/provider-aws/pkg/utils/pointer"
	custommanaged "github.com/crossplane-contrib/provider-aws/pkg/utils/reconciler/managed"
)

const (
	acceptLanguageEnglish = "en"

	msgProvisionedProductStatusSdkTainted        = "provisioned product has status TAINTED"
	msgProvisionedProductStatusSdkUnderChange    = "provisioned product is updating, availability depends on product"
	msgProvisionedProductStatusSdkPlanInProgress = "provisioned product is awaiting plan approval"
	msgProvisionedProductStatusSdkError          = "provisioned product has status ERROR"

	errCouldNotGetProvisionedProductOutputs = "could not get provisioned product outputs"
	errCouldNotGetCFParameters              = "could not get cloudformation stack parameters"
	errCouldNotDescribeRecord               = "could not describe record"
	errCouldNotLookupProduct                = "could not lookup product"
	errAwsAPICodeInvalidParametersException = "Last Successful Provisioning Record doesn't exist."
)

type custom struct {
	*external
	client clientset.Client
	cache  cache
}

type cache struct {
	getProvisionedProductOutputs []*svcsdk.RecordOutput
	lastProvisioningParameters   []*svcapitypes.ProvisioningParameter
}

// SetupProvisionedProduct adds a controller that reconciles a ProvisionedProduct
func SetupProvisionedProduct(mgr ctrl.Manager, o controller.Options) error {
	name := managed.ControllerName(svcapitypes.ProvisionedProductKind)
	awsCfg, err := awsconfig.LoadDefaultConfig(context.TODO())
	if err != nil {
		return err
	}
	cfClient := cfsdkv2.NewFromConfig(awsCfg)
	opts := []option{prepareSetupExternal(cfClient)}
	cps := []managed.ConnectionPublisher{managed.NewAPISecretPublisher(mgr.GetClient(), mgr.GetScheme())}
	if o.Features.Enabled(features.EnableAlphaExternalSecretStores) {
		cps = append(cps, connection.NewDetailsManager(mgr.GetClient(), v1alpha1.StoreConfigGroupVersionKind))
	}

	reconcilerOpts := []managed.ReconcilerOption{
		managed.WithCriticalAnnotationUpdater(custommanaged.NewRetryingCriticalAnnotationUpdater(mgr.GetClient())),
		managed.WithExternalConnecter(&connector{kube: mgr.GetClient(), opts: opts}),
		managed.WithInitializers(managed.NewNameAsExternalName(mgr.GetClient())),
		managed.WithPollInterval(o.PollInterval),
		managed.WithLogger(o.Logger.WithValues("controller", name)),
		managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name))),
		managed.WithConnectionPublishers(cps...),
	}

	if o.Features.Enabled(features.EnableAlphaManagementPolicies) {
		reconcilerOpts = append(reconcilerOpts, managed.WithManagementPolicies())
	}

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(o.ForControllerRuntime()).
		WithEventFilter(resource.DesiredStateChanged()).
		For(&svcapitypes.ProvisionedProduct{}).
		Complete(managed.NewReconciler(mgr,
			resource.ManagedKind(svcapitypes.ProvisionedProductGroupVersionKind),
			reconcilerOpts...))
}

func prepareSetupExternal(cfClient *cfsdkv2.Client) func(*external) {
	return func(e *external) {
		c := &custom{client: &clientset.CustomServiceCatalogClient{CfClient: cfClient, Client: e.client}}
		e.preCreate = preCreate
		e.preUpdate = c.preUpdate
		e.lateInitialize = c.lateInitialize
		e.isUpToDate = c.isUpToDate
		e.preObserve = c.preObserve
		e.postObserve = c.postObserve
		e.preDelete = preDelete
	}
}

func (c *custom) lateInitialize(spec *svcapitypes.ProvisionedProductParameters, _ *svcsdk.DescribeProvisionedProductOutput) error {
	acceptLanguageEnglish := acceptLanguageEnglish
	spec.AcceptLanguage = pointer.LateInitialize(spec.AcceptLanguage, &acceptLanguageEnglish)
	return nil
}

func preCreate(_ context.Context, ds *svcapitypes.ProvisionedProduct, input *svcsdk.ProvisionProductInput) error {
	input.ProvisionToken = aws.String(genIdempotencyToken())
	input.ProvisionedProductName = aws.String(meta.GetExternalName(ds))
	return nil
}

func (c *custom) preUpdate(_ context.Context, ds *svcapitypes.ProvisionedProduct, input *svcsdk.UpdateProvisionedProductInput) error {
	input.UpdateToken = aws.String(genIdempotencyToken())
	input.ProvisionedProductName = aws.String(meta.GetExternalName(ds))
	return nil
}

func (c *custom) preObserve(_ context.Context, ds *svcapitypes.ProvisionedProduct, input *svcsdk.DescribeProvisionedProductInput) error {
	input.Name = aws.String(meta.GetExternalName(ds))
	c.cache.lastProvisioningParameters = ds.Status.AtProvider.DeepCopy().LastProvisioningParameters
	return nil
}

func (c *custom) isUpToDate(_ context.Context, ds *svcapitypes.ProvisionedProduct, resp *svcsdk.DescribeProvisionedProductOutput) (bool, string, error) { //nolint:gocyclo
	// If the product is undergoing change, we want to assume that it is not up-to-date. This will force this resource
	// to be queued for an update (which will be skipped due to UNDER_CHANGE), and once that update fails, we will
	// recheck the status again. This will allow us to quickly transition from UNDER_CHANGE to AVAILABLE without having
	// to wait for the entire polling interval to pass before re-checking the status.
	if ptr.Deref(resp.ProvisionedProductDetail.Status, "") == string(svcapitypes.ProvisionedProductStatus_SDK_UNDER_CHANGE) ||
		ptr.Deref(resp.ProvisionedProductDetail.Status, "") == string(svcapitypes.ProvisionedProductStatus_SDK_ERROR) {
		return true, "", nil
	}

	getPPOutputInput := &svcsdk.GetProvisionedProductOutputsInput{ProvisionedProductId: resp.ProvisionedProductDetail.Id}
	getPPOutput, err := c.client.GetProvisionedProductOutputs(getPPOutputInput)
	if err != nil {
		// We want to specifically handle this exception, since it will occur when something
		// is wrong with the provisioned product (error on creation, tainted, etc.)
		// We will be able to handle those specific cases in postObserve
		var aerr awserr.Error
		if ok := errors.As(err, &aerr); ok && aerr.Code() == svcsdk.ErrCodeInvalidParametersException && aerr.Message() == errAwsAPICodeInvalidParametersException {
			return false, "", nil
		}
		return false, "", errors.Wrap(err, errCouldNotGetProvisionedProductOutputs)
	}
	c.cache.getProvisionedProductOutputs = getPPOutput.Outputs

	productOrArtifactIsChanged, artifactDiff, err := c.productOrArtifactIsChanged(ds, resp.ProvisionedProductDetail)
	if err != nil {
		return false, "", errors.Wrap(err, "could not discover if product or artifact ids have changed")
	}
	provisioningParamsAreChanged, paramsDiff, err := c.provisioningParamsAreChanged(ds)
	if err != nil {
		return false, "", errors.Wrap(err, "could not compare provisioning parameters with previous ones")
	}

	if productOrArtifactIsChanged || provisioningParamsAreChanged {
		return false, fmt.Sprintf("%s; %s", artifactDiff, paramsDiff), nil
	}
	return true, "", nil
}

func (c *custom) postObserve(_ context.Context, ds *svcapitypes.ProvisionedProduct, resp *svcsdk.DescribeProvisionedProductOutput, obs managed.ExternalObservation, err error) (managed.ExternalObservation, error) {
	if err != nil {
		return managed.ExternalObservation{}, err
	}

	describeRecordInput := svcsdk.DescribeRecordInput{Id: resp.ProvisionedProductDetail.LastRecordId}
	describeRecordOutput, err := c.client.DescribeRecord(&describeRecordInput)
	if err != nil {
		return managed.ExternalObservation{}, errors.Wrap(err, errCouldNotDescribeRecord)
	}

	setConditions(describeRecordOutput, resp, ds)

	var outputs = make(map[string]*svcapitypes.RecordOutput)
	for _, v := range c.cache.getProvisionedProductOutputs {
		outputs[*v.OutputKey] = &svcapitypes.RecordOutput{
			Description: v.Description,
			OutputValue: v.OutputValue}
	}

	ds.Status.AtProvider.Outputs = outputs
	ds.Status.AtProvider.ARN = resp.ProvisionedProductDetail.Arn
	ds.Status.AtProvider.CreatedTime = &metav1.Time{Time: *resp.ProvisionedProductDetail.CreatedTime}
	ds.Status.AtProvider.LastProvisioningRecordID = resp.ProvisionedProductDetail.LastProvisioningRecordId
	ds.Status.AtProvider.LaunchRoleARN = resp.ProvisionedProductDetail.LaunchRoleArn
	ds.Status.AtProvider.Status = resp.ProvisionedProductDetail.Status
	ds.Status.AtProvider.StatusMessage = resp.ProvisionedProductDetail.StatusMessage
	ds.Status.AtProvider.ProvisionedProductType = resp.ProvisionedProductDetail.Type
	ds.Status.AtProvider.RecordType = describeRecordOutput.RecordDetail.RecordType
	ds.Status.AtProvider.LastProductID = describeRecordOutput.RecordDetail.ProductId
	ds.Status.AtProvider.LastProvisioningArtifactID = describeRecordOutput.RecordDetail.ProvisioningArtifactId
	ds.Status.AtProvider.LastProvisioningParameters = ds.Spec.ForProvider.ProvisioningParameters

	return obs, nil
}

func preDelete(_ context.Context, ds *svcapitypes.ProvisionedProduct, input *svcsdk.TerminateProvisionedProductInput) (bool, error) {
	if ptr.Deref(ds.Status.AtProvider.Status, "") == string(svcapitypes.ProvisionedProductStatus_SDK_UNDER_CHANGE) {
		return true, nil
	}
	input.TerminateToken = aws.String(genIdempotencyToken())
	input.ProvisionedProductName = aws.String(meta.GetExternalName(ds))
	return false, nil
}

func setConditions(describeRecordOutput *svcsdk.DescribeRecordOutput, resp *svcsdk.DescribeProvisionedProductOutput, ds *svcapitypes.ProvisionedProduct) {
	ppStatus := aws.StringValue(resp.ProvisionedProductDetail.Status)
	switch {
	case ppStatus == string(svcapitypes.ProvisionedProductStatus_SDK_AVAILABLE):
		ds.SetConditions(xpv1.Available())
	case ppStatus == string(svcapitypes.ProvisionedProductStatus_SDK_UNDER_CHANGE):
		recordType := ptr.Deref(describeRecordOutput.RecordDetail.RecordType, "UPDATE_PROVISIONED_PRODUCT")
		switch {
		case recordType == "PROVISION_PRODUCT":
			ds.SetConditions(xpv1.Creating())
		case recordType == "UPDATE_PROVISIONED_PRODUCT":
			ds.SetConditions(xpv1.Unavailable().WithMessage(msgProvisionedProductStatusSdkUnderChange))
		case recordType == "TERMINATE_PROVISIONED_PRODUCT":
			ds.SetConditions(xpv1.Deleting())
		}
	case ppStatus == string(svcapitypes.ProvisionedProductStatus_SDK_PLAN_IN_PROGRESS):
		ds.SetConditions(xpv1.Unavailable().WithMessage(msgProvisionedProductStatusSdkPlanInProgress))
	case ppStatus == string(svcapitypes.ProvisionedProductStatus_SDK_ERROR):
		ds.SetConditions(xpv1.Unavailable().WithMessage(msgProvisionedProductStatusSdkError))
	case ppStatus == string(svcapitypes.ProvisionedProductStatus_SDK_TAINTED):
		ds.SetConditions(xpv1.Unavailable().WithMessage(msgProvisionedProductStatusSdkTainted))
	}
}

func (c *custom) provisioningParamsAreChanged(ds *svcapitypes.ProvisionedProduct) (bool, string, error) { //nolint:gocyclo
	type ProvisioningParameter struct {
		Key   string
		Value string
	}

	// Compare provisioning params from desired state whits params from previous reconciliation loop(if it exists)
	if c.cache.lastProvisioningParameters != nil {
		less := func(a *svcapitypes.ProvisioningParameter, b *svcapitypes.ProvisioningParameter) bool {
			return pointer.StringValue(a.Key) < pointer.StringValue(b.Key)
		}
		if diff := cmp.Diff(c.cache.lastProvisioningParameters, ds.Spec.ForProvider.ProvisioningParameters, cmpopts.SortSlices(less)); diff != "" {
			return true, diff, nil
		}
	}

	cfStackParams, err := c.client.GetCloudformationStackParameters(c.cache.getProvisionedProductOutputs)
	if err != nil {
		return false, "", errors.Wrap(err, errCouldNotGetCFParameters)
	}

	cfStackKeyValue := make(map[string]string)
	for _, v := range cfStackParams {
		if v.ParameterKey != nil {
			cfStackKeyValue[*v.ParameterKey] = ptr.Deref(v.ParameterValue, "")
		}
	}
	var diff []ProvisioningParameter
	for _, v := range ds.Spec.ForProvider.ProvisioningParameters {
		// In this comparison the controller ignores spaces from the left and right of the parameter value from
		// the desired state. Because on cloudformation side spaces are also trimmed
		if v.Key != nil {
			if cfv, ok := cfStackKeyValue[*v.Key]; !ok || strings.TrimSpace(ptr.Deref(v.Value, "")) != cfv {
				diff = append(diff, ProvisioningParameter{Key: *v.Key, Value: ptr.Deref(v.Value, "")})
			}
		}
	}
	if diff != nil {
		return true, fmt.Sprintf("ProvisioningParameters are changed: %v", diff), nil
	}

	return false, "", nil
}

func (c *custom) productOrArtifactIsChanged(ds *svcapitypes.ProvisionedProduct,
	resp *svcsdk.ProvisionedProductDetail) (bool, string, error) {
	// ProvisioningArtifactID and ProvisioningArtifactName are mutual exclusive params, the same about ProductID and ProductName
	// But if describe a provisioned product aws api will return only IDs, so it's impossible to compare names with ids
	// Two conditional statements below work only of IDs
	var diffList []string
	if ds.Spec.ForProvider.ProvisioningArtifactID != nil {
		diff := cmp.Diff(*ds.Spec.ForProvider.ProvisioningArtifactID, *resp.ProvisioningArtifactId)
		diffList = append(diffList, diff)
	}
	if ds.Spec.ForProvider.ProductID != nil {
		diff := cmp.Diff(*ds.Spec.ForProvider.ProductID, *resp.ProductId)
		diffList = append(diffList, diff)
	}
	// In case if the desired state has ProvisioningArtifactName the controller will run func `getArtifactID`, which produces
	// additional request to aws api to resolve the artifact id(based on ProductId/ProductName)
	// for further comparison with artifact id in the current state
	if ds.Spec.ForProvider.ProvisioningArtifactName != nil {
		desiredArtifactID, err := c.getArtifactID(ds)
		if err != nil {
			return false, "", err
		}
		diff := cmp.Diff(desiredArtifactID, *resp.ProvisioningArtifactId)
		diffList = append(diffList, diff)
	}
	// If desired state includes ProductName the controller will resolve ID via func `getProductID`,
	// which also produce additional api request
	if ds.Spec.ForProvider.ProductName != nil {
		desiredProductID, err := c.getProductID(ds.Spec.ForProvider.ProductName)
		if err != nil {
			return false, "", err
		}
		diff := cmp.Diff(desiredProductID, *resp.ProductId)
		diffList = append(diffList, diff)
	}

	finalDiff := ""
	for _, diff := range diffList {
		if diff != "" {
			finalDiff += diff
		}
	}
	if finalDiff != "" {
		return true, finalDiff, nil
	}

	return false, "", nil
}

func (c *custom) getArtifactID(ds *svcapitypes.ProvisionedProduct) (string, error) {
	input := svcsdk.DescribeProductInput{
		Id:   ds.Spec.ForProvider.ProductID,
		Name: ds.Spec.ForProvider.ProductName,
	}
	// DescribeProvisioningArtifact method fits much better, but it has a bug - it returns nothing if a product is a part of imported portfolio
	output, err := c.client.DescribeProduct(&input)
	if err != nil {
		return "", errors.Wrap(err, errCouldNotLookupProduct)
	}
	for _, artifact := range output.ProvisioningArtifacts {
		if ptr.Deref(ds.Spec.ForProvider.ProvisioningArtifactName, "") == *artifact.Name ||
			ptr.Deref(ds.Spec.ForProvider.ProvisioningArtifactID, "") == *artifact.Id {
			return *artifact.Id, nil
		}
	}
	return "", errors.Wrap(errors.New("artifact not found"), errCouldNotLookupProduct)
}

func (c *custom) getProductID(productName *string) (string, error) {
	input := svcsdk.DescribeProductInput{Name: productName}
	// DescribeProvisioningArtifact method fits much better, but it has a bug - it returns nothing if a product is a part of imported portfolio
	output, err := c.client.DescribeProduct(&input)
	if err != nil {
		return "", errors.Wrap(err, errCouldNotLookupProduct)
	}
	return ptr.Deref(output.ProductViewSummary.ProductId, ""), nil
}

func genIdempotencyToken() string {
	return fmt.Sprintf("provider-aws-%s", uuid.New())
}
