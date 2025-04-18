# Copyright 2022-2025 The nagare media authors
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

# Options

OS             ?= $(HOST_OS)
ARCH           ?= $(HOST_ARCH)
VERSION        ?= dev

GIT_COMMIT     ?= $(shell git rev-parse --short HEAD || echo "unknown")
GIT_TREE_STATE ?= $(shell sh -c 'if test -z "$$(git status --porcelain 2>/dev/null)"; then echo clean; else echo dirty; fi')
BUILD_DATE     ?= $(shell date -u +"%Y-%m-%dT%TZ")

IMAGE_REGISTRY  ?= $(shell cat build/package/image/IMAGE_REGISTRY)
IMAGE_TAG       ?= $(VERSION)
IMAGE_PLATFORMS ?= # by default only ${OS}/${ARCH} is built
BUILDX_OUTPUT   ?= "--load"

TESTENV_K8S_VERSION                   ?= 1.32.0
TESTENV_INGRESS_NGINX_VERSION         ?= 4.11.3 # Helm chart versions
TESTENV_CERT_MANAGER_VERSION          ?= 1.16.1 # Helm chart versions
TESTENV_KUBE_PROMETHEUS_STACK_VERSION ?= 66.2.1 # Helm chart versions
TESTENV_TEMPO_VERSION                 ?= 1.14.0 # Helm chart versions
TESTENV_MINIO_VERSION                 ?= 5.3.0  # Helm chart versions
TESTENV_NATS_VERSION                  ?= 1.2.6  # Helm chart versions

CONTROLLER_TOOLS_VERSION ?= v0.17.2
KUSTOMIZE_VERSION        ?= v5.6.0
YQ_VERSION               ?= v4.45.1
KUSTOMIZE_INSTALL_SCRIPT ?= "https://raw.githubusercontent.com/kubernetes-sigs/kustomize/master/hack/install_kustomize.sh"

# Do not change
HOST_OS     = $(shell which go >/dev/null 2>&1 && go env GOOS)
HOST_ARCH   = $(shell which go >/dev/null 2>&1 && go env GOARCH)
GOVERSION   = $(shell awk '/^go/ { print $$2 }' go.mod)
PKG         = $(shell awk '/^module/ { print $$2 }' go.mod)
CMDS        = $(shell find ./cmd/ -maxdepth 1 -mindepth 1 -type d -exec basename {} \;)
IMAGES      = $(shell find ./build/package/image -maxdepth 1 -mindepth 1 -type d -exec basename {} \;)
SHELL       = /usr/bin/env bash -o pipefail
.SHELLFLAGS = -ec

# Targets

.DEFAULT_GOAL:=help

##@ General

.PHONY: help
help: ## Print this help
	@awk 'BEGIN                      { FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n" } \
				/^[a-zA-Z_0-9-]+:.*?##/    { printf "  \033[36m%-42s\033[0m %s\n", $$1, $$2 } \
				/^## [a-zA-Z_0-9-]+:.*?##/ { printf "  \033[36m%-42s\033[0m %s\n", substr($$1, 4), $$2 } \
				/^##@/                     { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' \
				$(MAKEFILE_LIST)

