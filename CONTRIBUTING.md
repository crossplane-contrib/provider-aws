## How to Contribute

The Crossplane project is under [Apache 2.0 license](LICENSE). We accept contributions via
GitHub pull requests. This document outlines some of the conventions related to
development workflow, commit message formatting, contact points and other
resources to make it easier to get your contribution accepted.

## Certificate of Origin

By contributing to this project you agree to the Developer Certificate of
Origin (DCO). This document was created by the Linux Kernel community and is a
simple statement that you, as a contributor, have the legal right to make the
contribution. See the [DCO](DCO) file for details.

Contributors sign-off that they adhere to these requirements by adding a
Signed-off-by line to commit messages. For example:

```
This is my commit message

Signed-off-by: Random J Developer <random@developer.example.org>
```

Git even has a -s command line option to append this automatically to your
commit message:

```
git commit -s -m 'This is my commit message'
```

If you have already made a commit and forgot to include the sign-off, you can amend your last commit
to add the sign-off with the following command, which can then be force pushed.

```
git commit --amend -s
```

We use a [DCO bot](https://github.com/apps/dco) to enforce the DCO on each pull
request and branch commits.

## Getting Started

- Fork the repository on GitHub
- Read the [install](INSTALL.md) for build and test instructions
- Play with the project, submit bugs, submit patches!

## Contribution Flow

This is a rough outline of what a contributor's workflow looks like:

- Create a branch from where you want to base your work (usually master).
- Make your changes and arrange them in readable commits.
- Make sure your commit messages are in the proper format (see below).
- Push your changes to the branch in your fork of the repository.
- Make sure all linters and tests pass, and add any new tests as appropriate.
- Submit a pull request to the original repository.

## Building

Details about building crossplane can be found in [INSTALL.md](INSTALL.md).

## Coding Style and Linting

Crossplane projects are written in Go. Coding style is enforced by
[golangci-lint](https://github.com/golangci/golangci-lint). Crossplane's linter
configuration is [documented here](.golangci.yml). Builds will fail locally and
in CI if linter warnings are introduced:

```bash
$ make build
==> Linting /REDACTED/go/src/github.com/crossplaneio/crossplane/cluster/charts/crossplane
[INFO] Chart.yaml: icon is recommended

1 chart(s) linted, no failures
20:31:42 [ .. ] helm dep crossplane 0.1.0-136.g2dfb012.dirty
No requirements found in /REDACTED/go/src/github.com/crossplaneio/crossplane/cluster/charts/crossplane/charts.
20:31:42 [ OK ] helm dep crossplane 0.1.0-136.g2dfb012.dirty
20:31:42 [ .. ] golangci-lint
pkg/clients/azure/redis/redis.go:35:7: exported const `NamePrefix` should have comment or be unexported (golint)
const NamePrefix = "acr"
      ^
20:31:55 [FAIL]
make[2]: *** [go.lint] Error 1
make[1]: *** [build.all] Error 2
make: *** [build] Error 2
```

Note that Jenkins builds will not output linter warnings in the build log.
Instead `upbound-bot` will leave comments on your pull request when a build
fails due to linter warnings. You can run `make lint` locally to help determine
whether you've fixed any linter warnings detected by Jenkins.

In some cases linter warnings will be false positives. `golangci-lint` supports
`//nolint[:lintername]` comment directives in order to quell them. Please
include an explanatory comment if you must add a `//nolint` comment. You may
also submit a PR against [`.golangci.yml`](.golangci.yml) if you feel
particular linter warning should be permanently disabled.

## Comments

Comments should be added to all new methods and structures as is appropriate for the coding
language. Additionally, if an existing method or structure is modified sufficiently, comments should
be created if they do not yet exist and updated if they do.

The goal of comments is to make the code more readable and grokkable by future developers. Once you
have made your code as understandable as possible, add comments to make sure future developers can
understand (A) what this piece of code's responsibility is within Crossplane's architecture and (B) why it
was written as it was.

The below blog entry explains more the why's and how's of this guideline.
https://blog.codinghorror.com/code-tells-you-how-comments-tell-you-why/

For Go, Crossplane follows standard godoc guidelines.
A concise godoc guideline can be found here: https://blog.golang.org/godoc-documenting-go-code

## Commit Messages

We follow a rough convention for commit messages that is designed to answer two
questions: what changed and why. The subject line should feature the what and
the body of the commit should describe the why.

```
aws: add support for RDS controller

this commit enables controllers and apis for RDS. It
enables provisioning a RDS database.
```

The format can be described more formally as follows:

```
<subsystem>: <what changed>
<BLANK LINE>
<why this change was made>
<BLANK LINE>
<footer>
```

The first line is the subject and should be no longer than 70 characters, the
second line is always blank, and other lines should be wrapped at 80 characters.
This allows the message to be easier to read on GitHub as well as in various
git tools.


## Adding New Resources

### Project Organization
The Crossplane project is based on and intially created by using [Kubebuilder is a framework for building Kubernetes APIs](https://github.com/kubernetes-sigs/kubebuilder).

The Crossplane project organizes resources (api types and controllers) by grouping them by Cloud Provider with further sub-group by resource type 

The Kubebuilder framework does not provide good support for projects with multiple groups and group tiers which contain resources with overlapping names. 
For example:
```
pkg
├── apis
│   ├── aws
│   │   ├── apis.go
│   │   └── database
│   │       ├── group.go
│   │       └── v1alpha1
│   │           ├── doc.go
│   │           ├── rds_instance_types.go
│   │           ├── rds_instance_types_test.go
│   │           ├── register.go
│   │           ├── v1alpha1_suite_test.go
│   │           └── zz_generated.deepcopy.go
│   └── gcp
│       ├── apis.go
│       └── database
│           ├── group.go
│           └── v1alpha1
│               ├── cloudsql_instance_types.go
│               ├── cloudsql_instance_types_test.go
│               ├── doc.go
│               ├── register.go
│               ├── v1alpha1_suite_test.go
│               └── zz_generated.deepcopy.go
└── controller
    ├── aws
    │   ├── controller.go
    │   └── database
    │       ├── database_suite_test.go
    │       ├── rds_instance.go
    │       └── rds_instance_test.go
    └── gcp
        ├── controller.go
        └── database
            ├── cloudsql_instance.go
            ├── cloudsql_instance_test.go
            └── database_suite_test.go
```
In above example we have two groups with sub-group (tiers):
- aws/rds
- gcp/cloudsql

In addition both groups contain types with the same name: `Instance`

### Creating New Resource
There are several different ways you can approach the creation of the new resources:
#### Manual
Good ol' copy & paste of existing resource for both apis and controller (if new controller is needed for your resource) and update the copied code to tailor your needs.

#### Kubebuilder With New Project
Create and Initialize a new (temporary) kubebuilder project and create new resources: apis and controller(s), then copy them into Crossplane project following the established project organization.

To verify that new artifacts run: 
```bash
make build test
```

## Local Build and Test

To learn more about the developer iteration workflow, including how to locally test new types/controllers, please refer to the [Local Build](cluster/local/README.md) instructions.
