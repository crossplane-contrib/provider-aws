# Upgrade from v0.21.x to v0.22.x

There are number of resources in `v0.22.0` that were upgraded to `v1beta1`. Most
of those upgrades require no manual intervention since their schemas didn't
change. However, IAM resources were used to be represented in `identity.aws.crossplane.io`
but now they are moved to `iam.aws.crossplane.io`, their versions got upgraded
to `v1beta1` and `IAM` prefix is dropped from their name.

Thanks to package manager, old CRDs are not deleted once you upgrade to the new
version; only their controllers are stopped, which gives users time to create
corresponding new types.

The first step is to upgrade to the new version:
```bash
# Find the name of your Provider object.
kubectl get provider.pkg
# Patch it with the new version. Make sure you use the latest patch release of
# v0.22.x
kubectl patch provider.pkg <provider object name> -p '{"spec":{"package": "crossplane/provider-aws:v0.22.0"}}' --type=merge
```

At this point, there is no controller reconciling your existing IAM custom resources.
We will create new CRs that correspond to those old IAM CRs but have new type,
and then clean up the old CRs

## Using Composition

< TO BE FILLED >

```bash
# The number of old CRs in identity group.
kubectl get aws -o name | grep 'identity.aws.crossplane.io' | wc -l
# The number of new CRs in iam group. This number should be equal to the number
# of old CRs from identity group.
kubectl get aws -o name | grep 'iam.aws.crossplane.io' | wc -l
```

Now, we can delete old CRs:
```bash
kubectl get aws -o name \
  | grep 'identity.aws.crossplane.io' \
  | xargs kubectl delete
# Since no controller is reconciling them, we need to remove the finalizer to
# unblock deletion.
kubectl get aws -o name \
  | grep 'identity.aws.crossplane.io' \
  | xargs -I {} kubectl patch {} -p '{"metadata":{"finalizers": []}}' --type=merge
```

Done!

## Without Composition

If you don't use composition to deploy the managed resources, then you'll need
to do the re-creation part manually.

