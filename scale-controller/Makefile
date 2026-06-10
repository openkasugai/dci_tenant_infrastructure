# Image URL to use all building/pushing image targets
IMG ?= controller:latest
ALERT_RECEIVER_IMG ?= card-alert-receiver:latest

# Get the currently used golang install path (in GOPATH/bin, unless GOBIN is set)
ifeq (,$(shell go env GOBIN))
GOBIN=$(shell go env GOPATH)/bin
else
GOBIN=$(shell go env GOBIN)
endif

# CONTAINER_TOOL defines the container tool to be used for building images.
# Be aware that the target commands are only tested with Docker which is
# scaffolded by default. However, you might want to replace it to use other
# tools. (i.e. podman)
CONTAINER_TOOL ?= docker

PRJ_IMAGE_BUILD_TOOL ?= buildah

# Setting SHELL to bash allows bash commands to be executed by recipes.
# Options are set to exit when a recipe line exits non-zero or a piped command fails.
SHELL = /usr/bin/env bash -o pipefail
.SHELLFLAGS = -ec

.PHONY: all
all: build

##@ General

# The help target prints out all targets with their descriptions organized
# beneath their categories. The categories are represented by '##@' and the
# target descriptions by '##'. The awk command is responsible for reading the
# entire set of makefiles included in this invocation, looking for lines of the
# file as xyz: ## something, and then pretty-format the target and help. Then,
# if there's a line with ##@ something, that gets pretty-printed as a category.
# More info on the usage of ANSI control characters for terminal formatting:
# https://en.wikipedia.org/wiki/ANSI_escape_code#SGR_parameters
# More info on the awk command:
# http://linuxcommand.org/lc3_adv_awk.php

.PHONY: help
help: ## Display this help.
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_0-9-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

##@ Development

.PHONY: manifests
manifests: manifests-kustomize manifests-helm ## Generate WebhookConfiguration, ClusterRole and CustomResourceDefinition objects.
manifests-kustomize: controller-gen
	@echo "Generating CRDs and RBAC for kustomize..."
	$(CONTROLLER_GEN) rbac:roleName=manager-role crd webhook paths="./..." output:crd:artifacts:config=config/crd/bases

.PHONY: generate
generate: controller-gen ## Generate code containing DeepCopy, DeepCopyInto, and DeepCopyObject method implementations.
	$(CONTROLLER_GEN) object:headerFile="hack/boilerplate.go.txt" paths="./..."

.PHONY: fmt
fmt: ## Run go fmt against code.
	go fmt ./...

.PHONY: vet
vet: ## Run go vet against code.
	go vet ./...

# option for go test
PACKAGE_NAME ?= $(shell go list -m)
GOTEST_OPTS ?=
GOTEST_TARGET_PKGS ?= $(shell go list ./... \
	| grep -v -e "$(PACKAGE_NAME)/pkg/alertreceiver$$" \
	| grep -v -e "$(PACKAGE_NAME)/pkg/alertreceiver/common$$" \
	| grep -v $(PACKAGE_NAME)/api/ \
	| grep -v $(PACKAGE_NAME)/test/ \
	)
GOTEST_COVER_PKGS ?= $(shell go list ./... \
	| grep -v $(PACKAGE_NAME)/test/ \
	| tr '\n' ',' \
	| sed -e 's/,$$//' \
	)
ENV_DOCKER_PROVIDER_ENABLE ?= true
.PHONY: test
test: manifests generate fmt vet setup-envtest ## Run tests.
	@echo ""
	env ENV_DOCKER_PROVIDER_ENABLE="${ENV_DOCKER_PROVIDER_ENABLE}" \
		KUBEBUILDER_ASSETS="$(shell $(ENVTEST) use $(ENVTEST_K8S_VERSION) --bin-dir $(LOCALBIN) -p path)" \
		go test $(GOTEST_TARGET_PKGS) \
		-coverpkg="$(GOTEST_COVER_PKGS)" \
		-coverprofile cover.out $(GOTEST_OPTS)
	go tool cover -html=cover.out -o coverage.html

