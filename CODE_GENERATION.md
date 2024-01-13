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

## Git Submodules

First of all, you need to sync and update the submodules. This operation is required to `make` tasks work properly

```console
make submodules
```

## Code generation

AWS groups their resources under _services_ and the code generator works on
per-service basis. For example, `rds` is a service that contains `DBInstance`,
`DBCluster` and many other related resources. The first thing to do is to figure
out which service the resource we'd like to generate belongs to.

Take a look at the full list [here](https://github.com/aws/aws-sdk-go/tree/v1.37.10/models/apis)
and make note of the service name. For example, in order to generate Lambda
resources, you need to use `lambda` as service name. Once you figure that out,
the following is the commands you need to run to generate CRDs and controllers:

```console
touch apis/<service id>/generator-config.yaml
make services SERVICES=<service id>
```

The first run will take a while since it clones the code generator and AWS SDK.
Once it is completed, you'll see new folders in `apis` and `pkg/controller`
directories. By default, it generates every resource it can recognize but that's
not what you want if you'd like to add support for only a few resources. In order
to ignore some of the generated resources, you need to create an
`apis/<serviceid>/generator-config.yaml` file with the following content:

```yaml
ignore:
  resource_names:
    - DBInstance
    - DBCluster
    - <other CRD kinds you want to ignore>
```

If you'd like to specify an API version for a generated resource other than
the default `v1alpha1`, you can do so with a configuration in `generator-config.yaml`
as follows:
```yaml
resources:
  "Distribution":
    api_versions:
    - name: "v1beta1"
```
If not overridden in the `generator-config.yaml` file, the default version generated is
`v1alpha1`.

When you re-run the generation with this configuration, existing files won't be
deleted. So you may want to delete everything in `apis/<serviceid>/v1alpha1`and
then re-run the command.

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

### Setup Api Group

If a new group of api's has been introduced, we will need write a `Setup` function to satisfy
the interface which we will need to register the contollers. For example, if you added `globalaccelerator`
api group, we will need a file in `pkg/controller/globalaccelerator/setup.go with following contents:

```golang
package globalaccelerator

import (
	ctrl "sigs.k8s.io/controller-runtime"
	"github.com/crossplane/crossplane-runtime/pkg/controller"

	"github.com/crossplane-contrib/provider-aws/pkg/controller/globalaccelerator/accelerator"
	"github.com/crossplane-contrib/provider-aws/pkg/controller/globalaccelerator/listener"
	"github.com/crossplane-contrib/provider-aws/pkg/controller/globalaccelerator/endpointgroup"
	"github.com/crossplane-contrib/provider-aws/pkg/utils/setup"
)

// Setup athena controllers.
func Setup(mgr ctrl.Manager, o controller.Options) error {
	return setup.SetupControllers(
		mgr, o,
		accelerator.SetupAccelerator,
		listener.SetupListener,
		endpointgroup.SetupEndpointGroup,
	)
}
```

Now you need to make sure this function is called in setup phase [here](https://github.com/crossplane-contrib/provider-aws/blob/2e52b0a8b9ca9efa132e82549b9a48c13345dd27/pkg/controller/aws.go#L81).

### Setup Controller

The generated controller needs to be registered with the main controller manager
of the provider. Create a file called `pkg/controller/<serviceid>/<managedResource>/setup.go` and
add the setup function like the following:
```golang
// SetupStage adds a controller that reconciles Stage.
func SetupStage(mgr ctrl.Manager, o controller.Options) error {
	name := managed.ControllerName(svcapitypes.StageGroupKind)

	cps := []managed.ConnectionPublisher{managed.NewAPISecretPublisher(mgr.GetClient(), mgr.GetScheme())}
	if o.Features.Enabled(features.EnableAlphaExternalSecretStores) {
		cps = append(cps, connection.NewDetailsManager(mgr.GetClient(), v1alpha1.StoreConfigGroupVersionKind))
	}

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(o.ForControllerRuntime()).
		WithEventFilter(resource.DesiredStateChanged()).
		For(&svcapitypes.Stage{}).
		Complete(managed.NewReconciler(mgr,
			resource.ManagedKind(svcapitypes.StageGroupVersionKind),
			managed.WithExternalConnecter(&connector{kube: mgr.GetClient()}),
			managed.WithPollInterval(o.PollInterval),
			managed.WithLogger(o.Logger.WithValues("controller", name)),
			managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name))),
			managed.WithConnectionPublishers(cps...)))
}
```

#### Register CRD

If the group didn't exist before, we need to register its schema [here](https://github.com/crossplane/provider-aws/blob/master/apis/aws.go).

### Referencer Fields

In Crossplane, fields whose value can be taken from another CRD have two additional
fields called `*Ref` and `*Selector`. At the moment, there is no metadata in
either ACK or AWS SDK about these relations, so we will need to define them manually.

Let's say you're working on `Route` resource and there is a parameter called
`ApiId` in the CRD and its value can be taken from an `API` managed resource.
In order to define the relation, we first make ACK omit the field from the code
generation by adding the instances of SDK calls it exists in like the following
in `generator-config.yaml`:

```yaml
ignore:
  field_paths:
    - CreateRouteInput.ApiId
    - DeleteRouteInput.ApiId
