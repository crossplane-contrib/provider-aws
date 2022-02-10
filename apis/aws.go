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

// Package apis contains Kubernetes API groups for AWS cloud provider.
package apis

import (
	"k8s.io/apimachinery/pkg/runtime"

	acmv1alpha1 "github.com/crossplane/provider-aws/apis/acm/v1alpha1"
	acmv1beta1 "github.com/crossplane/provider-aws/apis/acm/v1beta1"
	acmpcav1alpha1 "github.com/crossplane/provider-aws/apis/acmpca/v1alpha1"
	acmpcav1beta1 "github.com/crossplane/provider-aws/apis/acmpca/v1beta1"
	apigatewayv2 "github.com/crossplane/provider-aws/apis/apigatewayv2/v1alpha1"
	athenav1alpha1 "github.com/crossplane/provider-aws/apis/athena/v1alpha1"
	cachev1alpha1 "github.com/crossplane/provider-aws/apis/cache/v1alpha1"
	cachev1beta1 "github.com/crossplane/provider-aws/apis/cache/v1beta1"
	cloudfrontv1alpha1 "github.com/crossplane/provider-aws/apis/cloudfront/v1alpha1"
	cloudwatchlogsv1alpha1 "github.com/crossplane/provider-aws/apis/cloudwatchlogs/v1alpha1"
	databasev1beta1 "github.com/crossplane/provider-aws/apis/database/v1beta1"
	docdbv1alpha1 "github.com/crossplane/provider-aws/apis/docdb/v1alpha1"
	dynamodbv1alpha1 "github.com/crossplane/provider-aws/apis/dynamodb/v1alpha1"
	ec2manualv1alpha1 "github.com/crossplane/provider-aws/apis/ec2/manualv1alpha1"
	ec2v1alpha1 "github.com/crossplane/provider-aws/apis/ec2/v1alpha1"
	ec2v1beta1 "github.com/crossplane/provider-aws/apis/ec2/v1beta1"
	ecrv1alpha1 "github.com/crossplane/provider-aws/apis/ecr/v1alpha1"
	ecrv1beta1 "github.com/crossplane/provider-aws/apis/ecr/v1beta1"
	efsv1alpha1 "github.com/crossplane/provider-aws/apis/efs/v1alpha1"
	eksmanualv1alpha1 "github.com/crossplane/provider-aws/apis/eks/manualv1alpha1"
	eksv1alpha1 "github.com/crossplane/provider-aws/apis/eks/v1alpha1"
	eksv1beta1 "github.com/crossplane/provider-aws/apis/eks/v1beta1"
	elasticachev1alpha1 "github.com/crossplane/provider-aws/apis/elasticache/v1alpha1"
	elasticloadbalancingv1alpha1 "github.com/crossplane/provider-aws/apis/elasticloadbalancing/v1alpha1"
	elbv2v1alpha1 "github.com/crossplane/provider-aws/apis/elbv2/v1alpha1"
	gluev1alpha1 "github.com/crossplane/provider-aws/apis/glue/v1alpha1"
	iamv1beta1 "github.com/crossplane/provider-aws/apis/iam/v1beta1"
	iotv1alpha1 "github.com/crossplane/provider-aws/apis/iot/v1alpha1"
	kafkav1alpha1 "github.com/crossplane/provider-aws/apis/kafka/v1alpha1"
	kinesisv1alpha1 "github.com/crossplane/provider-aws/apis/kinesis/v1alpha1"
	kmsv1alpha1 "github.com/crossplane/provider-aws/apis/kms/v1alpha1"
	lambdav1alpha1 "github.com/crossplane/provider-aws/apis/lambda/v1alpha1"
	mqv1alpha1 "github.com/crossplane/provider-aws/apis/mq/v1alpha1"
	neptunev1alpha1 "github.com/crossplane/provider-aws/apis/neptune/v1alpha1"
	notificationv1alpha1 "github.com/crossplane/provider-aws/apis/notification/v1alpha1"
	ramv1alpha1 "github.com/crossplane/provider-aws/apis/ram/v1alpha1"
	rdsv1alpha1 "github.com/crossplane/provider-aws/apis/rds/v1alpha1"
	redshiftv1alpha1 "github.com/crossplane/provider-aws/apis/redshift/v1alpha1"
	route53v1alpha1 "github.com/crossplane/provider-aws/apis/route53/v1alpha1"
	route53resolvermanualv1alpha1 "github.com/crossplane/provider-aws/apis/route53resolver/manualv1alpha1"
	route53resolverv1alpha1 "github.com/crossplane/provider-aws/apis/route53resolver/v1alpha1"
	s3v1alpha2 "github.com/crossplane/provider-aws/apis/s3/v1alpha3"
	s3v1beta1 "github.com/crossplane/provider-aws/apis/s3/v1beta1"
	secretsmanagerv1alpha1 "github.com/crossplane/provider-aws/apis/secretsmanager/v1alpha1"
	servicediscoveryv1alpha1 "github.com/crossplane/provider-aws/apis/servicediscovery/v1alpha1"
	sfnv1alpha1 "github.com/crossplane/provider-aws/apis/sfn/v1alpha1"
	snsv1beta1 "github.com/crossplane/provider-aws/apis/sns/v1beta1"
	sqsv1beta1 "github.com/crossplane/provider-aws/apis/sqs/v1beta1"
	transferv1alpha1 "github.com/crossplane/provider-aws/apis/transfer/v1alpha1"
	awsv1alpha3 "github.com/crossplane/provider-aws/apis/v1alpha3"
	awsv1beta1 "github.com/crossplane/provider-aws/apis/v1beta1"
)

