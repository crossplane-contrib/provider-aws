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
	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/connection"
	"github.com/crossplane/crossplane-runtime/pkg/controller"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"

	svcapitypes "github.com/crossplane-contrib/provider-aws/apis/dynamodb/v1alpha1"
	"github.com/crossplane-contrib/provider-aws/apis/v1alpha1"
	aws "github.com/crossplane-contrib/provider-aws/pkg/clients"
	"github.com/crossplane-contrib/provider-aws/pkg/features"
)

// ManagesKind returns the kind this controller manages
func ManagesKind() string {
	return svcapitypes.TableGroupKind
}

// SetupTable adds a controller that reconciles Table.
func SetupTable(mgr ctrl.Manager, o controller.Options) error {
	name := managed.ControllerName(svcapitypes.TableGroupKind)
	opts := []option{
		func(e *external) {
			e.preObserve = preObserve
			e.postObserve = postObserve
			e.preCreate = preCreate
			e.preDelete = preDelete
			e.postDelete = postDelete
			e.lateInitialize = lateInitialize
			e.isUpToDate = isUpToDate
			u := &updateClient{client: e.client}
			e.preUpdate = u.preUpdate
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
		For(&svcapitypes.Table{}).
		Complete(managed.NewReconciler(mgr,
			resource.ManagedKind(svcapitypes.TableGroupVersionKind),
			managed.WithExternalConnecter(&connector{kube: mgr.GetClient(), opts: opts}),
			managed.WithInitializers(
				managed.NewNameAsExternalName(mgr.GetClient()),
				managed.NewDefaultProviderConfig(mgr.GetClient()),
				&tagger{kube: mgr.GetClient()}),
			managed.WithPollInterval(o.PollInterval),
			managed.WithLogger(o.Logger.WithValues("controller", name)),
			managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name))),
			managed.WithConnectionPublishers(cps...)))
}

func preObserve(_ context.Context, cr *svcapitypes.Table, obj *svcsdk.DescribeTableInput) error {
	obj.TableName = aws.String(meta.GetExternalName(cr))
	return nil
}
func preCreate(_ context.Context, cr *svcapitypes.Table, obj *svcsdk.CreateTableInput) error {
	obj.TableName = aws.String(meta.GetExternalName(cr))
	return nil
}
func preDelete(_ context.Context, cr *svcapitypes.Table, obj *svcsdk.DeleteTableInput) (bool, error) {
	obj.TableName = aws.String(meta.GetExternalName(cr))
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
	switch aws.StringValue(resp.Table.TableStatus) {
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
		"tableArn":          []byte(aws.StringValue(resp.Table.TableArn)),
		"latestStreamArn":   []byte(aws.StringValue(resp.Table.LatestStreamArn)),
		"latestStreamLabel": []byte(aws.StringValue(resp.Table.LatestStreamLabel)),
	}

	return obs, nil
}

type tagger struct {
	kube client.Client
}

func (e *tagger) Initialize(ctx context.Context, mg resource.Managed) error {
	cr, ok := mg.(*svcapitypes.Table)
	if !ok {
		return errors.New(errUnexpectedObject)
	}
	tagMap := map[string]string{}
	for _, t := range cr.Spec.ForProvider.Tags {
		tagMap[aws.StringValue(t.Key)] = aws.StringValue(t.Value)
	}
	for k, v := range resource.GetExternalTags(cr) {
		tagMap[k] = v
	}
	tags := make([]*svcapitypes.Tag, 0)
	for k, v := range tagMap {
		tags = append(tags, &svcapitypes.Tag{Key: aws.String(k), Value: aws.String(v)})
	}
	sort.Slice(tags, func(i, j int) bool {
		return aws.StringValue(tags[i].Key) < aws.StringValue(tags[j].Key)
	})
	if cmp.Equal(cr.Spec.ForProvider.Tags, tags) {
		return nil
	}
	cr.Spec.ForProvider.Tags = tags
	return errors.Wrap(e.kube.Update(ctx, cr), "cannot update Table Spec")
}