```

Once you see that it doesn't exist under `RouteParameters` anymore, you can add
that field and its `ApiIdRef` and `ApiIdSelector` fields to `Custom<CRDName>Parameters`
struct we created in the earlier section. The addition will be similar to the
following:

```go
// APIID is actually required but since it's reference-able, it's not marked
// as required.
type CustomAPIMappingParameters struct {
  // ApiId is the ID for the API.
  // +immutable
  // +crossplane:generate:reference:type=API
  ApiId *string `json:"apiId,omitempty"`

  // ApiIdRef is a reference to an API used to set
  // the APIID.
  // +optional
  ApiIdRef *xpv1.Reference `json:"apiIdRef,omitempty"`

  // ApiIdSelector selects references to API used
  // to set the APIID.
  // +optional
  ApiIdSelector *xpv1.Selector `json:"apiIdSelector,omitempty"`
}
```

After you added the new field here, you need to handle usage of that field manually
in all hooks, like `preCreate`, `preObserve` etc., one by one since ACK won't
generate the assignment statements for it anymore.

Please note the line right before the definition of `ApiId` field that starts
with `+crossplane`. That's where we tell code generator that which kind this
field can reference to, and in this case it's `API` kind that lives under the
same package. If it was in another package like `ec2.VPC`, we could add the following
line to make it reference it:
```go
// +crossplane:generate:reference:type=github.com/crossplane/provider-aws/apis/ec2/v1alpha1.VPC
```

Once you're done, you can run `make generate`, the necessary reference resolvers
will be generated automatically.

In case you need to override some of the behavior like the name of the ref and
selector fields, or a custom function to fetch the value that is other than
external name of the referenced object, you can use the following comment markers:
```go
// +crossplane:generate:reference:type=API
// +crossplane:generate:reference:extractor=ApiARN()
// +crossplane:generate:reference:refFieldName=ApiIdRef
// +crossplane:generate:reference:selectorFieldName=ApiIdSelector
```

You can add the package prefix for `type` and `extractor` configurations if they
live in a different Go package.

Be aware that once you customize the extractor, you will need to implement it yourself.
An example can be found [here](https://github.com/crossplane-contrib/provider-aws/blob/72a6950/apis/lambda/v1beta1/referencers.go#L35).

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
func SetupStage(mgr ctrl.Manager, o controller.Options) error {
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
		WithOptions(o.ForControllerRuntime()).
		WithEventFilter(resource.DesiredStateChanged()).
		For(&svcapitypes.Stage{}).
		Complete(managed.NewReconciler(mgr,
			resource.ManagedKind(svcapitypes.StageGroupVersionKind),
			managed.WithExternalConnecter(&connector{kube: mgr.GetClient(), opts: opts}),
			managed.WithInitializers(),
			managed.WithPollInterval(o.PollInterval),
			managed.WithLogger(o.Logger.WithValues("controller", name)),
			managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name)))))
}

func preObserve(_ context.Context, cr *svcapitypes.Stage, obj *svcsdk.DescribeStageInput) error {
	obj.StageName = pointer.ToOrNilIfZeroValue(meta.GetExternalName(cr))
	return nil
}

func preCreate(_ context.Context, cr *svcapitypes.Stage, obj *svcsdk.CreateStageInput) error {
	obj.StageName = pointer.ToOrNilIfZeroValue(meta.GetExternalName(cr))
	return nil
}

func preUpdate(_ context.Context, cr *svcapitypes.Stage, obj *svcsdk.UpdateStageInput) error {
	obj.StageName = pointer.ToOrNilIfZeroValue(meta.GetExternalName(cr))
	return nil
}

func preDelete(_ context.Context, cr *svcapitypes.Stage, obj *svcsdk.DeleteStageInput) error {
	obj.StageName = pointer.ToOrNilIfZeroValue(meta.GetExternalName(cr))
	return nil
}
```