info: ## Print options
	@printf "\n"
	@printf "\033[1m%s\033[0m\n"          "Build"
	@printf "  \033[36m%-15s\033[0m %s\n"   "OS"             "$(OS)"
	@printf "  \033[36m%-15s\033[0m %s\n"   "ARCH"           "$(ARCH)"
	@printf "\n"
	@printf "\033[1m%s\033[0m\n"          "Version Info"
	@printf "  \033[36m%-15s\033[0m %s\n"   "VERSION"        "$(VERSION)"
	@printf "  \033[36m%-15s\033[0m %s\n"   "GIT_COMMIT"     "$(GIT_COMMIT)"
	@printf "  \033[36m%-15s\033[0m %s\n"   "GIT_TREE_STATE" "$(GIT_TREE_STATE)"
	@printf "  \033[36m%-15s\033[0m %s\n"   "BUILD_DATE"     "$(BUILD_DATE)"
	@printf "\n"
	@printf "\033[1m%s\033[0m\n"          "Container Image"
	@printf "  \033[36m%-15s\033[0m %s\n"   "IMAGE_REGISTRY" "$(IMAGE_REGISTRY)"
	@printf "  \033[36m%-15s\033[0m %s\n"   "IMAGE_TAG"      "$(IMAGE_TAG)"
	@printf "\n"
	@printf "\033[1m%s\033[0m\n"          "Test Environment"
	@printf "  \033[36m%-40s\033[0m %s\n"   "TESTENV_K8S_VERSION"                   "$(TESTENV_K8S_VERSION)"
	@printf "  \033[36m%-40s\033[0m %s\n"   "TESTENV_INGRESS_NGINX_VERSION"         "$(TESTENV_INGRESS_NGINX_VERSION)"
	@printf "  \033[36m%-40s\033[0m %s\n"   "TESTENV_CERT_MANAGER_VERSION"          "$(TESTENV_CERT_MANAGER_VERSION)"
	@printf "  \033[36m%-40s\033[0m %s\n"   "TESTENV_KUBE_PROMETHEUS_STACK_VERSION" "$(TESTENV_KUBE_PROMETHEUS_STACK_VERSION)"
	@printf "  \033[36m%-40s\033[0m %s\n"   "TESTENV_TEMPO_VERSION"                 "$(TESTENV_TEMPO_VERSION)"
	@printf "  \033[36m%-40s\033[0m %s\n"   "TESTENV_MINIO_VERSION"                 "$(TESTENV_MINIO_VERSION)"
	@printf "  \033[36m%-40s\033[0m %s\n"   "TESTENV_NATS_VERSION"                  "$(TESTENV_NATS_VERSION)"
	@printf "\n"
	@printf "\033[1m%s\033[0m\n"          "Build Dependencies"
	@printf "  \033[36m%-25s\033[0m %s\n"   "CONTROLLER_TOOLS_VERSION"      "$(CONTROLLER_TOOLS_VERSION)"
	@printf "  \033[36m%-25s\033[0m %s\n"   "KUSTOMIZE_VERSION"             "$(KUSTOMIZE_VERSION)"
	@printf "  \033[36m%-25s\033[0m %s\n"   "YQ_VERSION"                    "$(YQ_VERSION)"

##@ Development

.PHONY: generate
generate: generate-modules generate-manifests generate-go-deepcopy ## Generate all

.PHONY: generate-modules
generate-modules: ## Generate Go modules files
	@scripts/exec-local generate-modules

.PHONY: generate-manifests
generate-manifests: controller-gen yq ## Generate manifests (CRD, RBAC, etc.)
	@	CONTROLLER_GEN=$(CONTROLLER_GEN) \
		YQ=$(YQ) \
	scripts/exec-local generate-manifests

.PHONY: generate-go-deepcopy
generate-go-deepcopy: controller-gen ## Generate deep copy Go implementation
	@	CONTROLLER_GEN=$(CONTROLLER_GEN) \
	scripts/exec-local generate-go-deepcopy

.PHONY: fmt
fmt: ## Run go fmt against code
	@scripts/exec-local fmt

.PHONY: vet
vet: ## Run go vet against code
	@scripts/exec-local vet

.PHONY: kind-up
kind-up: ## Start the kind test cluster
	@	TESTENV_K8S_VERSION=$(TESTENV_K8S_VERSION) \
	TESTENV_INGRESS_NGINX_VERSION=$(TESTENV_INGRESS_NGINX_VERSION) \
	TESTENV_CERT_MANAGER_VERSION=$(TESTENV_CERT_MANAGER_VERSION) \
	TESTENV_KUBE_PROMETHEUS_STACK_VERSION=$(TESTENV_KUBE_PROMETHEUS_STACK_VERSION) \
	TESTENV_TEMPO_VERSION=$(TESTENV_TEMPO_VERSION) \
	TESTENV_MINIO_VERSION=$(TESTENV_MINIO_VERSION) \
	TESTENV_NATS_VERSION=$(TESTENV_NATS_VERSION) \
	scripts/exec-local kind-up

