# Building and Installing the Crossplane AWS Stack

`provider-aws` is composed of a golang project and can be built directly with standard `golang` tools. We currently support two different platforms for building:

* Linux: most modern distros should work although most testing has been done on Ubuntu
* Mac: macOS 10.6+ is supported

## Build Requirements

An Intel-based machine (recommend 2+ cores, 2+ GB of memory and 128GB of SSD). Inside your build environment (Docker for Mac or a VM), 6+ GB memory is also recommended.

The following tools are need on the host:

* curl
* docker (1.12+) or Docker for Mac (17+)
* git
* make
* golang
* rsync (if you're using the build container on mac)
* helm (v2.8.2+)
* kubebuilder (v1.0.4+)

## Build

You can build the Crossplane AWS Stack for the host platform by simply running the command below.
Building in parallel with the `-j` option is recommended.

```console
make -j4
```

The first time `make` is run, the build submodule will be synced and
updated. After initial setup, it can be updated by running `make submodules`.

Run `make help` for more options.

## Building inside the cross container

Official Crossplane builds are done inside a build container. This ensures that we get a consistent build, test and release environment. To run the build inside the cross container run:

```console
> build/run make -j4
```

The first run of `build/run` will build the container itself and could take a few minutes to complete, but subsequent builds should go much faster.

## Install

TBD: Steps to install the AWS provider package into a Crossplane cluster
