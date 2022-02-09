# Authenticating to AWS API

## Table of Contents

- [Overview](#overview)
- [Using IAM Roles for `ServiceAccounts`](#using-iam-roles-for-serviceaccounts)
- [Using kube2iam](#using-kube2iam)
- [Using `assumeRoleARN`](#using-assumerolearn)

## Overview

`provider-aws` requires credentials to be provided in order to authenticate to the
AWS API. This can be done in one of the following ways:

- Using a base64-encoded static credentials in a Kubernetes
  `Secret`. This is described in detail
  [here](https://crossplane.io/docs/v1.6/getting-started/install-configure.html#get-aws-account-keyfile).
- Using IAM Roles for Kubernetes `ServiceAccounts`. This
  functionality is only available when running Crossplane on EKS, and the
  feature has been
  [enabled](https://docs.aws.amazon.com/eks/latest/userguide/iam-roles-for-service-accounts.html)
  in the cluster.
- Using [kube2iam](https://github.com/jtblin/kube2iam). This solution allows
  to avoid using static credentials with non-EKS cluster.
- Using `assumeRoleARN` with the `provider-aws` to connect to
  other AWS accounts via one AWS account.

## Using IAM Roles for `ServiceAccounts`

Using IAM Roles for `ServiceAccounts` requires some additional setup for the
time-being. The steps for enabling are described below. Many of the steps can
also be found in the [AWS docs](https://docs.aws.amazon.com/eks/latest/userguide/enable-iam-roles-for-service-accounts.html).

### Steps

These steps assume you already have a running EKS cluster with a sufficiently
large node pool.

1. Connect to your EKS cluster

```console
$ aws eks --region "${AWS_REGION}" update-kubeconfig --name "${CLUSTER_NAME}"
```

2. Get AWS account information

Get AWS account information and pick an IAM role name. These will be used to
setup an OIDC provider and inject credentials into the `provider-aws` controller
`Pod`.

```console
$ AWS_ACCOUNT_ID=$(aws sts get-caller-identity --query "Account" --output text)
$ IAM_ROLE_NAME=provider-aws # name for IAM role, can be anything you want
```

3. Install Crossplane

Install Crossplane from `stable` channel:

```console
$ helm repo add crossplane-stable https://charts.crossplane.io/stable
$ helm install crossplane --create-namespace --namespace crossplane-system crossplane-stable/crossplane
```

`provider-aws` can be installed with the [Crossplane CLI](https://crossplane.io/docs/v1.6/getting-started/install-configure.html#install-crossplane-cli),
but we will do so manually so that we can also create and reference a
`ControllerConfig`:

```console
$ cat <<EOF | kubectl apply -f -
apiVersion: pkg.crossplane.io/v1alpha1
kind: ControllerConfig
metadata:
  name: aws-config
  annotations:
    eks.amazonaws.com/role-arn: arn:aws:iam::$AWS_ACCOUNT_ID:role/$IAM_ROLE_NAME
spec:
  podSecurityContext:
    fsGroup: 2000
---
apiVersion: pkg.crossplane.io/v1
kind: Provider
metadata:
  name: provider-aws
spec:
  package: crossplane/provider-aws:v${VERSION}
  controllerConfigRef:
    name: aws-config
EOF
```

4. Identify `provider-aws` service account

Make sure that the appropriate `ServiceAccount` exists:

```console
$ kubectl get serviceaccounts -n crossplane-system
```

If you used the install command above you should see a `ServiceAccount` in the
output named `provider-aws-*`. You should also see the `provider-aws-*`
controller `Pod` running if you execute `kubectl get pods -n crossplane-system`.
Set variables to match the name and namespace of this `ServiceAccount`:

```console
$ SERVICE_ACCOUNT_NAMESPACE=crossplane-system
$ SERVICE_ACCOUNT_NAME=$(kubectl get providers.pkg.crossplane.io provider-aws -o jsonpath="{.status.currentRevision}")
```

> The variable `$SERVICE_ACCOUNT_NAME` contains the default `ServiceAccount` name
> and changes with every provider release.

5. Create OIDC provider for cluster

*If you do not have `eksctl` installed you may use the [AWS Console](https://docs.aws.amazon.com/eks/latest/userguide/enable-iam-roles-for-service-accounts.html)*:

```console
$ eksctl utils associate-iam-oidc-provider \
    --cluster "${CLUSTER_NAME}" \
    --region "${AWS_REGION}" \
    --approve
```

6. Create IAM Role that `provider-aws` will use using `eksctl`

Create IAM role with trust relationship:

```console
$ eksctl create iamserviceaccount \
    --cluster "${CLUSTER_NAME}" \
    --region "${AWS_REGION}" \
    --name="${SERVICE_ACCOUNT_NAME}" \
    --namespace="${SERVICE_ACCOUNT_NAMESPACE}" \
    --role-name="${IAM_ROLE_NAME}" \
    --role-only \
    --attach-policy-arn="arn:aws:iam::aws:policy/AdministratorAccess" \
    --approve
```

6. Create IAM Role that `provider-aws` will use manually (skip if you created IAM
   Role using `eksctl`)

Set an variable that will be used in subsequent commands:

```console
$ OIDC_PROVIDER=$(aws eks describe-cluster --name "${CLUSTER_NAME}" --region "${AWS_REGION}" --query "cluster.identity.oidc.issuer" --output text | sed -e "s/^https:\/\///")
```

Create trust relationship for IAM role:

```console
$ cat > trust.yaml <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "Federated": "arn:aws:iam::${AWS_ACCOUNT_ID}:oidc-provider/${OIDC_PROVIDER}"
      },
      "Action": "sts:AssumeRoleWithWebIdentity",
      "Condition": {
        "StringLike": {
          "${OIDC_PROVIDER}:sub": "system:serviceaccount:${SERVICE_ACCOUNT_NAMESPACE}:provider-aws-*"
        }
      }
    }
  ]
}
EOF
```

> The default `ServiceAccount` name is the provider-aws revision and changes with every provider release.
> The conditional above wildcard matches the default `ServiceAccount` name in order to keep the role consistent across provider releases.

The above policy assumes a service account name of `provider-aws-*`.

Create an IAM role:

```console
$ aws iam create-role \
    --role-name "${IAM_ROLE_NAME}" \
    --assume-role-policy-document file://trust.json \
    --description "IAM role for provider-aws"
```

Associate a policy with the IAM role.
This example uses `AdministratorAccess`, but you should select a policy with
the minimum permissions required to provision your resources.

```console
$ aws iam attach-role-policy --role-name "${IAM_ROLE_NAME}" --policy-arn=arn:aws:iam::aws:policy/AdministratorAccess
```

1. Create `ProviderConfig`

Ensure that `ProviderConfig` resource kind was created:

```console
$ kubectl explain providerconfig --api-version='aws.crossplane.io/v1beta1'
```

To utilize those credentials to provision new resources, you must create a
`ProviderConfig` with `InjectedIdentity` in `.spec.credentials.source`:

```console
$ cat <<EOF | kubectl apply -f -
apiVersion: aws.crossplane.io/v1beta1
kind: ProviderConfig
metadata:
  name: aws-provider
spec:
  credentials:
    source: InjectedIdentity
EOF
```

You can now reference this `ProviderConfig` to provision any `provider-aws`
resources.

## Using kube2iam

This guide assumes that you already have:

- Created a policy with the minimum permissions required to provision your resources
- Created an IAM role that AWS Provider will assume to interact with AWS
- Associated the policy to the IAM role

Please refer to the previous section for details about these prerequisites.

Let's say the role you created is: `infra/k8s/crossplane`.

### Steps

1. Deploy a ControllerConfig

Crossplane provides a `ControllerConfig` type that allows you to customize the `Deployment` of a provider’s controller `Pod`.

A `ControllerConfig` can be created and referenced by any number of Provider objects that wish to use its configuration.

*Note: the kube2iam annotation must be under `spec.metadata.annotations` that will be added to the AWS provider `Pod`.*

```console
$ cat <<EOF | kubectl apply -f -
apiVersion: pkg.crossplane.io/v1alpha1
kind: ControllerConfig
metadata:
  name: aws-config
spec:
  metadata:
    annotations:
      # kube2iam annotation that will be added to the aws-provider defined in the next section
      iam.amazonaws.com/role: cdsf/k8s/kube2iam-crossplane-integration
  podSecurityContext:
    fsGroup: 2000
EOF
```
  
2. Deploy the `Provider` and a `ProviderConfig`

The AWS Provider is referencing the `ControllerConfig` we deployed in the previous step.
The `ProviderConfig` configures how AWS controllers will connect to AWS API.

```console
$ cat <<EOF | kubectl apply -f -
apiVersion: pkg.crossplane.io/v1
kind: Provider
metadata:
  name: aws-provider
spec:
  package: crossplane/provider-aws:${VERSION}
  controllerConfigRef:
    name: aws-config
---
apiVersion: aws.crossplane.io/v1beta1
kind: ProviderConfig
metadata:
  name: default
  spec:
  credentials:
    # Set source to 'InjectedIdentity' to be compliant with kube2iam behavior
    source: InjectedIdentity
EOF
```

*Note: Because the name of the `ProviderConfig` is `default` it will be used by any managed resources that do not explicitly reference a `ProviderConfig`.*

## Using `assumeRoleARN`

`provider-aws` will be configured to connect to AWS Account *A* via `InjectedIdentity`,
request temporary security credentials, and then `assumeRoleARN` to assume
a role in AWS Account *B* to manage the resources within AWS Account *B*.

The first thing that needs to be done is to create an IAM role within AWS Account *B* that `provider-aws` will `assumeRoleARN` into.

- From within the AWS console of AWS Account *B*, navigate to `IAM > Roles > Create role > Another AWS account`.
  - Enter the Account ID of Account *A* (the account `provider-aws` will call `assumeRoleARN` from).
  - (Optional) Check the box for `Require external ID`. This ensures requests
    coming from Account *A* can only use 'assumeRoleARN' if these requests pass the specified `externalID`.

Next, the `provider-aws` must be configured to use `assumeRoleARN`.
The code snippet below shows how to configure `provider-aws` to connect to
AWS Account *A* and `assumeRoleARN` into a role within AWS Account *B*.

```console
$ cat <<EOF | kubectl apply -f -
apiVersion: aws.crossplane.io/v1beta1
kind: ProviderConfig
metadata:
  name: account-b
spec:
  assumeRoleARN: "arn:aws:iam::999999999999:role/account-b"
  credentials:
    source: InjectedIdentity
EOF
```
