# ====================================================================================
# Setup Project

PROJECT_NAME := stack-aws
PROJECT_REPO := github.com/crossplaneio/$(PROJECT_NAME)

PLATFORMS ?= linux_amd64 linux_arm64
# -include will silently skip missing files, which allows us
# to load those files with a target in the Makefile. If only
# "include" was used, the make command would fail and refuse
# to run a target until the include commands succeeded.
-include build/makelib/common.mk

# ====================================================================================
# Setup Output

S3_BUCKET ?= crossplane.releases
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

GO_STATIC_PACKAGES = $(GO_PROJECT)/cmd/crossplane
GO_LDFLAGS += -X $(GO_PROJECT)/pkg/version.Version=$(VERSION)
GO_SUBDIRS += cmd pkg aws
-include build/makelib/golang.mk

# ====================================================================================
# Setup Kubebuilder

-include build/makelib/kubebuilder.mk

# ====================================================================================
# Setup Kubernetes tools

-include build/makelib/k8s_tools.mk

# ====================================================================================
# Setup Stacks

STACK_PACKAGE=stack-package
export STACK_PACKAGE
STACK_PACKAGE_REGISTRY=$(STACK_PACKAGE)/.registry
CRD_DIR=config/crd
STACK_PACKAGE_REGISTRY_SOURCE=config/stack/manifests

DOCKER_REGISTRY = crossplane
IMAGES = stack-aws
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

go.test.unit: $(KUBEBUILDER)

# Generate manifests e.g. CRD, RBAC etc.
manifests: vendor
	@$(INFO) Generating CRD manifests
	go run vendor/sigs.k8s.io/controller-tools/cmd/controller-gen/main.go crd:trivialVersions=true paths=./aws/... output:dir=$(CRD_DIR)
	@$(OK) Generating CRD manifests

# Generate a coverage report for cobertura applying exclusions on
# - generated file
cobertura:
	@cat $(GO_TEST_OUTPUT)/coverage.txt | \
		grep -v zz_generated.deepcopy | \
		$(GOCOVER_COBERTURA) > $(GO_TEST_OUTPUT)/cobertura-coverage.xml

# Ensure a PR is ready for review.
reviewable: vendor generate manifests lint

# integration tests
e2e.run: test-integration

# Run integration tests.
test-integration: $(KIND) $(KUBECTL) $(HELM)
	@$(INFO) running integration tests using kind $(KIND_VERSION)
	@$(ROOT_DIR)/cluster/local/integration_tests.sh || $(FAIL)
	@$(OK) integration tests passed

# Update the submodules, such as the common build scripts.
submodules:
	@git submodule sync
	@git submodule update --init --recursive

# This is for running out-of-cluster locally, and is for convenience. Running
# this make target will print out the command which was used. For more control,
# try running the binary directly with different arguments.
run: go.build
	@$(INFO) Running Crossplane locally out-of-cluster . . .
	@# To see other arguments that can be provided, run the command with --help instead
	$(GO_OUT_DIR)/$(PROJECT_NAME) --debug

# ====================================================================================
# Stacks related targets

# Initialize the stack package folder
$(STACK_PACKAGE_REGISTRY):
	@mkdir -p $(STACK_PACKAGE_REGISTRY)/resources
	@touch $(STACK_PACKAGE_REGISTRY)/app.yaml $(STACK_PACKAGE_REGISTRY)/install.yaml $(STACK_PACKAGE_REGISTRY)/rbac.yaml

build.artifacts: build-stack-package

build-stack-package: $(STACK_PACKAGE_REGISTRY)
# Copy CRDs over
#
# The reason this looks complicated is because it is
# preserving the original crd filenames and changing
# *.yaml to *.crd.yaml.
#
# An alternate and simpler-looking approach would
# be to cat all of the files into a single crd.yaml,
# but then we couldn't use per CRD metadata files.
	@$(INFO) building stack package in $(STACK_PACKAGE)
	@find $(CRD_DIR) -type f -name '*.yaml' | \
		while read filename ; do cat $$filename > \
		$(STACK_PACKAGE_REGISTRY)/resources/$$( basename $${filename/.yaml/.crd.yaml} ) \
		; done
	@cp -r $(STACK_PACKAGE_REGISTRY_SOURCE)/* $(STACK_PACKAGE_REGISTRY)

clean: clean-stack-package

clean-stack-package:
	@rm -rf $(STACK_PACKAGE)

.PHONY: manifests cobertura reviewable submodules fallthrough test-integration run clean-stack-package build-stack-package

# ====================================================================================
# Special Targets

define CROSSPLANE_MAKE_HELP
Crossplane Targets:
    manifests             Generate manifests e.g. CRD, RBAC etc.
    cobertura             Generate a coverage report for cobertura applying exclusions on generated files.
    reviewable            Ensure a PR is ready for review.
    submodules            Update the submodules, such as the common build scripts.
    run                   Run crossplane locally, out-of-cluster. Useful for development.
    build-stack-package   Builds the stack package contents in the stack package directory (./$(STACK_PACKAGE))
    clean-stack-package   Cleans out the generated stack package directory (./$(STACK_PACKAGE))

endef
# The reason CROSSPLANE_MAKE_HELP is used instead of CROSSPLANE_HELP is because the crossplane
# binary will try to use CROSSPLANE_HELP if it is set, and this is for something different.
export CROSSPLANE_MAKE_HELP

crossplane.help:
	@echo "$$CROSSPLANE_MAKE_HELP"

help-special: crossplane.help

.PHONY: crossplane.help help-special
