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

package controller

import (
	ctrl "sigs.k8s.io/controller-runtime"

	"github.com/crossplane-contrib/provider-aws/pkg/controller/acm"
	"github.com/crossplane-contrib/provider-aws/pkg/controller/acmpca"
	"github.com/crossplane-contrib/provider-aws/pkg/controller/apigateway"
	"github.com/crossplane-contrib/provider-aws/pkg/controller/apigatewayv2"
	"github.com/crossplane-contrib/provider-aws/pkg/controller/athena"
	"github.com/crossplane-contrib/provider-aws/pkg/controller/autoscaling"
	"github.com/crossplane-contrib/provider-aws/pkg/controller/batch"
	"github.com/crossplane-contrib/provider-aws/pkg/controller/cache"
	"github.com/crossplane-contrib/provider-aws/pkg/controller/cloudfront"
	"github.com/crossplane-contrib/provider-aws/pkg/controller/cloudsearch"
	"github.com/crossplane-contrib/provider-aws/pkg/controller/cloudwatchlogs"
	"github.com/crossplane-contrib/provider-aws/pkg/controller/cognitoidentity"
	"github.com/crossplane-contrib/provider-aws/pkg/controller/cognitoidentityprovider"
	"github.com/crossplane-contrib/provider-aws/pkg/controller/config"
	"github.com/crossplane-contrib/provider-aws/pkg/controller/database"
	"github.com/crossplane-contrib/provider-aws/pkg/controller/dax"
	"github.com/crossplane-contrib/provider-aws/pkg/controller/docdb"
	"github.com/crossplane-contrib/provider-aws/pkg/controller/dynamodb"
	"github.com/crossplane-contrib/provider-aws/pkg/controller/ec2"
	"github.com/crossplane-contrib/provider-aws/pkg/controller/ecr"
	"github.com/crossplane-contrib/provider-aws/pkg/controller/ecs"
	"github.com/crossplane-contrib/provider-aws/pkg/controller/efs"
	"github.com/crossplane-contrib/provider-aws/pkg/controller/eks"
	"github.com/crossplane-contrib/provider-aws/pkg/controller/elasticache"
	"github.com/crossplane-contrib/provider-aws/pkg/controller/elasticloadbalancing"
	"github.com/crossplane-contrib/provider-aws/pkg/controller/elbv2"
	"github.com/crossplane-contrib/provider-aws/pkg/controller/emrcontainers"
	"github.com/crossplane-contrib/provider-aws/pkg/controller/firehose"
	"github.com/crossplane-contrib/provider-aws/pkg/controller/globalaccelerator"
	"github.com/crossplane-contrib/provider-aws/pkg/controller/glue"
	"github.com/crossplane-contrib/provider-aws/pkg/controller/iam"
	"github.com/crossplane-contrib/provider-aws/pkg/controller/iot"
	"github.com/crossplane-contrib/provider-aws/pkg/controller/kafka"
	"github.com/crossplane-contrib/provider-aws/pkg/controller/kinesis"
	"github.com/crossplane-contrib/provider-aws/pkg/controller/kms"
	"github.com/crossplane-contrib/provider-aws/pkg/controller/lambda"
	"github.com/crossplane-contrib/provider-aws/pkg/controller/mq"
	"github.com/crossplane-contrib/provider-aws/pkg/controller/mwaa"
	"github.com/crossplane-contrib/provider-aws/pkg/controller/neptune"
	"github.com/crossplane-contrib/provider-aws/pkg/controller/opensearchservice"
	"github.com/crossplane-contrib/provider-aws/pkg/controller/prometheusservice"
	"github.com/crossplane-contrib/provider-aws/pkg/controller/ram"
	"github.com/crossplane-contrib/provider-aws/pkg/controller/rds"
	"github.com/crossplane-contrib/provider-aws/pkg/controller/redshift"
	"github.com/crossplane-contrib/provider-aws/pkg/controller/route53"
	"github.com/crossplane-contrib/provider-aws/pkg/controller/route53resolver"
	"github.com/crossplane-contrib/provider-aws/pkg/controller/s3"
	"github.com/crossplane-contrib/provider-aws/pkg/controller/s3control"
	"github.com/crossplane-contrib/provider-aws/pkg/controller/secretsmanager"
	"github.com/crossplane-contrib/provider-aws/pkg/controller/servicecatalog"
	"github.com/crossplane-contrib/provider-aws/pkg/controller/servicediscovery"
	"github.com/crossplane-contrib/provider-aws/pkg/controller/sesv2"
	"github.com/crossplane-contrib/provider-aws/pkg/controller/sfn"
	"github.com/crossplane-contrib/provider-aws/pkg/controller/sns"
	"github.com/crossplane-contrib/provider-aws/pkg/controller/sqs"
	"github.com/crossplane-contrib/provider-aws/pkg/controller/transfer"
	"github.com/crossplane-contrib/provider-aws/pkg/utils/controller"
	"github.com/crossplane-contrib/provider-aws/pkg/utils/setup"
)

