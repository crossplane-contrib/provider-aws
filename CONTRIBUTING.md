# Contributing to Provider AWS

See Crossplane contributing doc for guidelines that apply to all Crossplane
repositories.

## Upgrading a CRD API Version

API version of a CRD provides some guarantees about its stability and amount of
interventions you can expect during upgrades. Crossplane strictly follows the
conventions Kubernetes sets forth [here](https://kubernetes.io/docs/reference/using-api/#api-versioning).

The actions needed to be taken depends on whether the upgrade includes a breaking
schema change and from what level it is upgrading from and to, i.e. `alpha`/`beta`
etc.

#### Upgrade from `alpha` to `beta`

From `alpha` to `beta`, breaking changes can be made without early notice. However,
we do include a warning in release notes and provide instructions on how to migrate
from one to the other manually.

Step to upgrade from `v1alpha1` to `v1beta1`:
* Create a new folder called `v1beta1` in `apis/<group name>` and copy the following
  files from existing `v1alpha1` folder:
  * `register.go`: Replace all `v1alpha1` occurrences with `v1beta1`.
* Copy the Go type into `v1beta1` package.
  * Add `// +kubebuilder:storageversion` comment to the type comments.
* Delete all files in `v1alpha1` and `v1beta1` packages starting with `zz_` and
  run `make generate` so that functions that satisfy certain interfaces are generated
  for the new types.
  * If there are functions in those packages assuming the structs implement those
    interfaces, such as old reference resolvers, you'll need to comment them out
    for code generation to succeed and bring them back once it's completed.
* Add the new API package to `apis/aws.go` type registration list.
* Make sure the controller in `pkg/controller/<group name>/<kind name>/controller.go`
  uses the new struct from `v1beta1` package.
* Make sure the client functions in `pkg/clients/<group name>/<kind name>.go`
  uses the new struct from `v1beta1` package.
* The old Go type should not be used in anywhere, like controller, or reference
  resolver of another type. Its sole usage should be registration in `apis/aws.go`
* Deprecate the old type in `v1alpha1` package by adding the following comments
  to both `<Kind>` and `<Kind>List` structs with a little bit of detail about what
  they should do:
  * `// +kubebuilder:deprecatedversion:warning="Please use v1beta1 version of this resource that has identical schema."`
  * `// Deprecated: Please use v1beta1 version of this resource.`
* Run `make reviewable` and see if the corresponding CRD is updated in `package/crds`
  folder.
* Update the example in `examples/<group name>/<kind>.yaml` to have `v1beta1`.