# End-to-end tests
E2E_SCRIPT := $(CURDIR)/test/e2e/e2e.sh
MANAGEMENT_CLUSTER_NAME ?= sc-e2e-mc
WORKLOAD_CLUSTER_NAME ?= sc-e2e-wc
# Add '-p docker' to use DockerProvider as infrastructure provider in e2e tests.
E2E_EXTRA_OPTS ?= 
E2E_OPTS ?= -j $(CURDIR)/test/e2e/config.json $(E2E_EXTRA_OPTS)
E2E_LOGS_DIR ?= $(CURDIR)/test/e2e/e2e-logs

.PHONY: test-e2e
test-e2e: $(E2E_SCRIPT) | $(E2E_LOGS_DIR) ## Run the e2e tests. Expected an isolated environment using Kind.
	@echo "Running end-to-end tests..."
	@echo "When you stopped the tests in the middle, please delete clusters used in the tests:"
	@echo "* kind delete cluster -n $(MANAGEMENT_CLUSTER_NAME)"
	@echo "* kind delete cluster -n $(WORKLOAD_CLUSTER_NAME)"
	cd $(E2E_LOGS_DIR) && $(E2E_SCRIPT) -m $(MANAGEMENT_CLUSTER_NAME) -w $(WORKLOAD_CLUSTER_NAME) -c -x -d $(E2E_OPTS)

$(E2E_LOGS_DIR):
	mkdir -p $@

GOLINT_OPT_MAX_ISSUES ?= 0
GOLINT_OPT_MAX_SAME_ISSUES ?= 0
GOLINT_OPTS ?= --max-issues-per-linter=$(GOLINT_OPT_MAX_ISSUES) --max-same-issues=$(GOLINT_OPT_MAX_SAME_ISSUES)
.PHONY: lint
lint: golangci-lint ## Run golangci-lint linter
	$(GOLANGCI_LINT) run $(GOLINT_OPTS)

.PHONY: lint-fix
lint-fix: golangci-lint ## Run golangci-lint linter and perform fixes
	$(GOLANGCI_LINT) run --fix

.PHONY: lint-config
lint-config: golangci-lint ## Verify golangci-lint linter configuration
	$(GOLANGCI_LINT) config verify

##@ Build

.PHONY: build
build: manifests generate fmt vet ## Build manager binary.
	go build -o bin/manager cmd/controller/main.go

.PHONY: run
run: manifests generate fmt vet ## Run a controller from your host.
	go run ./cmd/controller/main.go

# If you wish to build the manager image targeting other platforms you can use the --platform flag.
# (i.e. docker build --platform linux/arm64). However, you must enable docker buildKit for it.
# More info: https://docs.docker.com/develop/develop-images/build_enhancements/
.PHONY: docker-build
docker-build: docker-build-manager docker-build-alert-receiver ## Build docker images with the manager and alert-receiver.
docker-build-manager: ## Build docker image with the manager.
	$(CONTAINER_TOOL) build -t ${IMG} .. -f $(CURDIR)/Dockerfile.local
# 	$(CONTAINER_TOOL) build -t ${IMG} .

.PHONY: docker-build-debug
docker-build-debug: ## Build docker image with the manager and the delve binary.
	$(CONTAINER_TOOL) build -t ${IMG} .. -f $(CURDIR)/Dockerfile.local.debug
# 	$(CONTAINER_TOOL) build -t ${IMG} . -f Dockerfile.local.debug

docker-build-alert-receiver: ## Build docker image with the alert-receiver.
	$(CONTAINER_TOOL) build -t ${ALERT_RECEIVER_IMG} .. -f $(CURDIR)/Dockerfile.alert-receiver.local

.PHONY: docker-push
docker-push: ## Push docker image with the manager.
	$(CONTAINER_TOOL) push ${IMG}