// Setup creates all AWS controllers with the supplied logger and adds them to
// the supplied manager.
func Setup(mgr ctrl.Manager, o controller.Options) error {
	b := setup.NewBatch(mgr, o, "")
	b.AddUnscoped(acm.Setup)
	b.AddUnscoped(acmpca.Setup)
	b.AddUnscoped(apigateway.Setup)
	b.AddUnscoped(apigatewayv2.Setup)
	b.AddUnscoped(athena.Setup)
	b.AddUnscoped(autoscaling.Setup)
	b.AddUnscoped(batch.Setup)
	b.AddUnscoped(cache.Setup)
	b.AddUnscoped(cloudfront.Setup)
	b.AddUnscoped(cloudsearch.Setup)
	b.AddUnscoped(cloudwatchlogs.Setup)
	b.AddUnscoped(cognitoidentity.Setup)
	b.AddUnscoped(cognitoidentityprovider.Setup)
	b.AddUnscoped(config.Setup)
	b.AddUnscoped(database.Setup)
	b.AddUnscoped(dax.Setup)
	b.AddUnscoped(docdb.Setup)
	b.AddUnscoped(dynamodb.Setup)
	b.AddProxy(ec2.Setup)
	b.AddUnscoped(ecr.Setup)
	b.AddUnscoped(ecs.Setup)
	b.AddUnscoped(efs.Setup)
	b.AddUnscoped(eks.Setup)
	b.AddUnscoped(elasticache.Setup)
	b.AddUnscoped(elasticloadbalancing.Setup)
	b.AddUnscoped(elbv2.Setup)
	b.AddUnscoped(emrcontainers.Setup)
	b.AddUnscoped(firehose.Setup)
	b.AddUnscoped(glue.Setup)
	b.AddUnscoped(globalaccelerator.Setup)
	b.AddUnscoped(iam.Setup)
	b.AddUnscoped(iot.Setup)
	b.AddUnscoped(kafka.Setup)
	b.AddUnscoped(kinesis.Setup)
	b.AddUnscoped(kms.Setup)
	b.AddUnscoped(lambda.Setup)
	b.AddUnscoped(mq.Setup)
	b.AddUnscoped(mwaa.Setup)
	b.AddUnscoped(neptune.Setup)
	b.AddUnscoped(opensearchservice.Setup)
	b.AddUnscoped(prometheusservice.Setup)
	b.AddUnscoped(ram.Setup)
	b.AddUnscoped(rds.Setup)
	b.AddUnscoped(redshift.Setup)
	b.AddProxy(route53.Setup)
	b.AddUnscoped(route53resolver.Setup)
	b.AddUnscoped(s3.Setup)
	b.AddUnscoped(s3control.Setup)
	b.AddUnscoped(secretsmanager.Setup)
	b.AddUnscoped(servicecatalog.Setup)
	b.AddUnscoped(servicediscovery.Setup)
	b.AddUnscoped(sesv2.Setup)
	b.AddUnscoped(sfn.Setup)
	b.AddUnscoped(sns.Setup)
	b.AddUnscoped(sqs.Setup)
	b.AddUnscoped(transfer.Setup)
	return b.Run()
}
