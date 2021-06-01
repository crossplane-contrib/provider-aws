# Contributing a New Resource Using AWS Go Code Generation Pipeline

[AWS Go code generator](https://github.com/aws-controllers-k8s/code-generator)
is designed to generate code for Go controllers. Working with AWS community, we
have added support to generate Crossplane controllers. The generated code covers
mostly AWS API related parts such as `forProvider` fields in CRDs and conversion
functions between CRD and AWS Go SDK structs. The rest of the controller is very
similar to manually implemented ones; they use the same Crossplane managed
reconciler, CRDs follow the same patterns and all additional functionalities of
Crossplane are supported.

Most of the generated code is ready to go but there are places where we need
Crossplane-specific functionality, or some hacks to work around non-generic
behavior of the AWS API.

This guide shows how to get a new resource support up and running step by step.

## Code generation

AWS groups their resources under _services_ and the code generator works on
per-service basis. For example, `rds` is a service that contains `DBInstance`,
`DBCluster` and many other related resources. The first thing to do is to figure
out which service the resource we'd like to generate belongs to.

Take a look at the full list [here](https://github.com/aws/aws-sdk-go/tree/v1.37.4/models/apis)
and make note of the service name. For example, in order to generate Lambda
resources, you need to use `lambda` as service name. Once you figure that out,
the following is the command you need to run to generate CRDs and controllers:

```console
make services SERVICES=<service id>
```

The first run will take a while since it clones the code generator and AWS SDK.
Once it is completed, you'll see new folders in `apis` and `pkg/controller`
directories. By default, it generates every resource it can recognize but that's
not what you want if you'd like to add support for only a few resources. In order
to ignore some of the generated resources, you need to create a
`apis/<serviceid>/v1alpha1/generator-config.yaml` file with the following content:

```yaml
ignore:
  resource_names:
    - DBInstance
    - DBCluster
    - <other CRD kinds you want to ignore>
```

When you re-run the generation with this configuration, existing files won't be
deleted. So you may want to delete everything in `apis/<serviceid>/v1alpha1` except
`generator-config.yaml` and then re-run the command.

> If `apis/<serviceid>` is created from scratch, please add the new service name
> to the list called GENERATED_SERVICES in Makefile.

If this step fails for some reason, please raise an issue in [code-generator](https://github.com/aws-controllers-k8s/code-generator)
and mention that you're using Crossplane pipeline.

### Crossplane Code Generation

After ACK generation process completes, we need to run Crossplane generation procedure
which includes kubebuilder and crossplane-tools processes. However, currently
there are some missing structs in `apis/<serviceid>/v1alpha1` folder called
`Custom<CRDName>Parameters` which we'll use in the next steps. For now, create a
file called `apis/<serviceid>/v1alpha1/custom_types.go` with empty structs to get
the code to compile. For example, let's say you have a CRD whose kind is `Stage`.
The empty struct would look like following:
```golang
// CustomStageParameters includes the custom fields of Stage.
type CustomStageParameters struct{}
```

Now the code compiles, run the following:
```
make generate
```

> This command will first run kubebuilder generation tools which take care of
`zz_generated.deepcopy.go` files and CRD YAMLs. Then it will run the generators
in crossplane-tools that will create `zz_generated.managed.go` and `zz_generated.managedlist.go`
files that help CRD structs satisfy the Go interfaces we use in crossplane-runtime.

## Mandatory Custom Parts

There are a few things we need to implement manually since there is no support for
their generation yet.

### Setup Controller

The generated controller needs to be registered with the main controller manager
of the provider. Create a file called `pkg/controller/<serviceid>/setup.go` and
add the setup function like the following:
```golang
// SetupStage adds a controller that reconciles Stage.
func SetupStage(mgr ctrl.Manager, l logging.Logger, rl workqueue.RateLimiter) error {
	name := managed.ControllerName(svcapitypes.StageGroupKind)
	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(controller.Options{
			RateLimiter: ratelimiter.NewDefaultManagedRateLimiter(rl),
		}).
		For(&svcapitypes.Stage{}).
		Complete(managed.NewReconciler(mgr,
			resource.ManagedKind(svcapitypes.StageGroupVersionKind),
			managed.WithExternalConnecter(&connector{kube: mgr.GetClient()}),
			managed.WithLogger(l.WithValues("controller", name)),
			managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name)))))
}
```

Now you need to make sure this function is called in setup phase [here](https://github.com/crossplane/provider-aws/blob/483058c/pkg/controller/aws.go#L84).

#### Register CRD

If the group didn't exist before, we need to register its schema [here](https://github.com/crossplane/provider-aws/blob/master/apis/aws.go).

### Referencer Fields

In Crossplane, fields whose value can be taken from another CRD have two additional
fields called `*Ref` and `*Selector`. At the moment, ACK doesn't know about them
so we add them manually.

You'll see that `<CRDName>Parameters` struct has an inline field called `Custom<CRDName>Parameters`
whose struct we left empty earlier. We'll add the additional fields there.

Note that some fields are required and reference-able but we cannot mark them required
since when a reference is given, their value is resolved after the creation. In
such cases, we need to omit that field and put it under `Custom<CRDName>Parameters`
alongside with its referencer and selector fields. To ignore, you need to add the
following in `generator-config.yaml` like the following:

```yaml
ignore:
  field_paths:
    - CreateRouteInput.ApiId
    - DeleteRouteInput.ApiId
```

In this example, we don't want `Route` resource to have a required `ApiId` field.
But in order to skip it, we need to make sure to ignore it in every API call like
above. In the future, we hope to mark it once but that's how it works right now.

Note that any field you ignore needs to be handled in the hook functions you'll
define in the next steps.

### Reference Resolvers

We need to implement the resolver functions for every reference-able field. They
are mostly identical and you can see examples in the existing implementations like
[`Stage`](https://github.com/crossplane/provider-aws/blob/c269977/apis/apigatewayv2/v1alpha1/referencers.go#L30) resource.

### External Name

Crossplane has the notion of external name that we put under annotations. It corresponds
to the identifier field of the resource in AWS SDK structs. We try to stick with
name if available and unique enough but sometimes have to fall back to ARN. The
requirement for a field to be external-name is that it should be sufficient to
make all the calls we need together with required fields. In `Stage` case, it's
`StageName` field.

We need to [ignore `StageName`](https://github.com/crossplane/provider-aws/blob/7273d40/apis/apigatewayv2/v1alpha1/generator-config.yaml#L4)
in generation and then inject the value of our annotation in its place in the SDK calls.

Likely, we need to do this injection before every SDK call. The following is an
example for hook functions injected:
```golang
// SetupStage adds a controller that reconciles Stage.
func SetupStage(mgr ctrl.Manager, l logging.Logger, rl workqueue.RateLimiter) error {
	name := managed.ControllerName(svcapitypes.StageGroupKind)
	opts := []option{
		func(e *external) {
			e.preObserve = preObserve
			e.preCreate = preCreate
			e.preUpdate = preUpdate
			e.preDelete = preDelete
		},
	}
	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(controller.Options{
			RateLimiter: ratelimiter.NewDefaultManagedRateLimiter(rl),
		}).
		For(&svcapitypes.Stage{}).
		Complete(managed.NewReconciler(mgr,
			resource.ManagedKind(svcapitypes.StageGroupVersionKind),
			managed.WithExternalConnecter(&connector{kube: mgr.GetClient(), opts: opts}),
			managed.WithInitializers(managed.NewDefaultProviderConfig(mgr.GetClient())),
			managed.WithLogger(l.WithValues("controller", name)),
			managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name)))))
}

func preObserve(_ context.Context, cr *svcapitypes.Stage, obj *svcsdk.DescribeStageInput) error {
	obj.StageName = aws.String(meta.GetExternalName(cr))
	return nil
}

func preCreate(_ context.Context, cr *svcapitypes.Stage, obj *svcsdk.CreateStageInput) error {
	obj.StageName = aws.String(meta.GetExternalName(cr))
	return nil
}

func preUpdate(_ context.Context, cr *svcapitypes.Stage, obj *svcsdk.UpdateStageInput) error {
	obj.StageName = aws.String(meta.GetExternalName(cr))
	return nil
}

func preDelete(_ context.Context, cr *svcapitypes.Stage, obj *svcsdk.DeleteStageInput) error {
	obj.StageName = aws.String(meta.GetExternalName(cr))
	return nil
}
```

If the external-name is decided by AWS after the creation, then you
need to inject `postCreate` to set the crossplane resource external-name to 
the unique identifier of the resource, for eg see [`apigatewayv2`](https://github.com/crossplane/provider-aws/blob/master/pkg/controller/apigatewayv2/api/setup.go#L77)
You can discover what you can inject by inspecting `zz_controller.go` file.

### Readiness Check

Every managed resource needs to report its readiness. We'll do that in `postObserve`
call like the following:
```golang
func postObserve(_ context.Context, cr *svcapitypes.Backup, resp *svcsdk.DescribeBackupOutput, obs managed.ExternalObservation, err error) (managed.ExternalObservation, error) {
	if err != nil {
		return managed.ExternalObservation{}, err
	}
	switch aws.StringValue(resp.BackupDescription.BackupDetails.BackupStatus) {
	case string(svcapitypes.BackupStatus_SDK_AVAILABLE):
		cr.SetConditions(xpv1.Available())
	case string(svcapitypes.BackupStatus_SDK_CREATING):
		cr.SetConditions(xpv1.Creating())
	case string(svcapitypes.BackupStatus_SDK_DELETED):
		cr.SetConditions(xpv1.Unavailable())
	}
	return obs, nil
}
```

Some resources get ready right when you create them, like `Stage`. In such cases,
`postObserve` could be just like the following:
```golang
func (*external) postObserve(_ context.Context, cr *svcapitypes.Stage, _ *svcsdk.GetStageOutput, obs managed.ExternalObservation, err error) (managed.ExternalObservation, error) {
  if err != nil {
    return managed.ExternalObservation{}, err
  }
  cr.SetConditions(xpv1.Available())
  return obs, nil
}
```

## Custom Parts for Beta Quality

The earlier section is sufficient to get a controller into a working state. But
there are some functionalities missing like late-initialization and update. These
are not supported by ACK and even if there is support for update, it usually has
some caveats that cannot be made generic. So, we need to implement them manually.

### Late Initialization

Late initialization means that if a user doesn't have a desired value for a field
and provider chooses a default, we need to fetch that value from provider and
assign it to the corresponding field in `spec`. You can take a look at [`RDSInstance`](https://github.com/crossplane/provider-aws/blob/087f587/pkg/clients/rds/rds.go#L368)
example to see how it looks like but in most cases, we have much less fields so
it's not that long.

In order to override default no-op late initialization code, we'll use the same
mechanism as others:
```golang
func SetupStage(mgr ctrl.Manager, l logging.Logger, limiter workqueue.RateLimiter) error {
	name := managed.ControllerName(svcapitypes.StageGroupKind)
	opts := []option{
		func(e *external) {
			e.lateInitialize = lateInitialize
			...
		},
	}
	....
func lateInitialize(*svcapitypes.StageParameters, *svcsdk.DescribeStageOutput) error {
  return nil
}
```

### Update

#### IsUpToDate

Before issuing an update call, Crossplane needs to know whether resource is up-to-date
or not. In order to do that, we write a comparison function between observation
and our `spec`. Here is an [example](https://github.com/crossplane/provider-aws/blob/4318043/pkg/clients/eks/nodegroup.go#L166).

After writing the function, you can use it instead of the default `alwaysUpToDate`
similar to others:
```golang
func SetupStage(mgr ctrl.Manager, l logging.Logger, limiter workqueue.RateLimiter) error {
	name := managed.ControllerName(svcapitypes.StageGroupKind)
	opts := []option{
		func(e *external) {
			e.isUpToDate = isUpToDate
			...
		},
	}
	....
	func isUpToDate(*svcapitypes.Stage, *svcsdk.DescribeStageOutput) bool {
return true
}
```

#### Update

ACK has partial support for update calls. You need to inspect what's generated
and inject custom code wherever necessary. In a lot of cases, Update calls are tricky
and may have to be done by several calls. You can see an injected [example here](https://github.com/crossplane/provider-aws/blob/b65c7f9/pkg/controller/dynamodb/table/hooks.go#L278)
with custom logic to work around an API quirk.

## Testing

### Unit Tests

For the generated code, you don't need to write any tests. However, we encourage
you to write tests for all custom code blocks you have written.

### End-to-End Test

You need to thoroughly test the controller by using example YAMLs. Make sure
every functionality you implement works by checking with AWS console. Once all
looks good, add the example YAML you used to `examples/<serviceid>/<crd-name>.yaml`

#### Setup

You can use any Kubernetes cluster to deploy provider-aws. We'll use `kind` to
get a dev cluster:

```console
kind create cluster --wait 5m
```

You can test your code either by local setup, which allows better debugging experience,
or in-cluster setup which is the real world setup.

##### Local

We will deploy the CRDs so that controllers can bind their watchers:
```console
kubectl apply -f package/crds
```

Now you can run the controller from your IDE or terminal and it'll start reconciling:
```console
go run cmd/provider/main.go
```

> You need to set KUBECONFIG environment variable if it is not already set.

##### In-cluster

You can build the images and deploy it into the cluster using Crossplane Package Manager
to test the real-world scenario. There are two Docker images to be built; one has
the controller image, the other one has package metadata and CRDs. We need to choose a
controller image before building metadata image. Then we'll start the build process.

Pre-requisites: 
* Install Crossplane on the kind cluster following the steps [here](https://crossplane.io/docs/v1.1/reference/install.html).
* Install Crossplane CLI following the steps [here](https://crossplane.io/docs/v1.1/getting-started/install-configure.html#install-crossplane-cli). 
* Docker hub account with your own public repository.

Follow the steps below to get set-up for in-cluster testing:
* Run `make build DOCKER_REGISTRY=<username> VERSION=test-version` in `provider-aws`. This will create two images mentioned above:
  * `build-<short-sha>/provider-aws-amd64:latest //Controller Image This is the one that will be running in the cluster.` 
  * `build-<short-sha>/provider-aws-controller-amd64:latest //Metadata Image This is the one we'll use when we install the provider.`
  

* Tag and push both the images to your personal docker repository.

```console
docker image tag build-<short-sha>/provider-aws-controller-amd64:latest <username>/provider-aws-controller:test-version
docker push <username>/provider-aws-controller:test-version

docker image tag build-<short-sha>/provider-aws-amd64:latest <username>/provider-aws:test-version
docker push <username>/provider-aws:test-version
```

* Now we got the provider package pushed, and we can install it just like an official provider.
  Run:
```console
kubectl crossplane install provider <username>/provider-aws:test-version
```

#### Testing YAML

After either one of [local](#local) or [in-cluster](#in-cluster) is completed, you'll
get a cluster ready to start reconciling custom resources you create. First, we
need to create a default `ProviderConfig` so that managed resources can use it to
authenticate to AWS API. Use the following command to get the current AWS CLI
user credentials:
```console
BASE64ENCODED_AWS_ACCOUNT_CREDS=$(echo -e "[default]\naws_access_key_id = $(aws configure get aws_access_key_id --profile $aws_profile)\naws_secret_access_key = $(aws configure get aws_secret_access_key --profile $aws_profile)" | base64  | tr -d "\n")
```

Use it to create a `Secret` and `ProviderConfig`:
```console
cat > providerconfig.yaml <<EOF
---
apiVersion: v1
kind: Secret
metadata:
  name: aws-account-creds
  namespace: crossplane-system
type: Opaque
data:
  credentials: ${BASE64ENCODED_AWS_ACCOUNT_CREDS}
---
apiVersion: aws.crossplane.io/v1beta1
kind: ProviderConfig
metadata:
  name: default
spec:
  credentials:
    source: Secret
    secretRef:
      namespace: crossplane-system
      name: aws-account-creds
      key: credentials
EOF

# apply it to the cluster:
kubectl apply -f "providerconfig.yaml"

# delete the credentials variable
unset BASE64ENCODED_AWS_ACCOUNT_CREDS
```

You can create a managed resource now. Here is an example for a DynamoDB Table:
```yaml
apiVersion: dynamodb.aws.crossplane.io/v1alpha1
kind: Table
metadata:
  name: sample-table
spec:
  forProvider:
    region: us-east-1
    attributeDefinitions:
      - attributeName: attribute1
        attributeType: S
    keySchema:
      - attributeName: attribute1
        keyType: HASH
    provisionedThroughput:
      readCapacityUnits: 1
      writeCapacityUnits: 1
    streamSpecification:
      streamEnabled: true
      streamViewType: NEW_AND_OLD_IMAGES
```

#### Scenarios

[Create and Delete](#create-and-delete) section is sufficient if you aim for alpha
quality. If you want to aim for beta-level quality, you need to make sure all
scenarios work.

##### Create and Delete

Once you create the YAML, you should see the corresponding resource in AWS console.
Make sure you see the resource and properties match.

Then delete the CR in the cluster and check AWS console to see if it's gone.

##### Update

Make a change in the custom resource with `kubectl edit` and see if the resource
got the update by checking AWS console.

##### Late Initialization

After the creation, some fields are populated by AWS if they are not given. Go to
AWS console and compare the properties you see there with the custom resource in
the cluster. You need to see all properties in custom resource as well, especially
the ones you didn't initially specify.

##### Too Frequent Updates

If `IsUpToDate` check is giving false negatives, controller will make update calls
repeatedly and this causes problems. You can usually notice this easily but it's good
to check either via a breakpoint in your IDE or a print statement to see if `Update`
is run even when there is no change on the custom resource.
