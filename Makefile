# The Go-based tools are defined in Makefile.tools.mk.
include Makefile.tools.mk

# The list of commands that can be built.
COMMANDS := controller extproc

# This is the package that contains the version information for the build.
GIT_COMMIT:=$(shell git rev-parse HEAD)
VERSION_PACKAGE := github.com/envoyproxy/ai-gateway/internal/version
GO_LDFLAGS += -X $(VERSION_PACKAGE).Version=$(GIT_COMMIT)

# This is the directory where the built artifacts will be placed.
OUTPUT_DIR ?= out

# Arguments for docker builds.
OCI_REGISTRY ?= ghcr.io/envoyproxy/ai-gateway
TAG ?= latest
ENABLE_MULTI_PLATFORMS ?= false
HELM_CHART_VERSION ?= v0.0.0-latest

# Arguments for go test. This can be used, for example, to run specific tests via
# `GO_TEST_ARGS="-run TestName/foo/etc -v -race"`.
GO_TEST_ARGS ?= -v -race
# Arguments for go test in e2e tests in addition to GO_TEST_ARGS, applicable to test-e2e, test-extproc, and test-controller.
GO_TEST_E2E_ARGS ?= -count=1

# This will print out the help message for contributing to the project.
.PHONY: help
help:
	@echo "Usage: make <target>"
	@echo ""
	@echo "All core targets needed for contributing:"
	@echo "  precommit       	 Run all necessary steps to prepare for a commit."
	@echo "  test            	 Run the unit tests for the codebase."
	@echo "  test-cel        	 Run the integration tests of CEL validation rules in API definitions with envtest."
	@echo "                  	 This will be needed when changing API definitions."
	@echo "  test-extproc    	 Run the integration tests for extproc without controller or k8s at all."
	@echo "  test-controller	 Run the integration tests for the controller with envtest."
	@echo "  test-e2e          	 Run the end-to-end tests with a local kind cluster."
	@echo ""
	@echo "For example, 'make precommit test' should be enough for initial iterations, and later 'make test-cel' etc."
	@echo "Note that some cases run by 'make test-e2e' or 'make test-extproc' use credentials and these will be skipped when not available."
	@echo ""
	@echo ""

# This runs the linter, formatter, and tidy on the codebase.
.PHONY: lint
lint: golangci-lint
	@echo "lint => ./..."
	@$(GOLANGCI_LINT) run --build-tags==test_cel_validation,test_controller,test_extproc ./...

.PHONY: codespell
CODESPELL_SKIP := $(shell cat .codespell.skip | tr \\n ',')
CODESPELL_IGNORE_WORDS := ".codespell.ignorewords"
codespell: $(CODESPELL)
	@echo "spell => ./..."
	@$(CODESPELL) --skip $(CODESPELL_SKIP) --ignore-words $(CODESPELL_IGNORE_WORDS)

.PHONY: yamllint
yamllint: $(YAMLLINT)
	@echo "yamllint => ./..."
	@$(YAMLLINT) --config-file=.yamllint $$(git ls-files :*.yml :*.yaml | xargs -L1 dirname | sort -u)

# This runs the formatter on the codebase as well as goimports via gci.
.PHONY: format
format: gci gofumpt
	@echo "format => *.go"
	@find . -type f -name '*.go' | xargs gofmt -s -w
	@find . -type f -name '*.go' | xargs $(GO_FUMPT) -l -w
	@echo "gci => *.go"
	@$(GCI) write -s standard -s default -s "prefix(github.com/envoyproxy/ai-gateway)" `find . -name '*.go'`

# This runs go mod tidy on every module.
.PHONY: tidy
tidy:
	@find . -name "go.mod" \
	| grep go.mod \
	| xargs -I {} bash -c 'dirname {}' \
	| xargs -I {} bash -c 'echo "tidy => {}"; cd {}; go mod tidy -v; '

# This re-generates the CRDs for the API defined in the api/v1alpha1 directory.
.PHONY: apigen
apigen: controller-gen
	@echo "apigen => ./api/v1alpha1/..."
	@$(CONTROLLER_GEN) object crd paths="./api/v1alpha1/..." output:dir=./api/v1alpha1 output:crd:dir=./manifests/charts/ai-gateway-helm/crds

# This runs all necessary steps to prepare for a commit.
.PHONY: precommit
precommit: tidy codespell apigen format lint editorconfig yamllint helm-lint

