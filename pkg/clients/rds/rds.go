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

package rds

import (
	"strings"

	awsclients "github.com/crossplaneio/stack-aws/pkg/clients"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/rds"
	"github.com/crossplaneio/stack-aws/apis/database/v1alpha2"
)

// Instance crossplane representation of the to AWS DBInstance
type Instance struct {
	Name     string
	ARN      string
	Status   string
	Endpoint string
}

// NewInstance returns new Instance structure
func NewInstance(instance *rds.DBInstance) *Instance {
	endpoint := ""
	if instance.Endpoint != nil {
		endpoint = aws.StringValue(instance.Endpoint.Address)
	}

	return &Instance{
		Name:     aws.StringValue(instance.DBInstanceIdentifier),
		ARN:      aws.StringValue(instance.DBInstanceArn),
		Status:   aws.StringValue(instance.DBInstanceStatus),
		Endpoint: endpoint,
	}
}

// Client defines RDS RDSClient operations
type Client interface {
	CreateDBInstanceRequest(*rds.CreateDBInstanceInput) rds.CreateDBInstanceRequest
	DescribeDBInstancesRequest(*rds.DescribeDBInstancesInput) rds.DescribeDBInstancesRequest
	ModifyDBInstanceRequest(*rds.ModifyDBInstanceInput) rds.ModifyDBInstanceRequest
	DeleteDBInstanceRequest(*rds.DeleteDBInstanceInput) rds.DeleteDBInstanceRequest
}

// NewClient creates new RDS RDSClient with provided AWS Configurations/Credentials
func NewClient(credentials []byte, region string) (Client, error) {
	cfg, err := awsclients.LoadConfig(credentials, awsclients.DefaultSection, region)
	if err != nil {
		return nil, err
	}
	return rds.New(*cfg), nil
}

// IsErrorAlreadyExists returns true if the supplied error indicates an instance
// already exists.
func IsErrorAlreadyExists(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(err.Error(), rds.ErrCodeDBInstanceAlreadyExistsFault)
}

// IsErrorNotFound helper function to test for ErrCodeDBInstanceNotFoundFault error
func IsErrorNotFound(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(err.Error(), rds.ErrCodeDBInstanceNotFoundFault)
}

// GenerateCreateDBInstanceInput from RDSInstanceSpec
func GenerateCreateDBInstanceInput(name, password string, spec *v1alpha2.RDSInstanceSpec) *rds.CreateDBInstanceInput {
	return &rds.CreateDBInstanceInput{
		DBInstanceIdentifier:  aws.String(name),
		AllocatedStorage:      aws.Int64(spec.ForProvider.Size),
		DBInstanceClass:       aws.String(spec.ForProvider.Class),
		Engine:                aws.String(spec.ForProvider.Engine),
		EngineVersion:         aws.String(spec.ForProvider.EngineVersion),
		MasterUsername:        aws.String(spec.ForProvider.MasterUsername),
		MasterUserPassword:    aws.String(password),
		BackupRetentionPeriod: aws.Int64(0),
		VpcSecurityGroupIds:   spec.ForProvider.SecurityGroupIDs,
		PubliclyAccessible:    aws.Bool(true),
		DBSubnetGroupName:     aws.String(spec.ForProvider.DBSubnetGroupName),
	}
}

// GenerateModifyDBInstanceInput from RDSInstanceSpec
func GenerateModifyDBInstanceInput(name string, spec *v1alpha2.RDSInstanceSpec) *rds.ModifyDBInstanceInput {
	return &rds.ModifyDBInstanceInput{
		DBInstanceIdentifier: aws.String(name),
		AllocatedStorage:     aws.Int64(spec.ForProvider.Size),
		DBInstanceClass:      aws.String(spec.ForProvider.Class),
		//Engine:                aws.String(spec.ForProvider.Engine),
		EngineVersion: aws.String(spec.ForProvider.EngineVersion),
		//MasterUsername:        aws.String(spec.ForProvider.MasterUsername),
		//MasterUserPassword:    aws.String(password),
		BackupRetentionPeriod: aws.Int64(0),
		VpcSecurityGroupIds:   spec.ForProvider.SecurityGroupIDs,
		PubliclyAccessible:    aws.Bool(true),
		DBSubnetGroupName:     aws.String(spec.ForProvider.DBSubnetGroupName),
	}
}