.PHONY: kind-down
kind-down: ## Stop the kind test cluster
	@scripts/exec-local kind-down

.PHONY: skaffold
skaffold: kind-up ## Execute skaffold command against kind test cluster
	@	ARCH="$(ARCH)" \
		VERSION="$(VERSION)" \
		GOVERSION="$(GOVERSION)" \
		ARGS="$(ARGS)" \
	scripts/exec-local skaffold

.PHONY: skaffold-run
skaffold-run: ## Execute skaffold run pipeline against kind test cluster
	@$(MAKE) skaffold ARGS="run --tail --no-prune=false --cache-artifacts=false --watch-image='none'"

.PHONY: skaffold-debug
skaffold-debug: ## Execute skaffold debug pipeline against kind test cluster
	@$(MAKE) skaffold ARGS="debug --tail --no-prune=false --cache-artifacts=false --watch-image='none'"

##@ Build

.PHONY: build
build: $(addprefix build-, $(CMDS)) ## Build all binaries

## build-functions:                      ## Build functions binary
## build-gateway-nbmp:                   ## Build NBMP gateway binary 
## build-task-shim:                      ## Build task shim binary
## build-workflow-manager:               ## Build workflow manager binary
## build-workflow-manager-helper:        ## Build workflow manager helper binary
## build-workflow-opentelemetry-adapter: ## Build workflow OpenTelemetry adapter binary
## build-workflow-vacuum:                ## Build workflow vacuum binary
build-%: generate-modules generate-go-deepcopy fmt vet
	@	CMD="$*" \
		PKG="$(PKG)" \
		OS="$(OS)" \
		ARCH="$(ARCH)" \
		VERSION="$(VERSION)" \
		GIT_COMMIT="$(GIT_COMMIT)" \
		GIT_TREE_STATE="$(GIT_TREE_STATE)" \
		BUILD_DATE="$(BUILD_DATE)" \
	scripts/exec-local build

.PHONY: output
output: output-all output-crds output-deployment ## Generate all outputs

.PHONY: output-all
output-all: kustomize generate-manifests ## Output all manifests
	@	KUSTOMIZE=$(KUSTOMIZE) \
	scripts/exec-local output-all

.PHONY: output-crds
output-crds: yq output-all ## Output CRD manifests
	@	YQ=$(YQ) \
	scripts/exec-local output-crds

.PHONY: output-deployment
output-deployment: yq output-all ## Output deployment manifests
	@	YQ=$(YQ) \
	scripts/exec-local output-deployment

.PHONY: clean
clean: ## Cleanup build output
	@scripts/exec-local clean

##@ Container Image

.PHONY: image
image: $(addprefix image-, $(IMAGES)) ## Build all container images

## image-function-data-buffer-fs:             ## Build data-buffer-fs function container image
## image-function-data-copy:                  ## Build data-copy function container image
## image-function-data-discard:               ## Build data-discard function container image
## image-function-generic-noop:               ## Build generic-noop function container image
## image-function-generic-sleep:              ## Build generic-sleep function container image
## image-function-media-encode:               ## Build media-encode function container image
## image-function-media-generate-testpattern: ## Build media-generate-testpattern function container image
## image-function-media-merge:                ## Build media-merge function container image
## image-function-media-metadata-technical:   ## Build media-metadata-technical function container image
## image-function-media-package-hls:          ## Build media-package-hls function container image
## image-function-script-lua:                 ## Build script-lua function container image
## image-gateway-nbmp:                        ## Build NBMP gateway container image
## image-workflow-manager:                    ## Build workflow manager container image
## image-workflow-manager-helper:             ## Build workflow manager helper container image
## image-workflow-opentelemetry-adapter:      ## Build workflow OpenTelemetry adapter container image
## image-workflow-vacuum:                     ## Build workflow vacuum container image
image-%:
	@	IMAGE="$*" \
		IMAGE_REGISTRY="$(IMAGE_REGISTRY)" \
		IMAGE_TAG="$(IMAGE_TAG)" \
		GOVERSION="$(GOVERSION)" \
		OS="$(OS)" \
		ARCH="$(ARCH)" \
		PLATFORMS="$(IMAGE_PLATFORMS)" \
		VERSION="$(VERSION)" \
		GIT_COMMIT="$(GIT_COMMIT)" \
		GIT_TREE_STATE="$(GIT_TREE_STATE)" \
		BUILD_DATE="$(BUILD_DATE)" \
		BUILDX_OUTPUT="$(BUILDX_OUTPUT)" \
	scripts/exec-local image