# This runs precommit and checks for any differences in the codebase, failing if there are any.
.PHONY: check
check: precommit
	@if [ ! -z "`git status -s`" ]; then \
		echo "The following differences will fail CI until committed:"; \
		git diff --exit-code; \
	fi

# This runs the editorconfig-checker on the codebase.
editorconfig: editorconfig-checker
	@echo "running editorconfig-checker"
	@$(EDITORCONFIG_CHECKER)

# This runs the unit tests for the codebase.
.PHONY: test
test:
	@echo "test => ./..."
	@go test $(GO_TEST_ARGS) ./...

ENVTEST_K8S_VERSIONS ?= 1.29.0 1.30.0 1.31.0

# This runs the integration tests of CEL validation rules in API definitions.
#
# This requires the EnvTest binary to be built.
.PHONY: test-cel
test-cel: envtest apigen
	@for k8sVersion in $(ENVTEST_K8S_VERSIONS); do \
  		echo "Run CEL Validation on k8s $$k8sVersion"; \
        KUBEBUILDER_ASSETS="$$($(ENVTEST) use $$k8sVersion -p path)" \
                 go test ./tests/cel-validation $(GO_TEST_ARGS) $(GO_TEST_E2E_ARGS) --tags test_cel_validation; \
    done

# This runs the end-to-end tests for extproc without controller or k8s at all.
# It is useful for the fast iteration of the extproc code.
#
# This requires the extproc binary to be built as well as Envoy binary to be available in the PATH.
.PHONY: test-extproc # This requires the extproc binary to be built.
test-extproc: build.extproc
	@$(MAKE) build.extproc_custom_router CMD_PATH_PREFIX=examples
	@$(MAKE) build.testupstream CMD_PATH_PREFIX=tests
	@echo "Run ExtProc test"
	@go test ./tests/extproc/... $(GO_TEST_ARGS) $(GO_TEST_E2E_ARGS) -tags test_extproc -v

# This runs the end-to-end tests for the controller with EnvTest.
.PHONY: test-controller
test-controller: envtest apigen
	@for k8sVersion in $(ENVTEST_K8S_VERSIONS); do \
  		echo "Run Controller tests on k8s $$k8sVersion"; \
        KUBEBUILDER_ASSETS="$$($(ENVTEST) use $$k8sVersion -p path)" \
                 go test ./tests/controller $(GO_TEST_ARGS) $(GO_TEST_E2E_ARGS) -tags test_controller; \
    done

# This runs the end-to-end tests for the controller and extproc with a local kind cluster.
#
# This requires the docker images to be built.
.PHONY: test-e2e
test-e2e: kind
	@$(MAKE) docker-build DOCKER_BUILD_ARGS="--load"
	@$(MAKE) docker-build.testupstream CMD_PATH_PREFIX=tests DOCKER_BUILD_ARGS="--load"
	@echo "Run E2E tests"
	@go test ./tests/e2e/... $(GO_TEST_ARGS) $(GO_TEST_E2E_ARGS) -tags test_e2e

# This builds a binary for the given command under the internal/cmd directory.
#
# Example:
# - `make build.controller`: will build the cmd/controller directory.
# - `make build.extproc`: will build the cmd/extproc directory.
# - `make build.extproc_custom_router CMD_PATH_PREFIX=examples`: will build the examples/extproc_custom_router directory.
# - `make build.testupstream CMD_PATH_PREFIX=tests`: will build the tests/testupstream directory.
#
# By default, this will build for the current GOOS and GOARCH.
# To build for multiple platforms, set the GOOS_LIST and GOARCH_LIST variables.
#
# Example:
# - `make build.controller GOOS_LIST="linux darwin" GOARCH_LIST="amd64 arm64"`
CMD_PATH_PREFIX ?= cmd
GOOS_LIST ?= $(shell go env GOOS)
GOARCH_LIST ?= $(shell go env GOARCH)
.PHONY: build.%
build.%:
	$(eval COMMAND_NAME := $(subst build.,,$@))
	@mkdir -p $(OUTPUT_DIR)
	@for goos in $(GOOS_LIST); do \
		for goarch in $(GOARCH_LIST); do \
			echo "-> Building $(COMMAND_NAME) for $$goos/$$goarch"; \
			CGO_ENABLED=0 GOOS=$$goos GOARCH=$$goarch go build -ldflags "$(GO_LDFLAGS)" \
				-o $(OUTPUT_DIR)/$(COMMAND_NAME)-$$goos-$$goarch ./$(CMD_PATH_PREFIX)/$(COMMAND_NAME); \
			echo "<- Built $(OUTPUT_DIR)/$(COMMAND_NAME)-$$goos-$$goarch"; \
		done; \
	done