If the external-name is decided by AWS after the creation (like in most EC2
resources such as `vpc-id` of `VPC`), then you need to inject `postCreate` to
set the crossplane resource external-name to the unique identifier of the
resource, for eg see [`apigatewayv2`](https://github.com/crossplane/provider-aws/blob/72a6950/pkg/controller/apigatewayv2/api/setup.go#L85)
You can discover what you can inject by inspecting `zz_controller.go` file.

### Errors On Observe

In some situations aws api returns an error in cases where the resource does not exist.
It will be noticeable when the resource never gets created but you see an error from the api
which describes the object. You should see `return ok && awsErr.Code() == "ResourceNotFoundException"` in
the `IsNotFound`-function of `zz_conversion.go`.

To tell ack which error indicates that the resource is not present, add a similar config to `generator-config.yaml`:

```
resources:
  Table:
    fields:
      PointInTimeRecoveryEnabled:
        from:
          operation: UpdateContinuousBackups
          path: PointInTimeRecoverySpecification.PointInTimeRecoveryEnabled
    exceptions:
      errors:
        404:
          code: ResourceNotFoundException
```


### Readiness Check

Every managed resource needs to report its readiness. We'll do that in `postObserve`
call like the following:

```golang
func postObserve(_ context.Context, cr *svcapitypes.Backup, resp *svcsdk.DescribeBackupOutput, obs managed.ExternalObservation, err error) (managed.ExternalObservation, error) {
	if err != nil {
		return managed.ExternalObservation{}, err
	}
	switch pointer.StringValue(resp.BackupDescription.BackupDetails.BackupStatus) {
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

You can build the image and deploy it into the cluster using Crossplane Package Manager
to test the real-world scenario.

Pre-requisites:
* Install Crossplane on the kind cluster following the steps [here](https://crossplane.io/docs/master/reference/install.html).
* Install Crossplane CLI following the steps [here](https://crossplane.io/docs/master/getting-started/install-configure.html#install-crossplane-cli).
* Docker hub account with your own public repository.

Follow the steps below to get set-up for in-cluster testing:
* Run `make build VERSION=test-version` in `provider-aws`. This will create two artifacts
  * `build-<short-sha>/provider-aws-amd64:latest` Controller image
  * `_output/xpkg/linux_amd64/provider-aws-foo.xpkg` OCI image bundle with the controller and package.yaml


* Tag and push both the images to your personal docker repository:

```console
make publish.artifacts PLATFORMS=linux_amd64 XPKG_REG_ORGS=index.docker.io/<username> VERSION=test-version
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
  name: example
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