func lateInitialize(in *svcapitypes.TableParameters, t *svcsdk.DescribeTableOutput) error { // nolint:gocyclo,unparam
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
		in.BillingMode = aws.String(svcsdk.BillingModeProvisioned)
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
			in.SSESpecification.Enabled = aws.Bool(*t.Table.SSEDescription.Status == string(svcapitypes.SSEStatus_ENABLED))
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
		in.StreamSpecification = &svcapitypes.StreamSpecification{StreamEnabled: aws.Bool(false, aws.FieldRequired)}
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
func createPatch(in *svcsdk.DescribeTableOutput, target *svcapitypes.TableParameters) (*svcapitypes.TableParameters, error) {
	currentParams := &svcapitypes.TableParameters{}
	if err := lateInitialize(currentParams, in); err != nil {
		return nil, err
	}

	jsonPatch, err := aws.CreateJSONPatch(currentParams, target)
	if err != nil {
		return nil, err
	}
	patch := &svcapitypes.TableParameters{}
	if err := json.Unmarshal(jsonPatch, patch); err != nil {
		return nil, err
	}
	return patch, nil
}

func isUpToDate(cr *svcapitypes.Table, resp *svcsdk.DescribeTableOutput) (bool, error) {
	// A table that's currently creating, deleting, or updating can't be
	// updated, so we temporarily consider it to be up-to-date no matter
	// what.
	switch aws.StringValue(cr.Status.AtProvider.TableStatus) {
	case string(svcapitypes.TableStatus_SDK_UPDATING), string(svcapitypes.TableStatus_SDK_CREATING), string(svcapitypes.TableStatus_SDK_DELETING):
		return true, nil
	}

	// Similarly, a table that's currently updating its SSE status can't be
	// updated, so we temporarily consider it to be up-to-date.
	if cr.Status.AtProvider.SSEDescription != nil && aws.StringValue(cr.Status.AtProvider.SSEDescription.Status) == string(svcapitypes.SSEStatus_UPDATING) {
		return true, nil
	}

	patch, err := createPatch(resp, &cr.Spec.ForProvider)
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
	case len(diffGlobalSecondaryIndexes(GenerateGlobalSecondaryIndexDescriptions(cr.Spec.ForProvider.GlobalSecondaryIndexes), resp.Table.GlobalSecondaryIndexes)) != 0:
		return false, nil
	}
	return true, nil
}

type updateClient struct {
	client svcsdkapi.DynamoDBAPI
}

func (e *updateClient) preUpdate(ctx context.Context, cr *svcapitypes.Table, u *svcsdk.UpdateTableInput) error {
	filtered := &svcsdk.UpdateTableInput{
		TableName:            aws.String(meta.GetExternalName(cr)),
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
	out, err := e.client.DescribeTableWithContext(ctx, &svcsdk.DescribeTableInput{TableName: aws.String(meta.GetExternalName(cr))})
	if err != nil {
		return aws.Wrap(err, errDescribe)
	}

	p, err := createPatch(out, &cr.Spec.ForProvider)
	if err != nil {
		return err
	}
	gsiUpdates := diffGlobalSecondaryIndexes(GenerateGlobalSecondaryIndexDescriptions(cr.Spec.ForProvider.GlobalSecondaryIndexes), out.Table.GlobalSecondaryIndexes)
	switch {
	case p.BillingMode != nil:
		filtered.BillingMode = u.BillingMode

		// NOTE(negz): You must include provisioned throughput when
		// updating the billing mode to PROVISIONED.
		if aws.StringValue(u.BillingMode) == string(svcapitypes.BillingMode_PROVISIONED) {
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
		desired[aws.StringValue(gsi.IndexName)] = gsi
		desiredKeys[i] = aws.StringValue(gsi.IndexName)
	}
	existing := map[string]*svcsdk.GlobalSecondaryIndexDescription{}
	existingKeys := make([]string, len(obs))
	for i, gsi := range obs {
		existing[aws.StringValue(gsi.IndexName)] = gsi
		existingKeys[i] = aws.StringValue(gsi.IndexName)
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
			if aws.Int64Value(desired[k].ProvisionedThroughput.WriteCapacityUnits) != aws.Int64Value(existingGSI.ProvisionedThroughput.WriteCapacityUnits) ||
				aws.Int64Value(desired[k].ProvisionedThroughput.ReadCapacityUnits) != aws.Int64Value(existingGSI.ProvisionedThroughput.ReadCapacityUnits) {
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
func GenerateGlobalSecondaryIndexDescriptions(p []*svcapitypes.GlobalSecondaryIndex) []*svcsdk.GlobalSecondaryIndexDescription { // nolint:gocyclo
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
