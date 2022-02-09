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
and then clean up the old CRs.

## Create using new CRDs

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

## Update Composition References

If you didn't use `Composition` to deploy these resources, you can skip this step and
continue with the clean up.

We need to edit each and every composite resource (XR) that references to the old ojects
to make them point to the new CRs we just created. In addition, we'll also need to
update `Composition` objects so that future creations use the new CRDs.

Since Crossplane is actively reconciling the composite resource and `Composition`, we
need to scale it down so that it doesn't create duplicate CRs while we're doing the
changes.

```bash
# Make sure to check the namespace.
kubectl -n crossplane-system scale deployment crossplane --replicas=0
```

We will first edit the `Composition` objects to use the new CRD `apiVersion` and `kind`s.
Find all `Composition`s to see which of them uses IAM resources:
```bash
kubectl get composition
```

In each element of the `spec.resources` array, make sure all IAM resources have the new
`apiVersion` and `kind`.

```yaml
apiVersion: apiextensions.crossplane.io/v1
kind: Composition
metadata:
  name: xpostgresqlinstances.aws.database.example.org
spec:
  ...
  resources:
    - name: user
      base:
        apiVersion: identity.aws.crossplane.io/v1alpha1 # <--- change this to the new group, iam.aws.crossplane.io/v1beta1
        kind: IAMUser # <--- change this to the new kind, i.e. "User" in this case.
        spec:
          ...
```


You will now need to edit composite resources that has composed resources with the old
CRDs. Get the list of all composite resources:
```bash
kubectl get composite
```

```yaml
# Assuming the composite type is XPostgreSQLInstance
kubectl edit XPostgreSQLInstance resource-wfmfp
```

The changes you need to make are as following:
```yaml
spec:
  ...
  resourceRefs:
  - apiVersion: identity.aws.crossplane.io/v1alpha1 # <--- change this to the new group, iam.aws.crossplane.io/v1beta1
    kind: IAMAccessKey # <--- change type to the new name, i.e. AccessKey in this case.
    name: platform-ref-aws-cluster-mwx8t-5j9hv # <--- make sure there is a resource with this name whose kind is AccessKey
  # Make the changes described above for every entry in this array.
  - apiVersion: identity.aws.crossplane.io/v1alpha1
    kind: IAMUser
    name: platform-ref-aws-cluster-mwx8t-klb7w
  - apiVersion: identity.aws.crossplane.io/v1beta1
    kind: IAMRole
    name: platform-ref-aws-cluster
  writeConnectionSecretToRef:
  ...
```

At this point, everything looks like they were created with these `apiVersion` and `kind`
in the first place. Before scaling Crossplane up again, let's check how many `identity` resources
exist in the cluster at this point so that we can be sure no duplicate is created
by composition controller:
```bash
kubectl get aws -o name | grep 'identity.aws.crossplane.io' | wc -l
```

We can bring back Crossplane:
```bash
kubectl -n crossplane-system scale deployment crossplane --replicas=1
```

After it's up and running check how many resources are there now and compare with the initial number:
```bash
kubectl get aws -o name | grep 'identity.aws.crossplane.io' | wc -l
```

All good if the numbers match!

## Clean up the old CRs

Now that the new ones took place of the old ones, we can delete the old ones:
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