# PLATFORMS defines the target platforms for the manager image be built to provide support to multiple
# architectures. (i.e. make docker-buildx IMG=myregistry/mypoperator:0.0.1). To use this option you need to:
# - be able to use docker buildx. More info: https://docs.docker.com/build/buildx/
# - have enabled BuildKit. More info: https://docs.docker.com/develop/develop-images/build_enhancements/
# - be able to push the image to your registry (i.e. if you do not set a valid value via IMG=<myregistry/image:<tag>> then the export will fail)
# To adequately provide solutions that are compatible with multiple platforms, you should consider using this option.
PLATFORMS ?= linux/arm64,linux/amd64,linux/s390x,linux/ppc64le
.PHONY: docker-buildx
docker-buildx: ## Build and push docker image for the manager for cross-platform support
	# copy existing Dockerfile and insert --platform=${BUILDPLATFORM} into Dockerfile.cross, and preserve the original Dockerfile
	sed -e '1 s/\(^FROM\)/FROM --platform=\$$\{BUILDPLATFORM\}/; t' -e ' 1,// s//FROM --platform=\$$\{BUILDPLATFORM\}/' Dockerfile > Dockerfile.cross
	- $(CONTAINER_TOOL) buildx create --name scale-controller-builder
	$(CONTAINER_TOOL) buildx use scale-controller-builder
	- $(CONTAINER_TOOL) buildx build --push --platform=$(PLATFORMS) --tag ${IMG} -f Dockerfile.cross .
	- $(CONTAINER_TOOL) buildx rm scale-controller-builder
	rm Dockerfile.cross

.PHONY: buildah-build
buildah-build: buildah-build-manager buildah-build-alert-receiver ## Build buildah image with the manager.
buildah-build-manager: ## Build buildah image with the manager.
	$(PRJ_IMAGE_BUILD_TOOL) bud -t ${IMG} -f $(CURDIR)/Dockerfile.local ..
# 	$(PRJ_IMAGE_BUILD_TOOL) bud -t ${IMG} .

buildah-build-alert-receiver: ## Build buildah image with the alert-receiver.
	$(PRJ_IMAGE_BUILD_TOOL) bud -t ${ALERT_RECEIVER_IMG} -f $(CURDIR)/Dockerfile.alert-receiver.local ..

.PHONY: buildah-push
buildah-push: ## Push image with the manager.
	$(PRJ_IMAGE_BUILD_TOOL) push ${IMG}

.PHONY: build-installer
build-installer: manifests generate kustomize ## Generate a consolidated YAML with CRDs and deployment.
	mkdir -p dist
	cd config/manager && $(KUSTOMIZE) edit set image controller=${IMG}
	$(KUSTOMIZE) build config/default > dist/install.yaml

.PHONY: build-installer-debug
build-installer-debug: manifests generate kustomize ## Generate a consolidated YAML with CRDs and deployment.
	mkdir -p dist
	cd config/manager && $(KUSTOMIZE) edit set image controller=${IMG}
	$(KUSTOMIZE) build config/overlays/debug > dist/install.yaml

.PHONY: build-installer-docker
build-installer-docker: manifests generate kustomize ## Generate a consolidated YAML with CRDs and deployment.
	mkdir -p dist
	cd config/manager && $(KUSTOMIZE) edit set image controller=${IMG}
	$(KUSTOMIZE) build config/overlays/docker > dist/install.yaml

.PHONY: build-installer-docker-debug
build-installer-docker-debug: manifests generate kustomize ## Generate a consolidated YAML with CRDs and deployment.
	mkdir -p dist
	cd config/manager && $(KUSTOMIZE) edit set image controller=${IMG}
	$(KUSTOMIZE) build config/overlays/docker-debug > dist/install.yaml

##@ Deployment

ifndef ignore-not-found
  ignore-not-found = false
endif

.PHONY: install
install: manifests kustomize ## Install CRDs into the K8s cluster specified in ~/.kube/config.
	$(KUSTOMIZE) build config/crd | $(KUBECTL) apply -f -

