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
func Setup(mgr ctrl.Manager, o controller.OptionsSet) error {
	b := setup.NewBatch(mgr, o, "")
	b.AddProxyXp(acm.Setup)
	b.AddProxyXp(acmpca.Setup)
	b.AddProxyXp(apigateway.Setup)
	b.AddProxyXp(apigatewayv2.Setup)
	b.AddProxyXp(athena.Setup)
	b.AddProxyXp(autoscaling.Setup)
	b.AddProxyXp(batch.Setup)
	b.AddProxyXp(cache.Setup)
	b.AddProxyXp(cloudfront.Setup)
	b.AddProxyXp(cloudsearch.Setup)
	b.AddProxyXp(cloudwatchlogs.Setup)
	b.AddProxyXp(cognitoidentity.Setup)
	b.AddProxyXp(cognitoidentityprovider.Setup)
	b.AddProxyXp(config.Setup)
	b.AddProxyXp(database.Setup)
	b.AddProxyXp(dax.Setup)
	b.AddProxyXp(docdb.Setup)
	b.AddProxyXp(dynamodb.Setup)
	b.AddProxy(ec2.Setup)
	b.AddProxyXp(ecr.Setup)
	b.AddProxyXp(ecs.Setup)
	b.AddProxyXp(efs.Setup)
	b.AddProxyXp(eks.Setup)
	b.AddProxyXp(elasticache.Setup)
	b.AddProxyXp(elasticloadbalancing.Setup)
	b.AddProxyXp(elbv2.Setup)
	b.AddProxyXp(emrcontainers.Setup)
	b.AddProxyXp(firehose.Setup)
	b.AddProxyXp(glue.Setup)
	b.AddProxyXp(globalaccelerator.Setup)
	b.AddProxyXp(iam.Setup)
	b.AddProxyXp(iot.Setup)
	b.AddProxyXp(kafka.Setup)
	b.AddProxyXp(kinesis.Setup)
	b.AddProxyXp(kms.Setup)
	b.AddProxyXp(lambda.Setup)
	b.AddProxyXp(mq.Setup)
	b.AddProxyXp(mwaa.Setup)
	b.AddProxyXp(neptune.Setup)
	b.AddProxyXp(opensearchservice.Setup)
	b.AddProxyXp(prometheusservice.Setup)
	b.AddProxyXp(ram.Setup)
	b.AddProxyXp(rds.Setup)
	b.AddProxyXp(redshift.Setup)
	b.AddProxy(route53.Setup)
	b.AddProxyXp(route53resolver.Setup)
	b.AddProxyXp(s3.Setup)
	b.AddProxyXp(s3control.Setup)
	b.AddProxyXp(secretsmanager.Setup)
	b.AddProxyXp(servicecatalog.Setup)
	b.AddProxyXp(servicediscovery.Setup)
	b.AddProxyXp(sesv2.Setup)
	b.AddProxyXp(sfn.Setup)
	b.AddProxyXp(sns.Setup)
	b.AddProxyXp(sqs.Setup)
	b.AddProxyXp(transfer.Setup)
	return b.Run()
}
