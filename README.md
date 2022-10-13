# provider-aws

## Overview

This `provider-aws` repository is the Crossplane infrastructure provider for
[Amazon Web Services (AWS)](https://aws.amazon.com). The provider that is built
from the source code in this repository can be installed into a Crossplane
control plane and adds the following new functionality:

* Custom Resource Definitions (CRDs) that model AWS infrastructure and services
  (e.g. [Amazon Relational Database Service (RDS)](https://aws.amazon.com/rds/),
  [EKS clusters](https://aws.amazon.com/eks/), etc.)
* Controllers to provision these resources in AWS based on the users desired
  state captured in CRDs they create
* Implementations of Crossplane's portable resource abstractions, enabling AWS
  resources to fulfill a user's general need for cloud services

## Getting Started and Documentation

For getting started guides, installation, deployment, and administration, see
our [Documentation](https://crossplane.io/docs).

## Contributing

provider-aws is a community driven project and we welcome contributions. See the
Crossplane
[Contributing](https://github.com/crossplane/crossplane/blob/master/CONTRIBUTING.md)
guidelines to get started.

### Adding New Resource

We use AWS Go code generation pipeline to generate new controllers. See [Code Generation Guide](CODE_GENERATION.md)
to add a new resource.

## Releases

AWS Provider is released every 4 weeks and we issue patch releases as necessary.
For example, `v0.20.0` is released on October 19, 2021. The next minor
release `v0.21.0` will be cut on November 16, 2021, and so on.

## Report a Bug

For filing bugs, suggesting improvements, or requesting new features, please
open an [issue](https://github.com/crossplane/provider-aws/issues).

## Contact

Please use the following to reach members of the community:

* Slack: Join our [slack channel](https://slack.crossplane.io)
* Forums:
  [crossplane-dev](https://groups.google.com/forum/#!forum/crossplane-dev)
* Twitter: [@crossplane_io](https://twitter.com/crossplane_io)
* Email: [info@crossplane.io](mailto:info@crossplane.io)

## Roadmap

provider-aws goals and milestones are currently tracked in the Crossplane
repository. More information can be found in
[ROADMAP.md](https://github.com/crossplane/crossplane/blob/master/ROADMAP.md).

## Governance and Owners

provider-aws is run according to the same
[Governance](https://github.com/crossplane/crossplane/blob/master/GOVERNANCE.md)
and [Ownership](https://github.com/crossplane/crossplane/blob/master/OWNERS.md)
structure as the core Crossplane project.

## Code of Conduct

provider-aws adheres to the same [Code of
Conduct](https://github.com/crossplane/crossplane/blob/master/CODE_OF_CONDUCT.md)
as the core Crossplane project.

## Licensing

provider-aws is under the Apache 2.0 license.

[![FOSSA Status](https://app.fossa.io/api/projects/git%2Bgithub.com%2Fcrossplane%2Fprovider-aws.svg?type=large)](https://app.fossa.io/projects/git%2Bgithub.com%2Fcrossplane%2Fprovider-aws?ref=badge_large)