.PHONY: uninstall
uninstall: manifests kustomize ## Uninstall CRDs from the K8s cluster specified in ~/.kube/config. Call with ignore-not-found=true to ignore resource not found errors during deletion.
	$(KUSTOMIZE) build config/crd | $(KUBECTL) delete --ignore-not-found=$(ignore-not-found) -f -

.PHONY: deploy
deploy: manifests kustomize ## Deploy controller to the K8s cluster specified in ~/.kube/config.
	cd config/manager && $(KUSTOMIZE) edit set image controller=${IMG}
	$(KUSTOMIZE) build config/default | $(KUBECTL) apply -f -

.PHONY: undeploy
undeploy: kustomize ## Undeploy controller from the K8s cluster specified in ~/.kube/config. Call with ignore-not-found=true to ignore resource not found errors during deletion.
	$(KUSTOMIZE) build config/default | $(KUBECTL) delete --ignore-not-found=$(ignore-not-found) -f -

.PHONY: deploy-debug
deploy-debug: manifests kustomize ## Deploy controller to the K8s cluster specified in ~/.kube/config.
	cd config/manager && $(KUSTOMIZE) edit set image controller=${IMG}
	$(KUSTOMIZE) build config/overlays/debug | $(KUBECTL) apply -f -

.PHONY: undeploy-debug
undeploy-debug: kustomize ## Undeploy controller from the K8s cluster specified in ~/.kube/config. Call with ignore-not-found=true to ignore resource not found errors during deletion.
	$(KUSTOMIZE) build config/overlays/debug | $(KUBECTL) delete --ignore-not-found=$(ignore-not-found) -f -

.PHONY: deploy-docker
deploy-docker: manifests kustomize ## Deploy controller for DockerProvider to the K8s cluster specified in ~/.kube/config.
	cd config/manager && $(KUSTOMIZE) edit set image controller=${IMG}
	$(KUSTOMIZE) build config/overlays/docker | $(KUBECTL) apply -f -

.PHONY: undeploy-docker
undeploy-docker: kustomize ## Undeploy controller for DockerProvider from the K8s cluster specified in ~/.kube/config. Call with ignore-not-found=true to ignore resource not found errors during deletion.
	$(KUSTOMIZE) build config/overlays/docker | $(KUBECTL) delete --ignore-not-found=$(ignore-not-found) -f -

.PHONY: deploy-docker-debug
deploy-docker-debug: manifests kustomize ## Deploy controller for DockerProvider to the K8s cluster specified in ~/.kube/config.
	cd config/manager && $(KUSTOMIZE) edit set image controller=${IMG}
	$(KUSTOMIZE) build config/overlays/docker-debug | $(KUBECTL) apply -f -

.PHONY: undeploy-docker-debug
undeploy-docker-debug: kustomize ## Undeploy controller for DockerProvider from the K8s cluster specified in ~/.kube/config. Call with ignore-not-found=true to ignore resource not found errors during deletion.
	$(KUSTOMIZE) build config/overlays/docker-debug | $(KUBECTL) delete --ignore-not-found=$(ignore-not-found) -f -

.PHONY: deploy-with-alert-receiver
deploy-with-alert-receiver: manifests kustomize ## Deploy controller with alert-receiver to the K8s cluster specified in ~/.kube/config.
	cd config/manager && $(KUSTOMIZE) edit set image controller=${IMG}
	cd config/alert-receiver/receiver && $(KUSTOMIZE) edit set image alert-receiver=${ALERT_RECEIVER_IMG}
	$(KUSTOMIZE) build config/overlays/alert-receiver | $(KUBECTL) apply -f -

.PHONY: undeploy-with-alert-receiver
undeploy-with-alert-receiver: kustomize ## Undeploy controller with alert-receiver from the K8s cluster specified in ~/.kube/config. Call with ignore-not-found=true to ignore resource not found errors during deletion.
	$(KUSTOMIZE) build config/overlays/alert-receiver | $(KUBECTL) delete --ignore-not-found=$(ignore-not-found) -f -

