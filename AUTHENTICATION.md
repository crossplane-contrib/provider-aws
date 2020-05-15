# Authenticating to AWS API

`provider-aws` requires credentials to be provided in order to authenticate to the
AWS API. This can be done in one of two ways:

1. Base64 encoding static credentials in a Kubernetes `Secret`. This is
   described in detail
   [here](https://crossplane.io/docs/v0.8/cloud-providers/aws/aws-provider.html).
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

3. Create OIDC provider for cluster

*If you do not have `eksctl` installed you may use the [AWS
Console](https://docs.aws.amazon.com/eks/latest/userguide/enable-iam-roles-for-service-accounts.html)*

```
eksctl utils associate-iam-oidc-provider --cluster <cluster-name> --region <region> --approve
```

4. Create IAM Role that provider-aws will use

Set environment variables that will be used in subsequent commands:

```
AWS_ACCOUNT_ID=$(aws2 sts get-caller-identity --query "Account" --output text)
OIDC_PROVIDER=$(aws2 eks describe-cluster --name <cluster-name> --region <region> --query "cluster.identity.oidc.issuer" --output text | sed -e "s/^https:\/\///")
# namespace for provider-aws should match namespace of your ClusterPackageInstall
SERVICE_ACCOUNT_NAMESPACE=crossplane-system
# service account name for provider-aws should match name of your ClusterPackageInstall
SERVICE_ACCOUNT_NAME=provider-aws
IAM_ROLE_NAME=provider-aws # name for IAM role, can be anything you want
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

5. Install Crossplane and provider-aws

Install Crossplane from `alpha` channel:

```
kubectl create namespace crossplane-system
helm repo add crossplane-alpha https://charts.crossplane.io/alpha

helm install crossplane --namespace crossplane-system crossplane-alpha/crossplane
```

Install `provider-aws`:

```yaml
apiVersion: packages.crossplane.io/v1alpha1
kind: ClusterPackageInstall
metadata:
  name: provider-aws # crossplane will create service account with this name
  namespace: crossplane-system # service account will be created in this namespace
spec:
  package: "crossplane/provider-aws:v0.6.0"
```

6. Enable IAM Role for provider-aws Service Account

First, check to make sure that the appropriate `ServiceAccount` exists:

```
kubectl get serviceaccounts -n crossplane-system
```

If you used the `ClusterPackageInstall` above you should see a `ServiceAccount` in
the output named `provider-aws`. You should also see the `provider-aws` controller
`Pod` running if you execute `kubectl get pods -n crossplane-system`. To inject
the credential information into the Pod, we must annotate its `ServiceAccount`
with the desired IAM role:

```
kubectl annotate serviceaccount -n $SERVICE_ACCOUNT_NAMESPACE $SERVICE_ACCOUNT_NAME \
eks.amazonaws.com/role-arn=arn:aws:iam::$AWS_ACCOUNT_ID\:role/$IAM_ROLE_NAME
```

Because the credentials get injected when the `Pod` starts, we need to restart
it. This can be done by deleting the `Pod` and letting its `Deployment` recreate
it:

```
kubectl delete pod <pod-name> -n crossplane-system
```

You should immediately see another `Pod` be created for the `provider-aws`
controller. It will have access to credentials used to assume the IAM role.

7. Create `Provider`

To utilize those credentials to provision new resources, you must create a
`Provider` with `useServiceAccount: true`:

```
cat > aws-provider.yaml <<EOF
apiVersion: aws.crossplane.io/v1alpha3
kind: Provider
metadata:
  name: aws-provider
spec:
  useServiceAccount: true # this tells crossplane to look for credentials injected from service account
  region: us-west-2
EOF

kubectl apply -f aws-provider.yaml
```

You can now reference this `Provider` to provision any `provider-aws` resources.
