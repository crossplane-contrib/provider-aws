/*
Copyright 2020 The Crossplane Authors.

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

package table

import (
	"context"
	"encoding/json"
	"sort"
	"strings"

	svcsdk "github.com/aws/aws-sdk-go/service/dynamodb"
	svcsdkapi "github.com/aws/aws-sdk-go/service/dynamodb/dynamodbiface"
	"github.com/aws/aws-sdk-go/service/kms"
	"github.com/aws/aws-sdk-go/service/kms/kmsiface"
	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/connection"
	"github.com/crossplane/crossplane-runtime/pkg/controller"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	cpresource "github.com/crossplane/crossplane-runtime/pkg/resource"
	"github.com/pkg/errors"
	"k8s.io/utils/ptr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	svcapitypes "github.com/crossplane-contrib/provider-aws/apis/dynamodb/v1alpha1"
	"github.com/crossplane-contrib/provider-aws/apis/v1alpha1"
	"github.com/crossplane-contrib/provider-aws/pkg/features"
	connectaws "github.com/crossplane-contrib/provider-aws/pkg/utils/connect/aws"
	errorutils "github.com/crossplane-contrib/provider-aws/pkg/utils/errors"
	"github.com/crossplane-contrib/provider-aws/pkg/utils/jsonpatch"
	"github.com/crossplane-contrib/provider-aws/pkg/utils/pointer"
	custommanaged "github.com/crossplane-contrib/provider-aws/pkg/utils/reconciler/managed"
)

const (
	errResolveKMSMasterKeyArn = "cannot resolve kms master key ARN"
)

// SetupTable adds a controller that reconciles Table.
func SetupTable(mgr ctrl.Manager, o controller.Options) error {
	name := managed.ControllerName(svcapitypes.TableGroupKind)

	cps := []managed.ConnectionPublisher{managed.NewAPISecretPublisher(mgr.GetClient(), mgr.GetScheme())}
	if o.Features.Enabled(features.EnableAlphaExternalSecretStores) {
		cps = append(cps, connection.NewDetailsManager(mgr.GetClient(), v1alpha1.StoreConfigGroupVersionKind))
	}

	reconcilerOpts := []managed.ReconcilerOption{
		managed.WithCriticalAnnotationUpdater(custommanaged.NewRetryingCriticalAnnotationUpdater(mgr.GetClient())),
		managed.WithExternalConnecter(&customConnector{kube: mgr.GetClient()}),
		managed.WithInitializers(managed.NewNameAsExternalName(mgr.GetClient())),
		managed.WithPollInterval(o.PollInterval),
		managed.WithLogger(o.Logger.WithValues("controller", name)),
		managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name))),
		managed.WithConnectionPublishers(cps...),
	}

	if o.Features.Enabled(features.EnableAlphaManagementPolicies) {
		reconcilerOpts = append(reconcilerOpts, managed.WithManagementPolicies())
	}

	r := managed.NewReconciler(mgr,
		resource.ManagedKind(svcapitypes.TableGroupVersionKind),
		reconcilerOpts...)

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(o.ForControllerRuntime()).
		WithEventFilter(resource.DesiredStateChanged()).
		For(&svcapitypes.Table{}).
		Complete(r)
}

// customConnector is needed because the generated connector does not allow
// the creation of the kms client.
type customConnector struct {
	kube client.Client
}

func (c *customConnector) Connect(ctx context.Context, mg cpresource.Managed) (managed.ExternalClient, error) {
	cr, ok := mg.(*svcapitypes.Table)
	if !ok {
		return nil, errors.New(errUnexpectedObject)
	}
	sess, err := connectaws.GetConfigV1(ctx, c.kube, mg, cr.Spec.ForProvider.Region)
	if err != nil {
		return nil, errors.Wrap(err, errCreateSession)
	}

	// Custom options are created here instead of Setup because the config is
	// needed in order to create the kms client.
	opts := []option{
		func(e *external) {
			e.preObserve = preObserve
			e.postObserve = postObserve
			e.preCreate = preCreate
			e.preDelete = preDelete
			e.postDelete = postDelete
			e.lateInitialize = lateInitialize
			u := &updateClient{
				client:    e.client,
				clientkms: kms.New(sess),
			}
			e.preUpdate = u.preUpdate
			e.isUpToDate = u.isUpToDate
			e.postUpdate = u.postUpdate
		},
	}
	return newExternal(c.kube, svcsdk.New(sess), opts), nil
}

func (e *updateClient) postUpdate(_ context.Context, cr *svcapitypes.Table, obj *svcsdk.UpdateTableOutput, _ managed.ExternalUpdate, _ error) (managed.ExternalUpdate, error) {
	cbresult, err := e.client.DescribeContinuousBackups(&svcsdk.DescribeContinuousBackupsInput{
		TableName: pointer.ToOrNilIfZeroValue(meta.GetExternalName(cr)),
	})
	if err != nil {
		return managed.ExternalUpdate{}, err
	}
	pitrStatus := cbresult.ContinuousBackupsDescription.PointInTimeRecoveryDescription.PointInTimeRecoveryStatus
	pitrStatusBool := pitrStatusToBool(pitrStatus)

	if !isPitrUpToDate(cr, pitrStatusBool) {
		pitrSpecEnabled := ptr.Deref(cr.Spec.ForProvider.PointInTimeRecoveryEnabled, false)

		pitrInput := &svcsdk.UpdateContinuousBackupsInput{
			TableName: pointer.ToOrNilIfZeroValue(meta.GetExternalName(cr)),
			PointInTimeRecoverySpecification: (&svcsdk.PointInTimeRecoverySpecification{
				PointInTimeRecoveryEnabled: &pitrSpecEnabled,
			}),
		}

		_, err := e.client.UpdateContinuousBackups(pitrInput)
		if err != nil {
			return managed.ExternalUpdate{}, err
		}
	}

	return managed.ExternalUpdate{}, nil
}

func preObserve(_ context.Context, cr *svcapitypes.Table, obj *svcsdk.DescribeTableInput) error {
	obj.TableName = pointer.ToOrNilIfZeroValue(meta.GetExternalName(cr))

	return nil
}
func preCreate(_ context.Context, cr *svcapitypes.Table, obj *svcsdk.CreateTableInput) error {
	obj.TableName = pointer.ToOrNilIfZeroValue(meta.GetExternalName(cr))
	return nil
}
func preDelete(_ context.Context, cr *svcapitypes.Table, obj *svcsdk.DeleteTableInput) (bool, error) {
	obj.TableName = pointer.ToOrNilIfZeroValue(meta.GetExternalName(cr))
	return false, nil
}

func postDelete(_ context.Context, _ *svcapitypes.Table, _ *svcsdk.DeleteTableOutput, err error) error {
	if err == nil {
		return nil
	}
	// The DynamoDB API returns this error when you try to delete a table
	// that is already being deleted. Unfortunately we're passed a v1 SDK
	// error that has been munged by aws.Wrap, so we can't use errors.As to
	// identify it and must fall back to string matching.
	if strings.Contains(err.Error(), "ResourceInUseException") {
		return nil
	}
	return err
}

func postObserve(_ context.Context, cr *svcapitypes.Table, resp *svcsdk.DescribeTableOutput, obs managed.ExternalObservation, err error) (managed.ExternalObservation, error) {
	if err != nil {
		return managed.ExternalObservation{}, err
	}
	switch pointer.StringValue(resp.Table.TableStatus) {
	case string(svcapitypes.TableStatus_SDK_CREATING):
		cr.SetConditions(xpv1.Creating())
	case string(svcapitypes.TableStatus_SDK_DELETING):
		cr.SetConditions(xpv1.Deleting())
	case string(svcapitypes.TableStatus_SDK_ACTIVE):
		cr.SetConditions(xpv1.Available())
	case string(svcapitypes.TableStatus_SDK_ARCHIVED), string(svcapitypes.TableStatus_SDK_INACCESSIBLE_ENCRYPTION_CREDENTIALS), string(svcapitypes.TableStatus_SDK_ARCHIVING):
		cr.SetConditions(xpv1.Unavailable())
	}

	obs.ConnectionDetails = managed.ConnectionDetails{
		"tableName":         []byte(meta.GetExternalName(cr)),
		"tableArn":          []byte(pointer.StringValue(resp.Table.TableArn)),
		"latestStreamArn":   []byte(pointer.StringValue(resp.Table.LatestStreamArn)),
		"latestStreamLabel": []byte(pointer.StringValue(resp.Table.LatestStreamLabel)),
	}

	return obs, nil
}

func lateInitialize(in *svcapitypes.TableParameters, t *svcsdk.DescribeTableOutput) error { //nolint:gocyclo
	if t == nil {
		return nil
	}

	if len(in.AttributeDefinitions) == 0 && len(t.Table.AttributeDefinitions) != 0 {
		in.AttributeDefinitions = buildAttributeDefinitions(t.Table.AttributeDefinitions)
	}

	if in.GlobalSecondaryIndexes == nil && len(t.Table.GlobalSecondaryIndexes) != 0 {
		in.GlobalSecondaryIndexes = buildGlobalIndexes(t.Table.GlobalSecondaryIndexes)
	}

	if in.LocalSecondaryIndexes == nil && len(t.Table.LocalSecondaryIndexes) != 0 {
		in.LocalSecondaryIndexes = buildLocalIndexes(t.Table.LocalSecondaryIndexes)
	}

	if in.KeySchema == nil && len(t.Table.KeySchema) != 0 {
		in.KeySchema = buildAlphaKeyElements(t.Table.KeySchema)
	}
	if in.BillingMode == nil {
		// NOTE(negz): As far as I can tell DescribeTableOutput only
		// includes a BillingModeSummary when the billing mode is set to
		// PAY_PER_REQUEST. PROVISIONED seems to be he implied default.
		// Late initializing a value that the API only implies is
		// perhaps a bit of a violation, but it avoids a false positive
		// in our IsUpToDate logic which would otherwise detect a diff
		// between our desired state (PROVISIONED) and the actual state
		// (unspecified).
		in.BillingMode = pointer.ToOrNilIfZeroValue(svcsdk.BillingModeProvisioned)
		if t.Table.BillingModeSummary != nil {
			in.BillingMode = t.Table.BillingModeSummary.BillingMode
		}
	}
	if in.ProvisionedThroughput == nil && t.Table.ProvisionedThroughput != nil {
		in.ProvisionedThroughput = &svcapitypes.ProvisionedThroughput{
			ReadCapacityUnits:  t.Table.ProvisionedThroughput.ReadCapacityUnits,
			WriteCapacityUnits: t.Table.ProvisionedThroughput.WriteCapacityUnits,
		}
	}
	if t.Table.SSEDescription != nil {
		if in.SSESpecification == nil {
			in.SSESpecification = &svcapitypes.SSESpecification{}
		}
		if in.SSESpecification.Enabled == nil && t.Table.SSEDescription.Status != nil {
			in.SSESpecification.Enabled = pointer.ToOrNilIfZeroValue(*t.Table.SSEDescription.Status == string(svcapitypes.SSEStatus_ENABLED))
		}
		if in.SSESpecification.KMSMasterKeyID == nil && t.Table.SSEDescription.KMSMasterKeyArn != nil {
			in.SSESpecification.KMSMasterKeyID = t.Table.SSEDescription.KMSMasterKeyArn
		}
		if in.SSESpecification.SSEType == nil && t.Table.SSEDescription.SSEType != nil {
			in.SSESpecification.SSEType = t.Table.SSEDescription.SSEType
		}
	}
	if in.StreamSpecification == nil {
		// NOTE(negz): We late initialize StreamEnabled to false to
		// avoid IsUpToDate thinking it needs to explicitly make an
		// update to set StreamEnabled to false. DescribeTableOutput
		// omits StreamSpecification entirely when it's not enabled.
		in.StreamSpecification = &svcapitypes.StreamSpecification{StreamEnabled: ptr.To(false)}
		if t.Table.StreamSpecification != nil {
			in.StreamSpecification = &svcapitypes.StreamSpecification{
				StreamEnabled:  t.Table.StreamSpecification.StreamEnabled,
				StreamViewType: t.Table.StreamSpecification.StreamViewType,
			}
		}
	}

	return nil
}

func buildAlphaKeyElements(keys []*svcsdk.KeySchemaElement) []*svcapitypes.KeySchemaElement {
	if len(keys) == 0 {
		return nil
	}
	keyElements := make([]*svcapitypes.KeySchemaElement, len(keys))
	for i, val := range keys {
		keyElements[i] = &svcapitypes.KeySchemaElement{
			AttributeName: val.AttributeName,
			KeyType:       val.KeyType,
		}
	}
	return keyElements
}

func buildAttributeDefinitions(attributes []*svcsdk.AttributeDefinition) []*svcapitypes.AttributeDefinition {
	if len(attributes) == 0 {
		return nil
	}
	attributeDefinitions := make([]*svcapitypes.AttributeDefinition, len(attributes))
	for i, val := range attributes {
		attributeDefinitions[i] = &svcapitypes.AttributeDefinition{
			AttributeName: val.AttributeName,
			AttributeType: val.AttributeType,
		}
	}
	return attributeDefinitions
}

func buildGlobalIndexes(indexes []*svcsdk.GlobalSecondaryIndexDescription) []*svcapitypes.GlobalSecondaryIndex {
	if len(indexes) == 0 {
		return nil
	}
	globalSecondaryIndexes := make([]*svcapitypes.GlobalSecondaryIndex, len(indexes))
	for i, val := range indexes {
		globalSecondaryIndexes[i] = &svcapitypes.GlobalSecondaryIndex{
			IndexName: val.IndexName,
			KeySchema: buildAlphaKeyElements(val.KeySchema),
		}
		if val.Projection != nil {
			globalSecondaryIndexes[i].Projection = &svcapitypes.Projection{
				NonKeyAttributes: val.Projection.NonKeyAttributes,
				ProjectionType:   val.Projection.ProjectionType,
			}
		}
	}
	return globalSecondaryIndexes
}

func buildLocalIndexes(indexes []*svcsdk.LocalSecondaryIndexDescription) []*svcapitypes.LocalSecondaryIndex {
	if len(indexes) == 0 {
		return nil
	}
	localSecondaryIndexes := make([]*svcapitypes.LocalSecondaryIndex, len(indexes))
	for i, val := range indexes {
		localSecondaryIndexes[i] = &svcapitypes.LocalSecondaryIndex{
			IndexName: val.IndexName,
			KeySchema: buildAlphaKeyElements(val.KeySchema),
		}
		if val.Projection != nil {
			localSecondaryIndexes[i].Projection = &svcapitypes.Projection{
				NonKeyAttributes: val.Projection.NonKeyAttributes,
				ProjectionType:   val.Projection.ProjectionType,
			}
		}
	}
	return localSecondaryIndexes
}

// createPatch creates a *svcapitypes.TableParameters that has only the changed
// values between the target *svcapitypes.TableParameters and the current
// *dynamodb.TableDescription
func (e *updateClient) createPatch(ctx context.Context, in *svcsdk.DescribeTableOutput, spec *svcapitypes.TableParameters) (*svcapitypes.TableParameters, error) {
	target := spec.DeepCopy()
	currentParams := &svcapitypes.TableParameters{}
	if err := lateInitialize(currentParams, in); err != nil {
		return nil, err
	}

	if target.SSESpecification != nil && target.SSESpecification.KMSMasterKeyID != nil {
		kmsMasterKeyArn, err := e.getKMsKeyArnFromID(ctx, target.SSESpecification.KMSMasterKeyID)
		if err != nil {
			return nil, errors.Wrap(err, errResolveKMSMasterKeyArn)
		}
		target.SSESpecification.KMSMasterKeyID = kmsMasterKeyArn
	}

	jsonPatch, err := jsonpatch.CreateJSONPatch(currentParams, target)
	if err != nil {
		return nil, err
	}
	patch := &svcapitypes.TableParameters{}
	if err := json.Unmarshal(jsonPatch, patch); err != nil {
		return nil, err
	}
	return patch, nil
}

// getKMsKeyArnFromID from an arbitrary identifier. Might be ARN, ID or alias.
func (e *updateClient) getKMsKeyArnFromID(ctx context.Context, kmsKeyId *string) (*string, error) {
	res, err := e.clientkms.DescribeKeyWithContext(ctx, &kms.DescribeKeyInput{
		KeyId: kmsKeyId,
	})
	if err != nil {
		return nil, err
	}
	return res.KeyMetadata.Arn, nil
}

func (e *updateClient) isCoreResourceUpToDate(ctx context.Context, cr *svcapitypes.Table, resp *svcsdk.DescribeTableOutput) (bool, error) {
	// As continuous backup configuration lives in anoterh api, we extract the part of the isUpToDate logic
	// which is concerned about the actual table-endpoint into a separate function in order to make it testable

	patch, err := e.createPatch(ctx, resp, &cr.Spec.ForProvider)
	if err != nil {
		return false, err
	}

	// TODO(negz): Support updating tags if possible.
	// https://github.com/crossplane-contrib/provider-aws/issues/945

	// At least one of ProvisionedThroughput, BillingMode, UpdateStreamEnabled,
	// GlobalSecondaryIndexUpdates or SSESpecification or ReplicaUpdates is
	// required.
	switch {
	case patch.BillingMode != nil:
		return false, nil
	case patch.ProvisionedThroughput != nil:
		// TODO(negz): DescribeTableOutput appears to report that
		// ProvisionedThroughput is 0 when the billing mode is set to
		// PAY_PER_REQUEST. This means that if the billing mode is
		// changed from PROVISIONED to PAY_PER_REQUEST the provisioned
		// throughput must be set to 0 read and write or Crossplane will
		// think an update is needed and the update will fail because
		// you can't set provisioned throughput when the billing mode is
		// set to PAY_PER_REQUEST.
		return false, nil
	case patch.StreamSpecification != nil:
		return false, nil
	case patch.SSESpecification != nil:
		return false, nil
	case len(diffGlobalSecondaryIndexes(GenerateGlobalSecondaryIndexDescriptions(cr.Spec.ForProvider.GlobalSecondaryIndexes), resp.Table.GlobalSecondaryIndexes)) != 0:
		return false, nil
	}

	return true, nil
}

func isPitrUpToDate(cr *svcapitypes.Table, pitrStatusBool bool) bool {
	// it is not up to date when either point in time recovery is set and the state doesn't match the one on aws
	// or it is unset and it is enabled on aws
	return !((cr.Spec.ForProvider.PointInTimeRecoveryEnabled != nil && *cr.Spec.ForProvider.PointInTimeRecoveryEnabled != pitrStatusBool) ||
		(cr.Spec.ForProvider.PointInTimeRecoveryEnabled == nil && pitrStatusBool))
}

func (e *updateClient) isUpToDate(ctx context.Context, cr *svcapitypes.Table, resp *svcsdk.DescribeTableOutput) (bool, string, error) {
	// A table that's currently creating, deleting, or updating can't be
	// updated, so we temporarily consider it to be up-to-date no matter
	// what.
	switch pointer.StringValue(cr.Status.AtProvider.TableStatus) {
	case string(svcapitypes.TableStatus_SDK_UPDATING), string(svcapitypes.TableStatus_SDK_CREATING), string(svcapitypes.TableStatus_SDK_DELETING):
		return true, "", nil
	}

	// Similarly, a table that's currently updating its SSE status can't be
	// updated, so we temporarily consider it to be up-to-date.
	if cr.Status.AtProvider.SSEDescription != nil && pointer.StringValue(cr.Status.AtProvider.SSEDescription.Status) == string(svcapitypes.SSEStatus_UPDATING) {
		return true, "", nil
	}

	coreUpToDate, err := e.isCoreResourceUpToDate(ctx, cr, resp)
	if err != nil {
		return false, "", err
	}
	if !coreUpToDate {
		return false, "", nil
	}

	// point in time recovery status
	cbresult, err := e.client.DescribeContinuousBackupsWithContext(ctx, &svcsdk.DescribeContinuousBackupsInput{
		TableName: pointer.ToOrNilIfZeroValue(meta.GetExternalName(cr)),
	})
	if err != nil {
		return false, "", err
	}
	pitrStatus := cbresult.ContinuousBackupsDescription.PointInTimeRecoveryDescription.PointInTimeRecoveryStatus
	pitrStatusBool := pitrStatusToBool(pitrStatus)

	if !isPitrUpToDate(cr, pitrStatusBool) {
		return false, "", nil
	}

	return true, "", nil
}

func pitrStatusToBool(pitrStatus *string) bool {
	return ptr.Deref(pitrStatus, "") == string(svcapitypes.PointInTimeRecoveryStatus_ENABLED)
}

type updateClient struct {
	client    svcsdkapi.DynamoDBAPI
	clientkms kmsiface.KMSAPI
}

func (e *updateClient) preUpdate(ctx context.Context, cr *svcapitypes.Table, u *svcsdk.UpdateTableInput) error {
	filtered := &svcsdk.UpdateTableInput{
		TableName:            pointer.ToOrNilIfZeroValue(meta.GetExternalName(cr)),
		AttributeDefinitions: u.AttributeDefinitions,
	}

	// The AWS API requires us to do one kind of update at a time per
	// https://github.com/aws/aws-sdk-go/blob/v1.34.32/service/dynamodb/api.go#L5605
	// This means that we need to return a filtered UpdateTableInput that
	// contains at most one thing that needs updating. In order to be
	// eventually consistent (i.e. to eventually update all the things, one
	// on each reconcile pass) we need to determine the 'next' thing to
	// update on each pass. This means we need to diff actual vs desired
	// state here inside preUpdate. Unfortunately we read the actual state
	// during Observe, but don't typically pass it to update. We could stash
	// the observed state in a cache during postObserve then read it here,
	// but we typically prefer to be as stateless as possible even if it
	// means redundant API calls.
	out, err := e.client.DescribeTableWithContext(ctx, &svcsdk.DescribeTableInput{TableName: pointer.ToOrNilIfZeroValue(meta.GetExternalName(cr))})
	if err != nil {
		return errorutils.Wrap(err, errDescribe)
	}

	p, err := e.createPatch(ctx, out, &cr.Spec.ForProvider)
	if err != nil {
		return err
	}
	gsiUpdates := diffGlobalSecondaryIndexes(GenerateGlobalSecondaryIndexDescriptions(cr.Spec.ForProvider.GlobalSecondaryIndexes), out.Table.GlobalSecondaryIndexes)
	switch {
	case p.BillingMode != nil:
		filtered.BillingMode = u.BillingMode

		// NOTE(negz): You must include provisioned throughput when
		// updating the billing mode to PROVISIONED.
		if pointer.StringValue(u.BillingMode) == string(svcapitypes.BillingMode_PROVISIONED) {
			filtered.ProvisionedThroughput = u.ProvisionedThroughput
		}
	case p.ProvisionedThroughput != nil:
		// NOTE(negz): You may only included provisioned throughput when
		// the billing mode is PROVISIONED.
		filtered.ProvisionedThroughput = u.ProvisionedThroughput
	case p.StreamSpecification != nil:
		// NOTE(muvaf): Unless StreamEnabled is changed, updating stream
		// specification won't work.
		filtered.StreamSpecification = u.StreamSpecification
	case p.SSESpecification != nil:
		// NOTE(negz): Attempting to update the KMSMasterKeyId to its
		// current value returns an error
		filtered.SSESpecification = &svcsdk.SSESpecification{
			Enabled: u.SSESpecification.Enabled,
			SSEType: u.SSESpecification.SSEType,
		}
		if p.SSESpecification.KMSMasterKeyID != nil {
			filtered.SSESpecification.KMSMasterKeyId = u.SSESpecification.KMSMasterKeyId
		}
	case len(gsiUpdates) != 0:
		filtered.SetGlobalSecondaryIndexUpdates(gsiUpdates)
	}

	*u = *filtered
	return nil
}

func diffGlobalSecondaryIndexes(spec []*svcsdk.GlobalSecondaryIndexDescription, obs []*svcsdk.GlobalSecondaryIndexDescription) []*svcsdk.GlobalSecondaryIndexUpdate { //nolint:gocyclo
	// Linter is disabled because there isn't an easy good way to reduce the cyclo
	// complexity here.
	desired := map[string]*svcsdk.GlobalSecondaryIndexDescription{}
	desiredKeys := make([]string, len(spec))
	for i, gsi := range spec {
		desired[pointer.StringValue(gsi.IndexName)] = gsi
		desiredKeys[i] = pointer.StringValue(gsi.IndexName)
	}
	existing := map[string]*svcsdk.GlobalSecondaryIndexDescription{}
	existingKeys := make([]string, len(obs))
	for i, gsi := range obs {
		existing[pointer.StringValue(gsi.IndexName)] = gsi
		existingKeys[i] = pointer.StringValue(gsi.IndexName)
	}
	sort.Strings(desiredKeys)
	sort.Strings(existingKeys)
	// NOTE(muvaf): AWS API supports only a single deletion or creation at once,
	// i.e. we can create or delete only one GlobalSecondaryIndex with a single
	// UpdateTable call. However, we can make multiple updates.
	var updates []*svcsdk.GlobalSecondaryIndexUpdate
	for _, k := range desiredKeys {
		existingGSI, ok := existing[k]
		if !ok {
			gsi := []*svcsdk.GlobalSecondaryIndexUpdate{
				{
					Create: &svcsdk.CreateGlobalSecondaryIndexAction{
						IndexName:  desired[k].IndexName,
						KeySchema:  desired[k].KeySchema,
						Projection: desired[k].Projection,
					},
				},
			}
			if desired[k].ProvisionedThroughput != nil {
				gsi[0].Create.ProvisionedThroughput = &svcsdk.ProvisionedThroughput{
					ReadCapacityUnits:  desired[k].ProvisionedThroughput.ReadCapacityUnits,
					WriteCapacityUnits: desired[k].ProvisionedThroughput.WriteCapacityUnits,
				}
			}
			return gsi
		}
		if desired[k].ProvisionedThroughput != nil {
			if pointer.Int64Value(desired[k].ProvisionedThroughput.WriteCapacityUnits) != pointer.Int64Value(existingGSI.ProvisionedThroughput.WriteCapacityUnits) ||
				pointer.Int64Value(desired[k].ProvisionedThroughput.ReadCapacityUnits) != pointer.Int64Value(existingGSI.ProvisionedThroughput.ReadCapacityUnits) {
				u := &svcsdk.GlobalSecondaryIndexUpdate{
					Update: &svcsdk.UpdateGlobalSecondaryIndexAction{
						IndexName: desired[k].IndexName,
						ProvisionedThroughput: &svcsdk.ProvisionedThroughput{
							ReadCapacityUnits:  desired[k].ProvisionedThroughput.ReadCapacityUnits,
							WriteCapacityUnits: desired[k].ProvisionedThroughput.WriteCapacityUnits,
						},
					},
				}
				updates = append(updates, u)
			}
		}
	}
	if len(updates) != 0 {
		return updates
	}
	// At this point, we handled all creations and updates. The last thing to check
	// is whether there is a removal.
	for _, k := range existingKeys {
		if _, ok := desired[k]; !ok {
			return []*svcsdk.GlobalSecondaryIndexUpdate{
				{
					Delete: &svcsdk.DeleteGlobalSecondaryIndexAction{
						IndexName: existing[k].IndexName,
					},
				},
			}
		}
	}
	return nil
}

// GenerateGlobalSecondaryIndexDescriptions generates an array of GlobalSecondaryIndexDescriptions.
func GenerateGlobalSecondaryIndexDescriptions(p []*svcapitypes.GlobalSecondaryIndex) []*svcsdk.GlobalSecondaryIndexDescription { //nolint:gocyclo
	// Linter is disabled because this is a copy-paste from generated code and
	// very simple.
	result := make([]*svcsdk.GlobalSecondaryIndexDescription, len(p))
	for i, desiredGSI := range p {
		gsi := &svcsdk.GlobalSecondaryIndexDescription{}
		if desiredGSI.IndexName != nil {
			gsi.SetIndexName(*desiredGSI.IndexName)
		}
		if desiredGSI.KeySchema != nil {
			var keySchemaList []*svcsdk.KeySchemaElement
			for _, keySchemaIter := range desiredGSI.KeySchema {
				keySchema := &svcsdk.KeySchemaElement{}
				if keySchemaIter.AttributeName != nil {
					keySchema.SetAttributeName(*keySchemaIter.AttributeName)
				}
				if keySchemaIter.KeyType != nil {
					keySchema.SetKeyType(*keySchemaIter.KeyType)
				}
				keySchemaList = append(keySchemaList, keySchema)
			}
			gsi.SetKeySchema(keySchemaList)
		}
		if desiredGSI.Projection != nil {
			projection := &svcsdk.Projection{}
			if desiredGSI.Projection.NonKeyAttributes != nil {
				var nonKeyAttrList []*string
				for _, nonKeyAttrIter := range desiredGSI.Projection.NonKeyAttributes {
					nonKeyAttr := *nonKeyAttrIter
					nonKeyAttrList = append(nonKeyAttrList, &nonKeyAttr)
				}
				projection.SetNonKeyAttributes(nonKeyAttrList)
			}
			if desiredGSI.Projection.ProjectionType != nil {
				projection.SetProjectionType(*desiredGSI.Projection.ProjectionType)
			}
			gsi.SetProjection(projection)
		}
		if desiredGSI.ProvisionedThroughput != nil {
			provisionedThroughput := &svcsdk.ProvisionedThroughputDescription{}
			if desiredGSI.ProvisionedThroughput.ReadCapacityUnits != nil {
				provisionedThroughput.SetReadCapacityUnits(*desiredGSI.ProvisionedThroughput.ReadCapacityUnits)
			}
			if desiredGSI.ProvisionedThroughput.WriteCapacityUnits != nil {
				provisionedThroughput.SetWriteCapacityUnits(*desiredGSI.ProvisionedThroughput.WriteCapacityUnits)
			}
			gsi.SetProvisionedThroughput(provisionedThroughput)
		}
		result[i] = gsi
	}
	return result
}