##@ Dependencies

## Location to install dependencies to
LOCALBIN ?= $(shell pwd)/bin
$(LOCALBIN):
	mkdir -p $(LOCALBIN)

## Tool Binaries
KUBECTL ?= kubectl
KIND ?= kind
KUSTOMIZE ?= $(LOCALBIN)/kustomize
CONTROLLER_GEN ?= $(LOCALBIN)/controller-gen
ENVTEST ?= $(LOCALBIN)/setup-envtest
GOLANGCI_LINT = $(LOCALBIN)/golangci-lint
MOCKGEN ?= $(LOCALBIN)/mockgen

## Tool Versions
KUSTOMIZE_VERSION ?= v5.6.0
CONTROLLER_TOOLS_VERSION ?= v0.17.2
#ENVTEST_VERSION is the version of controller-runtime release branch to fetch the envtest setup script (i.e. release-0.20)
ENVTEST_VERSION ?= $(shell go list -m -f "{{ .Version }}" sigs.k8s.io/controller-runtime | awk -F'[v.]' '{printf "release-%d.%d", $$2, $$3}')
#ENVTEST_K8S_VERSION is the version of Kubernetes to use for setting up ENVTEST binaries (i.e. 1.31)
ENVTEST_K8S_VERSION ?= $(shell go list -m -f "{{ .Version }}" k8s.io/api | awk -F'[v.]' '{printf "1.%d", $$3}')
GOLANGCI_LINT_VERSION ?= v2.1.2
MOCKGEN_VERSION ?= v0.6.0

.PHONY: kustomize
kustomize: $(KUSTOMIZE) ## Download kustomize locally if necessary.
$(KUSTOMIZE): $(LOCALBIN)
	$(call go-install-tool,$(KUSTOMIZE),sigs.k8s.io/kustomize/kustomize/v5,$(KUSTOMIZE_VERSION))

.PHONY: controller-gen
controller-gen: $(CONTROLLER_GEN) ## Download controller-gen locally if necessary.
$(CONTROLLER_GEN): $(LOCALBIN)
	$(call go-install-tool,$(CONTROLLER_GEN),sigs.k8s.io/controller-tools/cmd/controller-gen,$(CONTROLLER_TOOLS_VERSION))

.PHONY: setup-envtest
setup-envtest: envtest ## Download the binaries required for ENVTEST in the local bin directory.
	@echo "Setting up envtest binaries for Kubernetes version $(ENVTEST_K8S_VERSION)..."
	@$(ENVTEST) use $(ENVTEST_K8S_VERSION) --bin-dir $(LOCALBIN) -p path || { \
		echo "Error: Failed to set up envtest binaries for version $(ENVTEST_K8S_VERSION)."; \
		exit 1; \
	}

.PHONY: envtest
envtest: $(ENVTEST) ## Download setup-envtest locally if necessary.
$(ENVTEST): $(LOCALBIN)
	$(call go-install-tool,$(ENVTEST),sigs.k8s.io/controller-runtime/tools/setup-envtest,$(ENVTEST_VERSION))

.PHONY: golangci-lint
golangci-lint: $(GOLANGCI_LINT) ## Download golangci-lint locally if necessary.
$(GOLANGCI_LINT): $(LOCALBIN)
	$(call go-install-tool,$(GOLANGCI_LINT),github.com/golangci/golangci-lint/v2/cmd/golangci-lint,$(GOLANGCI_LINT_VERSION))

.PHONY: mockgen-bin
mockgen-bin: $(MOCKGEN) ## Download mockgen locally if necessary.
$(MOCKGEN): $(LOCALBIN)
	$(call go-install-tool,$(MOCKGEN),go.uber.org/mock/mockgen,$(MOCKGEN_VERSION))

