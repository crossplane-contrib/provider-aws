# ====================================================================================
# Setup Project

PROJECT_NAME := provider-aws
PROJECT_REPO := github.com/crossplane/$(PROJECT_NAME)

PLATFORMS ?= linux_amd64 linux_arm64

CODE_GENERATOR_COMMIT ?= f8c0f4711a91e50335450521e3c49673d135d56d
GENERATED_SERVICES := $(shell find ./apis -type f -name generator-config.yaml | cut -d/ -f 3 | tr '\n' ' ')

# kind-related versions
KIND_VERSION ?= v0.11.1
KIND_NODE_IMAGE_TAG ?= v1.19.11

# -include will silently skip missing files, which allows us
# to load those files with a target in the Makefile. If only
# "include" was used, the make command would fail and refuse
# to run a target until the include commands succeeded.
-include build/makelib/common.mk

# ====================================================================================
# Setup Output

-include build/makelib/output.mk

# ====================================================================================
# Setup Go

# Set a sane default so that the nprocs calculation below is less noisy on the initial
# loading of this file
NPROCS ?= 1

# each of our test suites starts a kube-apiserver and running many test suites in
# parallel can lead to high CPU utilization. by default we reduce the parallelism
# to half the number of CPU cores.
GO_TEST_PARALLEL := $(shell echo $$(( $(NPROCS) / 2 )))

GO_STATIC_PACKAGES = $(GO_PROJECT)/cmd/provider
GO_LDFLAGS += -X $(GO_PROJECT)/pkg/version.Version=$(VERSION)
GO_SUBDIRS += cmd pkg apis
GO111MODULE = on
-include build/makelib/golang.mk

# ====================================================================================
# Setup Kubernetes tools

-include build/makelib/k8s_tools.mk

# ====================================================================================
# Setup Images

DOCKER_REGISTRY ?= crossplane
IMAGES = provider-aws provider-aws-controller
-include build/makelib/image.mk

# ====================================================================================
# Targets

# run `make help` to see the targets and options

# We want submodules to be set up the first time `make` is run.
# We manage the build/ folder and its Makefiles as a submodule.
# The first time `make` is run, the includes of build/*.mk files will
# all fail, and this target will be run. The next time, the default as defined
# by the includes will be run instead.
fallthrough: submodules
	@echo Initial setup complete. Running make again . . .
	@make

# Generate a coverage report for cobertura applying exclusions on
# - generated file
cobertura:
	@cat $(GO_TEST_OUTPUT)/coverage.txt | \
		grep -v zz_generated.deepcopy | \
		$(GOCOVER_COBERTURA) > $(GO_TEST_OUTPUT)/cobertura-coverage.xml

crds.clean:
	@$(INFO) cleaning generated CRDs
	@find package/crds -name '*.yaml' -exec sed -i.sed -e '1,2d' {} \; || $(FAIL)
	@find package/crds -name '*.yaml.sed' -delete || $(FAIL)
	@$(OK) cleaned generated CRDs

generate.init: services.all
generate.done: crds.clean

manifests:
	@$(WARN) Deprecated. Please run make generate instead.

# integration tests
e2e.run: test-integration

# Run integration tests.
test-integration: $(KIND) $(KUBECTL) $(HELM3)
	@$(INFO) running integration tests using kind $(KIND_VERSION)
	@KIND_NODE_IMAGE_TAG=${KIND_NODE_IMAGE_TAG} $(ROOT_DIR)/cluster/local/integration_tests.sh || $(FAIL)
	@$(OK) integration tests passed

# Update the submodules, such as the common build scripts.
submodules:
	@git submodule sync
	@git submodule update --init --recursive

# NOTE(hasheddan): the build submodule currently overrides XDG_CACHE_HOME in
# order to force the Helm 3 to use the .work/helm directory. This causes Go on
# Linux machines to use that directory as the build cache as well. We should
# adjust this behavior in the build submodule because it is also causing Linux
# users to duplicate their build cache, but for now we just make it easier to
# identify its location in CI so that we cache between builds.
go.cachedir:
	@go env GOCACHE

# This is for running out-of-cluster locally, and is for convenience. Running
# this make target will print out the command which was used. For more control,
# try running the binary directly with different arguments.
run: go.build
	@$(INFO) Running Crossplane locally out-of-cluster . . .
	@# To see other arguments that can be provided, run the command with --help instead
	$(GO_OUT_DIR)/provider --debug

.PHONY: cobertura manifests submodules fallthrough test-integration run crds.clean

# NOTE(muvaf): ACK Code Generator is a separate Go module, hence we need to
# be in its root directory to call "go run" properly.
services: $(GOIMPORTS)
	@if [ ! -d "$(WORK_DIR)/code-generator" ]; then \
		cd $(WORK_DIR) && git clone "https://github.com/aws-controllers-k8s/code-generator.git"; \
	fi
	@cd $(WORK_DIR)/code-generator && git fetch origin && git checkout $(CODE_GENERATOR_COMMIT)
	@for svc in $(SERVICES); do \
		$(INFO) Generating $$svc controllers and CRDs; \
		PATH="${PATH}:$(TOOLS_HOST_DIR)"; \
		cd $(WORK_DIR)/code-generator && go run -tags codegen cmd/ack-generate/main.go crossplane $$svc --output ../../ || exit 1; \
		$(OK) Generating $$svc controllers and CRDs; \
	done

services.all:
	@$(MAKE) services SERVICES="$(GENERATED_SERVICES)"

# ====================================================================================
# Special Targets

define CROSSPLANE_MAKE_HELP
Crossplane Targets:
    cobertura             Generate a coverage report for cobertura applying exclusions on generated files.
    submodules            Update the submodules, such as the common build scripts.
    run                   Run crossplane locally, out-of-cluster. Useful for development.

endef
# The reason CROSSPLANE_MAKE_HELP is used instead of CROSSPLANE_HELP is because the crossplane
# binary will try to use CROSSPLANE_HELP if it is set, and this is for something different.
export CROSSPLANE_MAKE_HELP

crossplane.help:
	@echo "$$CROSSPLANE_MAKE_HELP"

help-special: crossplane.help

.PHONY: crossplane.help help-special
