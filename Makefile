# ====================================================================================
# Setup Project

PROJECT_NAME := provider-aws
PROJECT_REPO := github.com/crossplane-contrib/$(PROJECT_NAME)

PLATFORMS ?= linux_amd64 linux_arm64

GENERATED_SERVICES ?= $(shell find ./apis -type f -name generator-config.yaml | cut -d/ -f 3 | tr '\n' ' ')

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
GOLANGCILINT_VERSION := $(shell grep 'GOLANGCI_VERSION' .github/workflows/ci.yml | sed -n 's/.*: *["'\'']*v\([0-9]\+\.[0-9]\+\.[0-9]\+\).*/\1/p')
-include build/makelib/golang.mk

# ====================================================================================
# Setup Kubernetes tools

UP_VERSION = v0.27.0
UP_CHANNEL = stable
UPTEST_VERSION = v0.11.0

-include build/makelib/k8s_tools.mk

# ====================================================================================
# Setup Images

IMAGES = provider-aws
-include build/makelib/imagelight.mk

# ====================================================================================
# Setup XPKG

XPKG_REG_ORGS ?= xpkg.upbound.io/crossplane-contrib index.docker.io/crossplanecontrib
# NOTE(hasheddan): skip promoting on xpkg.upbound.io as channel tags are
# inferred.
XPKG_REG_ORGS_NO_PROMOTE ?= xpkg.upbound.io/crossplane-contrib
XPKGS = provider-aws
-include build/makelib/xpkg.mk

# NOTE(hasheddan): we force image building to happen prior to xpkg build so that
# we ensure image is present in daemon.
xpkg.build.provider-aws: do.build.images

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

generate.init: services.all

manifests:
	@$(WARN) Deprecated. Please run make generate instead.

# integration tests
e2e.run: test-integration

# Run integration tests.
test-integration: $(KIND) $(KUBECTL) $(UP) $(HELM)
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

# NOTE(hasheddan): we must ensure up is installed in tool cache prior to build
# as including the k8s_tools machinery prior to the xpkg machinery sets UP to
# point to tool cache.
build.init: $(UP)

# This is for running out-of-cluster locally, and is for convenience. Running
# this make target will print out the command which was used. For more control,
# try running the binary directly with different arguments.
run: go.build
	@$(INFO) Running Crossplane locally out-of-cluster . . .
	@# To see other arguments that can be provided, run the command with --help instead
	$(GO_OUT_DIR)/provider --debug

.PHONY: cobertura manifests submodules fallthrough test-integration run crds.clean

AWS_SDK_GO_VERSION ?= $(shell awk '/github.com\/aws\/aws-sdk-go / { print $$2 }' < go.mod)
ACK_VERSION ?= $(shell awk '/github.com\/aws-controllers-k8s\/code-generator / { print $$2 }' < go.mod)
ACK_GENERATE_DIR ?= $(CACHE_DIR)/ack-generate
ACK_GENERATE ?= $(ACK_GENERATE_DIR)/ack-generate_$(ACK_VERSION)

# Because ack-generate is executed many times during generate, we build a binary
# to speed up code generation compared to go run.
$(ACK_GENERATE):
	@$(INFO) Building ack-generate $(ACK_VERSION)
	@mkdir -p "$(ACK_GENERATE_DIR)"
	@$(GO) build -o "$(ACK_GENERATE)" -tags codegen github.com/aws-controllers-k8s/code-generator/cmd/ack-generate || $(FAIL)
	@$(OK) Built ack-generate $(ACK_VERSION)

services: $(ACK_GENERATE) $(GOIMPORTS)
	@for svc in $(SERVICES); do \
		$(INFO) Generating $$svc controllers and CRDs; \
		PATH="${PATH}:$(TOOLS_HOST_DIR)"; \
		$(ACK_GENERATE) crossplane --aws-sdk-go-version $(AWS_SDK_GO_VERSION) $$svc --output=./ || exit 1; \
		$(OK) Generating $$svc controllers and CRDs; \
	done

services.all:
	@$(MAKE) services SERVICES="$(GENERATED_SERVICES)"

crddiff: $(UPTEST)
	@$(INFO) Checking breaking CRD schema changes
	@for crd in $${MODIFIED_CRD_LIST}; do \
		if ! git cat-file -e "$${GITHUB_BASE_REF}:$${crd}" 2>/dev/null; then \
			echo "CRD $${crd} does not exist in the $${GITHUB_BASE_REF} branch. Skipping..." ; \
			continue ; \
		fi ; \
		echo "Checking $${crd} for breaking API changes..." ; \
		changes_detected=$$($(UPTEST) crddiff revision <(git cat-file -p "$${GITHUB_BASE_REF}:$${crd}") "$${crd}" 2>&1) ; \
		if [[ $$? != 0 ]] ; then \
			printf "\033[31m"; echo "Breaking change detected!"; printf "\033[0m" ; \
			echo "$${changes_detected}" ; \
			echo ; \
		fi ; \
	done
	@$(OK) Checking breaking CRD schema changes

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
