# provider-aws

## Overview

This `provider-aws` repository is forked from https://github.com/crossplane/provider-aws.
We just start some controllers to avoid to apply all CRD.
Currently, [infra-cd](https://github.com/tidbcloud/infra-cd) apply following crds:
```antlrv4
elasticloadbalancing.aws.crossplane.io_elbs.yaml
elasticloadbalancing.aws.crossplane.io_elbattachments.yaml
kms.aws.crossplane.io_keys.yaml
route53.aws.crossplane.io_hostedzones.yaml
route53.aws.crossplane.io_resourcerecordsets.yaml
s3.aws.crossplane.io_buckets.yaml
s3.aws.crossplane.io_bucketpolicies.yaml
aws.crossplane.io_providerconfigs.yaml
identity.aws.crossplane.io_iamaccesskeys.yaml
identity.aws.crossplane.io_iamrolepolicyattachments.yaml
identity.aws.crossplane.io_iamroles.yaml
identity.aws.crossplane.io_iampolicies.yaml
ec2.aws.crossplane.io_vpcs.yaml
sqs.aws.crossplane.io_queues.yaml
aws.crossplane.io_providerconfigusages.yaml
vpcpeering.aws.crossplane.io_vpcpeeringconnections.yaml
```

# Build Docker Image
## Build binary
```shell
GOOS=linux go build -o docker/crossplane-aws-provider cmd/provider/main.go
```
## Build image
```shell
cd docker
docker build -t $REGISTRY/provider-aws:v0.19.2-dev .
```