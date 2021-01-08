# Contributing New Resource Using ACK

Provider AWS uses ACK code generation pipeline to generate the boilerplate for
CRDs and their controllers. Most of the generated code is ready to go but there
are places where we need Crossplane-specific functionality or some hacks to work
around non-generic behavior of the AWS API.

This guide shows how to get a new resource support up and running step by step.

## Code generation

This is the very first step. You need to clone ACK repository and run the following
command in the root directory:

```console
go -tags generate run cmd/ack-generate/main.go crossplane <ServiceID> --provider-dir <provider-aws directory>
```

`ServiceID` list is found here: https://github.com/aws/aws-sdk-go/tree/master/models/apis
For example, in order to generate Lambda resources, you need to use `lambda` as
Service ID.

Once this is completed, you'll see new folders in `apis` and `pkg/controller` directories.
By default, it generates every resource it can recognize but that's not what you
want if you'd like to add support for only a few resource. In order to ignore some
of the generated resources, you need to create a `apis/<serviceid>/v1alpha1/generator-config.yaml`
file with the following content:
```yaml
ignore:
  resource_names:
    - <CRD Name you want to ignore>
```

When you re-run the generation with this configuration, existing files won't be
deleted. So you may want to delete everything in `apis/<serviceid>/v1alpha1` except
`generator-config.yaml` and then re-run the command.

If the generation fails for any reason, please open an issue.

### Crossplane Code Generation

After ACK generation process completes, we need to run Crossplane generation procedure
which includes kubebuilder and crossplane-tools processes. However, currently
there are some missing structs in `apis/<serviceid>/v1alpha1` folder called
`Custom<CRDName>Parameters` which we'll use in the next steps. For now, create a
file called `apis/<serviceid>/v1alpha1/custom_types.go` with empty structs to get
the code to compile like the following:
```golang
// CustomStageParameters includes the custom fields of Stage.
type CustomStageParameters struct{}
```

Now the code compiles, run the following:
```
make generate
```

## Mandatory Custom Parts

There are a few things we need to implement manually since there is no support for
their generation yet.

### Setup Controller

The generated controller needs to be registered with the main controller manager.
Create a file called `pkg/controller/<serviceid>/setup.go` and add the setup function
like the following:
```golang
// SetupStage adds a controller that reconciles Stage.
func SetupStage(mgr ctrl.Manager, l logging.Logger) error {
	name := managed.ControllerName(svcapitypes.StageGroupKind)
	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
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

If the group didn't exist before, we need to register its schema [here](https://github.com/crossplane/provider-aws/blob/483058c/apis/aws.go#L77).

### Referencer Fields

In Crossplane, fields whose value can be taken from another CRD have two additional
fields called `*Ref` and `*Selector`. At the moment, ACK doesn't know about them
so we add them manually.

You'll see that `<CRDName>Parameters` struct has an inline field called `Custom<CRDName>Parameters`
whose struct we left empty earlier. We'll add the additional fields there.

Note that some fields are required and reference-able but we cannot mark them required
since when a reference is given, their value is resolver after the creation. In
such cases, we need to omit that field and put it under `Custom<CRDName>Parameters`
alongisde with its referencer and selector fields. To ignore, you need to add the
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
to the identifier field of the resource. We try to stick with name if available
and unique enough but sometimes have to fall back to ARN. The requirement for a
field to be external-name is that it should be sufficient to make all the calls
we need together with required fields. In `Stage` case, it's `StageName` field.

We need to [ignore `StageName`](https://github.com/crossplane/provider-aws/blob/7273d40/apis/apigatewayv2/v1alpha1/generator-config.yaml#L4)
in generation and then inject the value of our annotation in its place in the SDK calls.

Likely, we need to do this injection before every SDK call. The following is an
example for hook functions injected:
```golang
// SetupStage adds a controller that reconciles Stage.
func SetupStage(mgr ctrl.Manager, l logging.Logger) error {
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

If the external-name is decided by AWS after the creation, like in VPC, then you
need to inject `postCreate` and make sure to `meta.SetExternalName(cr, id)` there
and return `managed.ExternalCreation{ExternalNameAssigned: true}`.

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

Late initialization means that if user doesn't have a desired value for a field
and provider chooses a default, we need to fetch that value from provider and
assign it to the corresponding field in `spec`. You can take a look at [`RDSInstance`](https://github.com/crossplane/provider-aws/blob/087f587/pkg/clients/rds/rds.go#L368)
example to see how it looks like but in most cases, we have much less fields so
it's not that long.

In order to override default no-op late initialization code, we'll use the same
mechanism as others:
```golang
func SetupStage(mgr ctrl.Manager, l logging.Logger) error {
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
func SetupStage(mgr ctrl.Manager, l logging.Logger) error {
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
and may have to be done by several calls. You can see an injected [example here](https://github.com/crossplane/provider-aws/pull/485/files#diff-8a38ae8bcdbb080fbc35feb71701c1ada99de0c36c474d42899d1ce628cf3927R277)
with custom logic to work around an API quirk.

## Testing

You need to thoroughly test the controller by using example YAMLs. Make sure
every functionality you implement works by checking with AWS console. Once all looks
good, add the example YAML you used to `examples/<serviceid>/<crd-name>.yaml`