# go-install-tool will 'go install' any package with custom target and name of binary, if it doesn't exist
# $1 - target path with name of binary
# $2 - package url which can be installed
# $3 - specific version of package
define go-install-tool
@[ -f "$(1)-$(3)" ] || { \
set -e; \
package=$(2)@$(3) ;\
echo "Downloading $${package}" ;\
rm -f $(1) || true ;\
GOBIN=$(LOCALBIN) go install $${package} ;\
mv $(1) $(1)-$(3) ;\
} ;\
ln -sf $(1)-$(3) $(1)
endef

## Path to documentations and script names
PATHDOC ?= $(shell pwd)/doc
PATHPUML ?= $(PATHDOC)/puml
PATHHTML ?= $(PATHDOC)/html
SCRIPT_SCALE_CTRL_GOPLANTUML ?= $(PATHDOC)/scale-controller-goplantuml.sh
SCRIPT_SCALE_CTRL_GODOC_WGET ?= $(PATHDOC)/scale-controller-godoc-wget.sh

.PHONY: doc
doc: $(PATHPUML)/scale-controller-class-diagram.puml $(PATHHTML)/scale-controller-godoc ## Generate documentation files
$(PATHPUML)/scale-controller-class-diagram.puml: $(SCRIPT_SCALE_CTRL_GOPLANTUML)
	@cd $(PATHDOC) && $(SCRIPT_SCALE_CTRL_GOPLANTUML)

$(PATHHTML)/scale-controller-godoc: $(SCRIPT_SCALE_CTRL_GODOC_WGET)
	@cd $(PATHDOC) && $(SCRIPT_SCALE_CTRL_GODOC_WGET)

.PHONY: doc-clean
doc-clean: ## Clean up generated documentation files
	@rm -f $(PATHDOC)/pkg
	@test -e $(PATHHTML) && rm -r $(PATHHTML) || true
	@test -e $(PATHPUML) && rm -r $(PATHPUML) || true

## Path to script for create helm chart
PATH_CREATE_HELM_SCRIPT ?= $(shell pwd)/helm/create-helm.sh
.PHONY: manifests-helm
manifests-helm: helmchart ## Generate helm chart.
helmchart: $(PATH_CREATE_HELM_SCRIPT)
	@echo "Regenerating helm chart..."
	$<

MOCK_DIR := test/testutils/mock
$(MOCK_DIR):
	mkdir -p $(MOCK_DIR)
MOCK_FILE_CLIENT := $(MOCK_DIR)/client.go
MOCK_FILE_MANAGER := $(MOCK_DIR)/manager.go
MOCK_FILE_HELPERCLIENT := $(MOCK_DIR)/helperclient.go
MOCK_FILES := $(MOCK_FILE_CLIENT) $(MOCK_FILE_MANAGER) $(MOCK_FILE_HELPERCLIENT)
.PHONY: mockgen
mockgen: mockgen-bin $(MOCK_FILES) ## Generate mock interfaces for unit tests.
$(MOCK_FILE_CLIENT): | $(MOCK_DIR)
	@echo "Generating mock for controller-runtime Client and SubResourceWriter interfaces..."
	@$(MOCKGEN) -destination=$@ -package=mock sigs.k8s.io/controller-runtime/pkg/client Client,SubResourceWriter
$(MOCK_FILE_MANAGER): | $(MOCK_DIR)
	@echo "Generating mock for controller-runtime Manager interface..."
	@$(MOCKGEN) -destination=$@ -package=mock sigs.k8s.io/controller-runtime/pkg/manager Manager
$(MOCK_FILE_HELPERCLIENT): | $(MOCK_DIR)
	@echo "Generating mock for internal/helper/client interfaces..."
	@$(MOCKGEN) -destination=$@ -package=mock -source=internal/helper/client/interfaces.go

.PHONY: mockgen-clean
mockgen-clean: ## Clean up generated mock files.
	@rm -f $(MOCK_FILES)
	@rmdir --ignore-fail-on-non-empty $(MOCK_DIR)
