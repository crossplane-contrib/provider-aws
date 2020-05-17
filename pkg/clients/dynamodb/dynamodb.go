/*
Copyright 2019 The Crossplane Authors.

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

package dynamodb

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/google/go-cmp/cmp"

	"github.com/crossplane/provider-aws/apis/database/v1alpha1"
	awsclients "github.com/crossplane/provider-aws/pkg/clients"
)

// Client defines DyanmoDB client operations
type Client interface {
	DescribeTableRequest(input *dynamodb.DescribeTableInput) dynamodb.DescribeTableRequest
	CreateTableRequest(input *dynamodb.CreateTableInput) dynamodb.CreateTableRequest
	DeleteTableRequest(input *dynamodb.DeleteTableInput) dynamodb.DeleteTableRequest
	UpdateTableRequest(input *dynamodb.UpdateTableInput) dynamodb.UpdateTableRequest
}

// NewClient creates new DynamoDB Client with provided AWS Configurations/Credentials
func NewClient(ctx context.Context, credentials []byte, region string, auth awsclients.AuthMethod) (Client, error) {
	cfg, err := auth(ctx, credentials, awsclients.DefaultSection, region)
	if cfg == nil {
		return nil, err
	}
	return dynamodb.New(*cfg), err
}

// LateInitialize fills the empty fields in *v1alpha1.DynamoTableParameters with
// the values seen in dynamodb.TableDescription.
func LateInitialize(in *v1alpha1.DynamoTableParameters, t *dynamodb.TableDescription) { // nolint:gocyclo
	if t == nil {
		return
	}

	if len(in.AttributeDefinitions) == 0 && len(t.AttributeDefinitions) != 0 {
		in.AttributeDefinitions = buildAttributeDefinitions(t.AttributeDefinitions)
	}

	if len(in.GlobalSecondaryIndexes) == 0 && len(t.GlobalSecondaryIndexes) != 0 {
		in.GlobalSecondaryIndexes = buildGlobalIndexes(t.GlobalSecondaryIndexes)
	}

	if len(in.LocalSecondaryIndexes) == 0 && len(t.LocalSecondaryIndexes) != 0 {
		in.LocalSecondaryIndexes = buildLocalIndexes(t.LocalSecondaryIndexes)
	}

	if len(in.KeySchema) == 0 && len(t.KeySchema) != 0 {
		in.KeySchema = buildAlphaKeyElements(t.KeySchema)
	}

	if t.ProvisionedThroughput != nil {
		in.ProvisionedThroughput = &v1alpha1.ProvisionedThroughput{
			ReadCapacityUnits:  t.ProvisionedThroughput.ReadCapacityUnits,
			WriteCapacityUnits: t.ProvisionedThroughput.WriteCapacityUnits,
		}
	}
}

// CreatePatch creates a *v1alpha1.DynamoTableParameters that has only the changed
// values between the target *v1alpha1.DynamoTableParameters and the current
// *dynamodb.TableDescription
func CreatePatch(in *dynamodb.TableDescription, target *v1alpha1.DynamoTableParameters) (*v1alpha1.DynamoTableParameters, error) {
	currentParams := &v1alpha1.DynamoTableParameters{}
	LateInitialize(currentParams, in)

	jsonPatch, err := awsclients.CreateJSONPatch(currentParams, target)
	if err != nil {
		return nil, err
	}
	patch := &v1alpha1.DynamoTableParameters{}
	if err := json.Unmarshal(jsonPatch, patch); err != nil {
		return nil, err
	}
	return patch, nil
}

// GenerateCreateTableInput from DynamoTaleSpec
func GenerateCreateTableInput(name string, p *v1alpha1.DynamoTableParameters) *dynamodb.CreateTableInput {
	c := &dynamodb.CreateTableInput{
		KeySchema: buildTableKeyElements(p.KeySchema),
		TableName: aws.String(name),
		Tags:      buildDynamoTags(p.Tags),
	}

	if len(p.AttributeDefinitions) != 0 {
		c.AttributeDefinitions = make([]dynamodb.AttributeDefinition, len(p.AttributeDefinitions))
		for i, val := range p.AttributeDefinitions {
			c.AttributeDefinitions[i] = dynamodb.AttributeDefinition{
				AttributeName: aws.String(val.AttributeName),
				AttributeType: dynamodb.ScalarAttributeType(val.AttributeType),
			}
		}
	}

	if len(p.GlobalSecondaryIndexes) != 0 {
		c.GlobalSecondaryIndexes = make([]dynamodb.GlobalSecondaryIndex, len(p.GlobalSecondaryIndexes))
		for i, val := range p.GlobalSecondaryIndexes {
			c.GlobalSecondaryIndexes[i] = dynamodb.GlobalSecondaryIndex{
				IndexName: val.IndexName,
				KeySchema: buildTableKeyElements(val.KeySchema),
				Projection: &dynamodb.Projection{
					NonKeyAttributes: val.Projection.NonKeyAttributes,
					ProjectionType:   dynamodb.ProjectionType(val.Projection.ProjectionType),
				},
			}
		}
	}

	if p.ProvisionedThroughput != nil {
		c.ProvisionedThroughput = &dynamodb.ProvisionedThroughput{
			ReadCapacityUnits:  p.ProvisionedThroughput.ReadCapacityUnits,
			WriteCapacityUnits: p.ProvisionedThroughput.WriteCapacityUnits,
		}
	}

	if p.SSESpecification != nil {
		c.SSESpecification = &dynamodb.SSESpecification{
			Enabled:        p.StreamSpecification.StreamEnabled,
			KMSMasterKeyId: p.SSESpecification.SSEType,
			SSEType:        dynamodb.SSEType(*p.SSESpecification.SSEType),
		}
	}

	if p.StreamSpecification != nil {
		c.StreamSpecification = &dynamodb.StreamSpecification{
			StreamEnabled:  p.StreamSpecification.StreamEnabled,
			StreamViewType: dynamodb.StreamViewType(*p.SSESpecification.SSEType),
		}
	}

	return c
}

// GenerateUpdateTableInput from DynamoTaleSpec
func GenerateUpdateTableInput(name string, p *v1alpha1.DynamoTableParameters) *dynamodb.UpdateTableInput {
	u := &dynamodb.UpdateTableInput{
		TableName: aws.String(name),
	}

	if len(p.AttributeDefinitions) != 0 {
		u.AttributeDefinitions = make([]dynamodb.AttributeDefinition, len(p.AttributeDefinitions))
		for i, val := range p.AttributeDefinitions {
			u.AttributeDefinitions[i] = dynamodb.AttributeDefinition{
				AttributeName: aws.String(val.AttributeName),
				AttributeType: dynamodb.ScalarAttributeType(val.AttributeType),
			}
		}
	}

	if p.ProvisionedThroughput != nil {
		u.ProvisionedThroughput = &dynamodb.ProvisionedThroughput{
			ReadCapacityUnits:  p.ProvisionedThroughput.ReadCapacityUnits,
			WriteCapacityUnits: p.ProvisionedThroughput.WriteCapacityUnits,
		}
	}

	if p.SSESpecification != nil {
		u.SSESpecification = &dynamodb.SSESpecification{
			Enabled:        p.StreamSpecification.StreamEnabled,
			KMSMasterKeyId: p.SSESpecification.SSEType,
			SSEType:        dynamodb.SSEType(*p.SSESpecification.SSEType),
		}
	}

	if p.StreamSpecification != nil {
		u.StreamSpecification = &dynamodb.StreamSpecification{
			StreamEnabled:  p.StreamSpecification.StreamEnabled,
			StreamViewType: dynamodb.StreamViewType(*p.SSESpecification.SSEType),
		}
	}

	return u
}

// GenerateObservation is used to produce v1alpha1.DynamoTableObservation from
// dynamodb.TableDescription.
func GenerateObservation(t dynamodb.TableDescription) v1alpha1.DynamoTableObservation { // nolint:gocyclo

	o := v1alpha1.DynamoTableObservation{
		AttributeDefinitions:   buildAttributeDefinitions(t.AttributeDefinitions),
		GlobalSecondaryIndexes: buildGlobalIndexes(t.GlobalSecondaryIndexes),
		LocalSecondaryIndexes:  buildLocalIndexes(t.LocalSecondaryIndexes),
		ItemCount:              aws.Int64Value(t.ItemCount),
		KeySchema:              buildAlphaKeyElements(t.KeySchema),
		TableArn:               aws.StringValue(t.TableArn),
		TableID:                aws.StringValue(t.TableId),
		TableStatus:            string(t.TableStatus),
		TableName:              aws.StringValue(t.TableName),
	}

	if t.ProvisionedThroughput != nil {
		o.ProvisionedThroughput = v1alpha1.ProvisionedThroughput{
			ReadCapacityUnits:  t.ProvisionedThroughput.ReadCapacityUnits,
			WriteCapacityUnits: t.ProvisionedThroughput.WriteCapacityUnits,
		}
	}

	return o
}

// IsUpToDate checks whether there is a change in any of the modifiable fields.
func IsUpToDate(p v1alpha1.DynamoTableParameters, t dynamodb.TableDescription) (bool, error) {

	patch, err := CreatePatch(&t, &p)
	if err != nil {
		return false, err
	}
	return cmp.Equal(&v1alpha1.DynamoTableParameters{}, patch), nil
}

// IsErrorNotFound helper function to test for ErrCodeTableNotFoundException error
func IsErrorNotFound(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(err.Error(), dynamodb.ErrCodeResourceNotFoundException)
}

func buildDynamoTags(tags []v1alpha1.Tag) []dynamodb.Tag {
	if len(tags) == 0 {
		return nil
	}

	res := make([]dynamodb.Tag, len(tags))
	for i, t := range tags {
		res[i] = dynamodb.Tag{
			Key:   aws.String(t.Key),
			Value: aws.String(t.Value),
		}
	}

	return res
}

func buildAlphaKeyElements(keys []dynamodb.KeySchemaElement) []v1alpha1.KeySchemaElement {
	if len(keys) == 0 {
		return nil
	}
	keyElements := make([]v1alpha1.KeySchemaElement, len(keys))
	for i, val := range keys {
		keyElements[i] = v1alpha1.KeySchemaElement{
			AttributeName: aws.StringValue(val.AttributeName),
			KeyType:       string(val.KeyType),
		}
	}
	return keyElements
}

func buildTableKeyElements(keys []v1alpha1.KeySchemaElement) []dynamodb.KeySchemaElement {
	if len(keys) == 0 {
		return nil
	}
	keyElements := make([]dynamodb.KeySchemaElement, len(keys))
	for i, val := range keys {
		keyElements[i] = dynamodb.KeySchemaElement{
			AttributeName: aws.String(val.AttributeName),
			KeyType:       dynamodb.KeyType(val.KeyType),
		}
	}
	return keyElements
}

func buildAttributeDefinitions(attributes []dynamodb.AttributeDefinition) []v1alpha1.AttributeDefinition {
	if len(attributes) == 0 {
		return nil
	}
	attributeDefinitions := make([]v1alpha1.AttributeDefinition, len(attributes))
	for i, val := range attributes {
		attributeDefinitions[i] = v1alpha1.AttributeDefinition{
			AttributeName: *val.AttributeName,
			AttributeType: string(val.AttributeType),
		}
	}
	return attributeDefinitions
}

func buildGlobalIndexes(indexes []dynamodb.GlobalSecondaryIndexDescription) []v1alpha1.GlobalSecondaryIndex {
	if len(indexes) == 0 {
		return nil
	}
	globalSecondaryIndexes := make([]v1alpha1.GlobalSecondaryIndex, len(indexes))
	for i, val := range indexes {
		globalSecondaryIndexes[i] = v1alpha1.GlobalSecondaryIndex{
			IndexName: val.IndexName,
			KeySchema: buildAlphaKeyElements(val.KeySchema),
			Projection: &v1alpha1.Projection{
				NonKeyAttributes: val.Projection.NonKeyAttributes,
				ProjectionType:   string(val.Projection.ProjectionType),
			},
		}
	}
	return globalSecondaryIndexes
}

func buildLocalIndexes(indexes []dynamodb.LocalSecondaryIndexDescription) []v1alpha1.LocalSecondaryIndex {
	if len(indexes) == 0 {
		return nil
	}
	localSecondaryIndexes := make([]v1alpha1.LocalSecondaryIndex, len(indexes))
	for i, val := range indexes {
		localSecondaryIndexes[i] = v1alpha1.LocalSecondaryIndex{
			IndexName: val.IndexName,
			KeySchema: buildAlphaKeyElements(val.KeySchema),
			Projection: &v1alpha1.Projection{
				NonKeyAttributes: val.Projection.NonKeyAttributes,
				ProjectionType:   string(val.Projection.ProjectionType),
			},
		}
	}
	return localSecondaryIndexes
}
