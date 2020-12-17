# Authenticating to AWS API

`provider-aws` requires credentials to be provided in order to authenticate to the
AWS API. This can be done in one of two ways:

1. Base64 encoding static credentials in a Kubernetes `Secret`. This is
   described in detail
   [here](https://crossplane.io/docs/v0.13/getting-started/install-configure.html#select-provider).
2. Authenticating using IAM Roles for Kubernetes `ServiceAccounts`. This
   functionality is only available when running Crossplane on EKS, and the
   feature has been
   [enabled](https://docs.aws.amazon.com/eks/latest/userguide/iam-roles-for-service-accounts.html)
   in the cluster.

Using IAM Roles for Service Accounts requires some additional setup for the
time-being. The steps for enabling are described below. Many of the steps can
also be found in the [AWS
docs](https://docs.aws.amazon.com/eks/latest/userguide/enable-iam-roles-for-service-accounts.html).

## Steps

These steps assume you already have a running EKS cluster with a sufficiently
large node pool.

1. Connect to your EKS cluster

```
aws eks --region <region> update-kubeconfig --name <cluster-name>
```

2. Get AWS account information

Get AWS account information and pick an IAM role name. These will be used to
setup an OIDC provider and inject credentials into the provider-aws controller
pod.

```
export AWS_ACCOUNT_ID=$(aws sts get-caller-identity --query "Account" --output text)
export IAM_ROLE_NAME=provider-aws # name for IAM role, can be anything you want
```

3. Install Crossplane

Install Crossplane from `alpha` channel:

```
kubectl create namespace crossplane-system
helm repo add crossplane-alpha https://charts.crossplane.io/alpha

helm install crossplane --namespace crossplane-system crossplane-alpha/crossplane
```

`provider-aws` can be installed with the [Crossplane
CLI](https://crossplane.io/docs/v0.13/getting-started/install-configure.html#install-crossplane-cli),
but we will do so manually so that we can also create and reference a
`ControllerConfig`:

```
cat > provider-config.yaml <<EOF
apiVersion: pkg.crossplane.io/v1alpha1
kind: ControllerConfig
metadata:
  name: aws-config
  annotations:
    eks.amazonaws.com/role-arn: arn:aws:iam::$AWS_ACCOUNT_ID\:role/$IAM_ROLE_NAME
spec:
  podSecurityContext:
    fsGroup: 2000
---
apiVersion: pkg.crossplane.io/v1alpha1
kind: Provider
metadata:
  name: provider-aws
spec:
  package: crossplane/provider-aws:alpha
  controllerConfigRef:
    name: aws-config
EOF

kubectl apply -f provider-config.yaml
```

4. Identify provider-aws service account

Make sure that the appropriate `ServiceAccount` exists:

```
kubectl get serviceaccounts -n crossplane-system
```

If you used the install command above you should see a `ServiceAccount` in the
output named `provider-aws-*`. You should also see the `provider-aws-*`
controller `Pod` running if you execute `kubectl get pods -n crossplane-system`.
Set environment variables to match the name and namespace of this
`ServiceAccount`:

```
SERVICE_ACCOUNT_NAMESPACE=crossplane-system
SERVICE_ACCOUNT_NAME=provider-aws-<YOUR-SERVICE-ACCOUNT-EXTENSION>
```

5. Create OIDC provider for cluster

*If you do not have `eksctl` installed you may use the [AWS
Console](https://docs.aws.amazon.com/eks/latest/userguide/enable-iam-roles-for-service-accounts.html)*

```
eksctl utils associate-iam-oidc-provider --cluster <cluster-name> --region <region> --approve
```

6. Create IAM Role that provider-aws will use

Set environment variables that will be used in subsequent commands:

```
OIDC_PROVIDER=$(aws eks describe-cluster --name <cluster-name> --region <region> --query "cluster.identity.oidc.issuer" --output text | sed -e "s/^https:\/\///")
```

Create trust relationship for IAM role:

```
read -r -d '' TRUST_RELATIONSHIP <<EOF
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
        "StringEquals": {
          "${OIDC_PROVIDER}:sub": "system:serviceaccount:${SERVICE_ACCOUNT_NAMESPACE}:${SERVICE_ACCOUNT_NAME}"
        }
      }
    }
  ]
}
EOF
echo "${TRUST_RELATIONSHIP}" > trust.json
```

Create IAM role:

```
aws iam create-role --role-name $IAM_ROLE_NAME --assume-role-policy-document file://trust.json --description "IAM role for provider-aws"
```

Associate a policy with the IAM role. This example uses `AdministratorAccess`,
but you should select a policy with the minimum permissions required to
provision your resources.

```
aws iam attach-role-policy --role-name $IAM_ROLE_NAME --policy-arn=arn:aws:iam::aws:policy/AdministratorAccess
```

7. Create `ProviderConfig`

To utilize those credentials to provision new resources, you must create a
`ProviderConfig` with `source: InjectedIdentity`:

```
cat > provider-config.yaml <<EOF
apiVersion: aws.crossplane.io/v1beta1
kind: ProviderConfig
metadata:
  name: aws-provider
spec:
  credentials:
    source: InjectedIdentity
EOF

kubectl apply -f provider-config.yaml
```

You can now reference this `ProviderConfig` to provision any `provider-aws`
resources.