# This builds a docker image for a given command.
#
# Example:
# - `make docker-build.controller`: will build the controller command.
# - `make docker-build.extproc`: will build the extproc command.
#
# By default, this will build for the current GOARCH and linux.
# To build for multiple platforms, set the ENABLE_MULTI_PLATFORMS variable to true.
#
# Example:
# - `make docker-build.controller ENABLE_MULTI_PLATFORMS=true`
#
# Also, DOCKER_BUILD_ARGS can be set to pass additional arguments to the docker build command.
#
# Example:
# - `make docker-build.controller ENABLE_MULTI_PLATFORMS=true DOCKER_BUILD_ARGS="--push"` to push the image to the registry.
# - `make docker-build.controller ENABLE_MULTI_PLATFORMS=true DOCKER_BUILD_ARGS="--load"` to load the image after building.
#
# By default, the image tag is set to `latest`. `TAG` can be set to a different value.
#
# Example:
# - `make docker-build.controller TAG=v1.2.3`
#
# To build the main functions outside cmd/ directory, set CMD_PATH_PREFIX to the directory containing the main function.
#
# Example:
# - `make docker-build.extproc_custom_router CMD_PATH_PREFIX=examples`
# - `make docker-build.testupstream CMD_PATH_PREFIX=tests`
.PHONY: docker-build.%
ifeq ($(ENABLE_MULTI_PLATFORMS),true)
docker-build.%: GOARCH_LIST = amd64 arm64
docker-build.%: PLATFORMS = --platform linux/amd64,linux/arm64
endif
docker-build.%:
	$(eval COMMAND_NAME := $(subst docker-build.,,$@))
	@$(MAKE) build.$(COMMAND_NAME) GOOS_LIST="linux" GOARCH_LIST="$(GOARCH_LIST)"
	docker buildx build . -t $(OCI_REGISTRY)/$(COMMAND_NAME):$(TAG) --build-arg COMMAND_NAME=$(COMMAND_NAME) $(PLATFORMS) $(DOCKER_BUILD_ARGS)

# This builds docker images for all commands under cmd/ directory. All options for `docker-build.%` apply.
#
# Example:
# - `make docker-build`
# - `make docker-build ENABLE_MULTI_PLATFORMS=true DOCKER_BUILD_ARGS="--load"`
# - `make docker-build ENABLE_MULTI_PLATFORMS=true DOCKER_BUILD_ARGS="--push" TAG=v1.2.3`
.PHONE: docker-build
docker-build:
	@$(foreach COMMAND_NAME,$(COMMANDS),$(MAKE) docker-build.$(COMMAND_NAME);)

HELM_DIR := ./manifests/charts/ai-gateway-helm

# This lints the helm chart, ensuring that it is for packaging.
#
# This uses the locally installed helm binary (TODO make helm installed via Makefile.tools.mk).
.PHONY: helm-lint
helm-lint:
	@echo "helm-lint => .${HELM_DIR}"
	@helm lint ${HELM_DIR}

# This packages the helm chart into a tgz file, ready for deployment as well as for pushing to the OCI registry.
# This must pass before `helm-push` can be run as well as on any commit.
#
# This uses the locally installed helm binary (TODO make helm installed via Makefile.tools.mk).
.PHONY: helm-package
helm-package: helm-lint
	@echo "helm-package => ${HELM_DIR}"
	@helm package ${HELM_DIR} --version ${HELM_CHART_VERSION} -d ${OUTPUT_DIR}

# This pushes the helm chart to the OCI registry, requiring the access to the registry endpoint.
#
# This uses the locally installed helm binary (TODO make helm installed via Makefile.tools.mk).
.PHONY: helm-push
helm-push: helm-package
	@echo "helm-push => .${HELM_DIR}"
	@helm push ${OUTPUT_DIR}/ai-gateway-helm-${HELM_CHART_VERSION}.tgz oci://${OCI_REGISTRY}