##@ Deployment

.PHONY: install
install: output-crds ## Install CRDs
	kubectl apply --server-side -f out/crds.yaml

.PHONY: uninstall
uninstall: output-crds ## Uninstall CRDs
	kubectl delete --ignore-not-found=true -f out/crds.yaml

.PHONY: deploy
deploy: output-deployment ## Deploy application
	kubectl apply --server-side -f out/deploy.yaml

.PHONY: undeploy
undeploy: output-deployment ## Undeploy application
	kubectl delete --ignore-not-found=true -f out/deploy.yaml

##@ Build Dependencies

LOCALBIN       ?= $(shell pwd)/tmp
CONTROLLER_GEN ?= $(LOCALBIN)/controller-gen
ENVTEST        ?= $(LOCALBIN)/setup-envtest
KUSTOMIZE      ?= $(LOCALBIN)/kustomize
YQ             ?= $(LOCALBIN)/yq

$(LOCALBIN):
	mkdir -p $(LOCALBIN)

.PHONY: controller-gen
controller-gen: $(CONTROLLER_GEN) ## Download controller-gen locally if necessary
$(CONTROLLER_GEN): $(LOCALBIN)
	@  test -s $(LOCALBIN)/controller-gen \
	&& $(LOCALBIN)/controller-gen --version | grep -q $(CONTROLLER_TOOLS_VERSION) \
	|| GOBIN=$(LOCALBIN) go install sigs.k8s.io/controller-tools/cmd/controller-gen@$(CONTROLLER_TOOLS_VERSION)

.PHONY: envtest
envtest: $(ENVTEST) ## Download envtest-setup locally if necessary
$(ENVTEST): $(LOCALBIN)
	@  test -s $(LOCALBIN)/setup-envtest \
	|| GOBIN=$(LOCALBIN) go install sigs.k8s.io/controller-runtime/tools/setup-envtest@latest

.PHONY: kustomize
kustomize: $(KUSTOMIZE) ## Download kustomize locally if necessary
$(KUSTOMIZE): $(LOCALBIN)
	@if test -x $(LOCALBIN)/kustomize && ! $(LOCALBIN)/kustomize version | grep -q $(KUSTOMIZE_VERSION); then \
		echo "$(LOCALBIN)/kustomize version is not expected $(KUSTOMIZE_VERSION). Removing it before installing."; \
		rm -rf $(LOCALBIN)/kustomize; \
	fi
	@  test -s $(LOCALBIN)/kustomize \
	|| { curl -Ss $(KUSTOMIZE_INSTALL_SCRIPT) | bash -s -- $(subst v,,$(KUSTOMIZE_VERSION)) $(LOCALBIN); }

.PHONY: yq
yq: $(YQ) ## Download yq locally if necessary
$(YQ): $(LOCALBIN)
	@  test -s $(LOCALBIN)/yq \
	&& $(LOCALBIN)/yq --version | grep -q $(YQ_VERSION) \
	|| GOBIN=$(LOCALBIN) go install github.com/mikefarah/yq/v4@$(YQ_VERSION)

##@ Examples

## run-sink-pull-http-ffplay:    ## Run example sink HTTP (pull) ffplay
## run-sink-push-http-ffplay:    ## Run example sink HTTP (push) ffplay
## run-source-pull-http-ffmpeg:  ## Run example source HTTP (pull) ffmpeg
## run-source-push-http-ffmpeg:  ## Run example source HTTP (push) ffmpeg
run-%:
	@scripts/exec-local "run-$*"