Create new custom resources (CR) with the new metadata:
```bash
# IAMUser
kubectl get iamuser.identity.aws.crossplane.io -ojson \
  | jq 'del(.items[].metadata.namespace,.items[].metadata.resourceVersion,.items[].metadata.uid,.items[].metadata.generation) | .items[].metadata.creationTimestamp=null' \
  | sed 's|identity.aws.crossplane.io/v1alpha1|iam.aws.crossplane.io/v1beta1|g' \
  | sed 's|IAMUser|User|g' \
  | kubectl apply -f -
  
# IAMAccessKey
kubectl get iamaccesskey.identity.aws.crossplane.io -ojson \
  | jq 'del(.items[].metadata.namespace,.items[].metadata.resourceVersion,.items[].metadata.uid,.items[].metadata.generation) | .items[].metadata.creationTimestamp=null' \
  | sed 's|identity.aws.crossplane.io/v1alpha1|iam.aws.crossplane.io/v1beta1|g' \
  | sed 's|IAMAccessKey|AccessKey|g' \
  | kubectl apply -f -
  
# IAMGroup
kubectl get iamgroup.identity.aws.crossplane.io -ojson \
  | jq 'del(.items[].metadata.namespace,.items[].metadata.resourceVersion,.items[].metadata.uid,.items[].metadata.generation) | .items[].metadata.creationTimestamp=null' \
  | sed 's|identity.aws.crossplane.io/v1alpha1|iam.aws.crossplane.io/v1beta1|g' \
  | sed 's|IAMGroup|Group|g' \
  | kubectl apply -f -
  
# IAMGroupPolicyAttachment
kubectl get iamgrouppolicyattachment.identity.aws.crossplane.io -ojson \
  | jq 'del(.items[].metadata.namespace,.items[].metadata.resourceVersion,.items[].metadata.uid,.items[].metadata.generation) | .items[].metadata.creationTimestamp=null' \
  | sed 's|identity.aws.crossplane.io/v1alpha1|iam.aws.crossplane.io/v1beta1|g' \
  | sed 's|IAMGroupPolicyAttachment|GroupPolicyAttachment|g' \
  | kubectl apply -f -
  
# IAMGroupUserMembership
kubectl get iamgroupusermembership.identity.aws.crossplane.io -ojson \
  | jq 'del(.items[].metadata.namespace,.items[].metadata.resourceVersion,.items[].metadata.uid,.items[].metadata.generation) | .items[].metadata.creationTimestamp=null' \
  | sed 's|identity.aws.crossplane.io/v1alpha1|iam.aws.crossplane.io/v1beta1|g' \
  | sed 's|IAMGroupUserMembership|GroupUserMembership|g' \
  | kubectl apply -f -
  
# OpenIDConnectProvider
kubectl get openidconnectprovider.identity.aws.crossplane.io -ojson \
  | jq 'del(.items[].metadata.namespace,.items[].metadata.resourceVersion,.items[].metadata.uid,.items[].metadata.generation) | .items[].metadata.creationTimestamp=null' \
  | sed 's|identity.aws.crossplane.io/v1alpha1|iam.aws.crossplane.io/v1beta1|g' \
  | kubectl apply -f -
  
# IAMPolicy
kubectl get iampolicy.identity.aws.crossplane.io -ojson \
  | jq 'del(.items[].metadata.namespace,.items[].metadata.resourceVersion,.items[].metadata.uid,.items[].metadata.generation) | .items[].metadata.creationTimestamp=null' \
  | sed 's|identity.aws.crossplane.io/v1alpha1|iam.aws.crossplane.io/v1beta1|g' \
  | sed 's|IAMPolicy|Policy|g' \
  | kubectl apply -f -
  
# IAMUserPolicyAttachment
kubectl get iamuserpolicyattachment.identity.aws.crossplane.io -ojson \
  | jq 'del(.items[].metadata.namespace,.items[].metadata.resourceVersion,.items[].metadata.uid,.items[].metadata.generation) | .items[].metadata.creationTimestamp=null' \
  | sed 's|identity.aws.crossplane.io/v1alpha1|iam.aws.crossplane.io/v1beta1|g' \
  | sed 's|IAMUserPolicyAttachment|UserPolicyAttachment|g' \
  | kubectl apply -f -

# IAMRole
kubectl get iamrole.identity.aws.crossplane.io -ojson \
  | jq 'del(.items[].metadata.namespace,.items[].metadata.resourceVersion,.items[].metadata.uid,.items[].metadata.generation) | .items[].metadata.creationTimestamp=null' \
  | sed 's|identity.aws.crossplane.io/v1beta1|iam.aws.crossplane.io/v1beta1|g' \
  | sed 's|IAMRole|Role|g' \
  | kubectl apply -f -

# IAMRolePolicyAttachment
kubectl get iamrolepolicyattachment.identity.aws.crossplane.io -ojson \
  | jq 'del(.items[].metadata.namespace,.items[].metadata.resourceVersion,.items[].metadata.uid,.items[].metadata.generation) | .items[].metadata.creationTimestamp=null' \
  | sed 's|identity.aws.crossplane.io/v1beta1|iam.aws.crossplane.io/v1beta1|g' \
  | sed 's|IAMRolePolicyAttachment|RolePolicyAttachment|g' \
  | kubectl apply -f -
```

Check whether we have created a counterpart for all CR instances.
```bash
# The number of old CRs in identity group.
kubectl get aws -o name | grep 'identity.aws.crossplane.io' | wc -l
# The number of new CRs in iam group. This number should be equal to the number
# of old CRs from identity group.
kubectl get aws -o name | grep 'iam.aws.crossplane.io' | wc -l
```

Now, we can delete old CRs:
```bash
# Since no controller is reconciling them, we need to remove the finalizer to
# unblock deletion.
kubectl get aws -o name \
  | grep 'identity.aws.crossplane.io' \
  | xargs -I {} kubectl patch {} -p '{"metadata":{"finalizers": []}}' --type=merge
# Delete all resources in identity group
kubectl get aws -o name \
  | grep 'identity.aws.crossplane.io' \
  | xargs kubectl delete
```

You can check whether there is any remaining CRs from the old types:
```bash
kubectl get aws -o name | grep 'identity.aws.crossplane.io'
```

Done! You can check all resources by running `kubectl get managed`.