func init() {
	// Register the types with the Scheme so the components can map objects to GroupVersionKinds and back
	AddToSchemes = append(AddToSchemes,
		cachev1alpha1.SchemeBuilder.AddToScheme,
		cachev1beta1.SchemeBuilder.AddToScheme,
		databasev1beta1.SchemeBuilder.AddToScheme,
		docdbv1alpha1.AddToScheme,
		elasticloadbalancingv1alpha1.SchemeBuilder.AddToScheme,
		iamv1beta1.SchemeBuilder.AddToScheme,
		elasticachev1alpha1.SchemeBuilder.AddToScheme,
		elbv2v1alpha1.SchemeBuilder.AddToScheme,
		route53v1alpha1.SchemeBuilder.AddToScheme,
		notificationv1alpha1.SchemeBuilder.AddToScheme,
		ec2v1beta1.SchemeBuilder.AddToScheme,
		awsv1alpha3.SchemeBuilder.AddToScheme,
		awsv1beta1.SchemeBuilder.AddToScheme,
		acmv1alpha1.SchemeBuilder.AddToScheme,
		acmv1beta1.SchemeBuilder.AddToScheme,
		s3v1alpha2.SchemeBuilder.AddToScheme,
		s3v1beta1.SchemeBuilder.AddToScheme,
		secretsmanagerv1alpha1.SchemeBuilder.AddToScheme,
		servicediscoveryv1alpha1.SchemeBuilder.AddToScheme,
		acmpcav1alpha1.SchemeBuilder.AddToScheme,
		acmpcav1beta1.SchemeBuilder.AddToScheme,
		eksv1alpha1.SchemeBuilder.AddToScheme,
		eksv1beta1.SchemeBuilder.AddToScheme,
		sqsv1beta1.SchemeBuilder.AddToScheme,
		redshiftv1alpha1.SchemeBuilder.AddToScheme,
		eksmanualv1alpha1.SchemeBuilder.AddToScheme,
		ecrv1alpha1.SchemeBuilder.AddToScheme,
		ecrv1beta1.SchemeBuilder.AddToScheme,
		apigatewayv2.SchemeBuilder.AddToScheme,
		sfnv1alpha1.SchemeBuilder.AddToScheme,
		dynamodbv1alpha1.SchemeBuilder.AddToScheme,
		kmsv1alpha1.SchemeBuilder.AddToScheme,
		efsv1alpha1.SchemeBuilder.AddToScheme,
		rdsv1alpha1.SchemeBuilder.AddToScheme,
		ec2manualv1alpha1.SchemeBuilder.AddToScheme,
		ec2v1alpha1.SchemeBuilder.AddToScheme,
		lambdav1alpha1.SchemeBuilder.AddToScheme,
		cloudfrontv1alpha1.SchemeBuilder.AddToScheme,
		route53resolverv1alpha1.SchemeBuilder.AddToScheme,
		route53resolvermanualv1alpha1.SchemeBuilder.AddToScheme,
		kafkav1alpha1.SchemeBuilder.AddToScheme,
		transferv1alpha1.SchemeBuilder.AddToScheme,
		gluev1alpha1.SchemeBuilder.AddToScheme,
		mqv1alpha1.SchemeBuilder.AddToScheme,
		cloudwatchlogsv1alpha1.SchemeBuilder.AddToScheme,
		iotv1alpha1.SchemeBuilder.AddToScheme,
		athenav1alpha1.SchemeBuilder.AddToScheme,
		ramv1alpha1.SchemeBuilder.AddToScheme,
		kinesisv1alpha1.SchemeBuilder.AddToScheme,
		neptunev1alpha1.SchemeBuilder.AddToScheme,
		snsv1beta1.SchemeBuilder.AddToScheme,
	)
}

// AddToSchemes may be used to add all resources defined in the project to a Scheme
var AddToSchemes runtime.SchemeBuilder

// AddToScheme adds all Resources to the Scheme
func AddToScheme(s *runtime.Scheme) error {
	return AddToSchemes.AddToScheme(s)
}
