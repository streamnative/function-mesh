# Current Operator version
VERSION ?= 0.20.0
# Default image tag
DOCKER_REPO := $(if $(DOCKER_REPO),$(DOCKER_REPO),streamnative)
OPERATOR_IMG ?= ${DOCKER_REPO}/function-mesh:v$(VERSION)
OPERATOR_IMG_LATEST ?= ${DOCKER_REPO}/function-mesh:latest
ENVTEST_K8S_VERSION = 1.24.2

# IMAGE_TAG_BASE defines the docker.io namespace and part of the image name for remote images.
# This variable is used to construct full image tags for bundle and catalog images.
#
# For example, running 'make bundle-build bundle-push catalog-build catalog-push' will build and push both
# example.com/memcached-operator-bundle:$VERSION and example.com/memcached-operator-catalog:$VERSION.
IMAGE_TAG_BASE ?= ${DOCKER_REPO}/function-mesh

# BUNDLE_IMG defines the image:tag used for the bundle.
# You can use it as an arg. (E.g make bundle-build BUNDLE_IMG=<some-registry>/<project-name-bundle>:<tag>)
BUNDLE_IMG ?= $(IMAGE_TAG_BASE)-bundle:v$(VERSION)

GOOS := $(if $(GOOS),$(GOOS),linux)
GOARCH := $(if $(GOARCH),$(GOARCH),amd64)
GOENV  := CGO_ENABLED=0 GOOS=$(GOOS) GOARCH=$(GOARCH)
GO     := $(GOENV) go
GO_BUILD := $(GO) build -trimpath
GO_MAJOR_VERSION := $(shell $(GO) version | cut -c 14- | cut -d' ' -f1 | cut -d'.' -f1)
GO_MINOR_VERSION := $(shell $(GO) version | cut -c 14- | cut -d' ' -f1 | cut -d'.' -f2)

# Options for 'bundle-build'
ifneq ($(origin CHANNELS), undefined)
BUNDLE_CHANNELS := --channels=$(CHANNELS)
endif
ifneq ($(origin DEFAULT_CHANNEL), undefined)
BUNDLE_DEFAULT_CHANNEL := --default-channel=$(DEFAULT_CHANNEL)
endif
BUNDLE_METADATA_OPTS ?= $(BUNDLE_CHANNELS) $(BUNDLE_DEFAULT_CHANNEL)

# Image URL to use all building/pushing image targets
IMG ?= ${DOCKER_REPO}/function-mesh-operator:v$(VERSION)
# Produce CRDs that work back to Kubernetes 1.11 (no version conversion)
CRD_OPTIONS ?= "crd:maxDescLen=0,generateEmbeddedObjectMeta=true"

# Get the currently used golang install path (in GOPATH/bin, unless GOBIN is set)
ifeq (,$(shell go env GOBIN))
GOBIN=$(shell go env GOPATH)/bin
else
GOBIN=$(shell go env GOBIN)
endif

BUILD_DATETIME := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")

## Location to install dependencies to
LOCALBIN ?= $(shell pwd)/bin
$(LOCALBIN):
	mkdir -p $(LOCALBIN)

## Tool Binaries
KUSTOMIZE ?= $(LOCALBIN)/kustomize
CONTROLLER_GEN ?= $(LOCALBIN)/controller-gen
ENVTEST ?= $(LOCALBIN)/setup-envtest
YQ ?= $(LOCALBIN)/yq
E2E ?= $(LOCALBIN)/e2e

all: manager

# Run tests
ENVTEST_ASSETS_DIR=$(shell pwd)/testbin
test: generate fmt vet manifests envtest
	KUBEBUILDER_ASSETS="$(shell $(ENVTEST) use $(ENVTEST_K8S_VERSION) -p path)" GOLANG_PROTOBUF_REGISTRATION_CONFLICT=warn go test ./... -coverprofile cover.out

test-ginkgo: generate fmt vet manifests envtest
	KUBEBUILDER_ASSETS="$(shell $(ENVTEST) use $(ENVTEST_K8S_VERSION) -p path)" GOLANG_PROTOBUF_REGISTRATION_CONFLICT=warn go test ./controllers/ -v -ginkgo.v

.PHONY: envtest
envtest:
	test -s $(LOCALBIN)/setup-envtest || GOBIN=$(LOCALBIN) go install sigs.k8s.io/controller-runtime/tools/setup-envtest@v0.0.0-20240320141353-395cfc7486e6

# Build manager binary
manager: generate fmt vet
	$(GO_BUILD) -o bin/function-mesh-controller-manager main.go

# Run against the configured Kubernetes cluster in ~/.kube/config
run: generate fmt vet manifests
	ENABLE_WEBHOOKS=false go run ./main.go

# Install CRDs into a cluster
install: manifests kustomize crd
	kubectl apply -f manifests/crd.yaml

# Uninstall CRDs from a cluster
uninstall: manifests kustomize
	$(KUSTOMIZE) build config/crd | kubectl delete -f -

# Deploy controller in the configured Kubernetes cluster in ~/.kube/config
deploy: manifests kustomize
	cd config/manager && $(KUSTOMIZE) edit set image controller=${IMG}
	$(KUSTOMIZE) build config/default | kubectl apply -f -

# Generate manifests e.g. CRD, RBAC etc.
manifests: controller-gen
	$(CONTROLLER_GEN) $(CRD_OPTIONS) rbac:roleName=manager-role webhook paths="./..." output:crd:artifacts:config=config/crd/bases

helm-crds: manifests yq kustomize
	chmod +x ./hack/gen-helm-crd-templates.sh
	./hack/gen-helm-crd-templates.sh $(KUSTOMIZE) $(YQ)

# Run go fmt against code
fmt:
	go fmt ./...

# Run go vet against code
vet:
	go vet ./...

# Generate code
generate: controller-gen
	$(CONTROLLER_GEN) object:headerFile="hack/boilerplate.go.txt" paths="./..."

# Build the docker image
docker-build: test
	docker build --platform linux/amd64 . -t ${IMG}

# Build image for red hat certification
docker-build-redhat:
	docker build --platform linux/amd64 -f redhat.Dockerfile . -t ${IMG} --build-arg VERSION=${VERSION} --no-cache

# Push the docker image
image-push:
	docker push ${IMG}

# find or download controller-gen
# download controller-gen if necessary
controller-gen:
	$(call go-get-tool,$(CONTROLLER_GEN),sigs.k8s.io/controller-tools/cmd/controller-gen@v0.15.0)

kustomize: ## Download kustomize locally if necessary.
	$(call go-get-tool,$(KUSTOMIZE),sigs.k8s.io/kustomize/kustomize/v4@v4.5.5)

yq: ## Download yq locally if necessary.
	$(call go-get-tool,$(YQ),github.com/mikefarah/yq/v4@latest)

skywalking-e2e: ## Download e2e locally if necessary.
	$(call go-get-tool,$(E2E),github.com/apache/skywalking-infra-e2e/cmd/e2e@v1.2.0)

# Generate bundle manifests and metadata, then validate generated files.
.PHONY: bundle
bundle: yq kustomize manifests
	operator-sdk generate kustomize manifests -q
	cd config/manager && $(KUSTOMIZE) edit set image controller=$(IMG)
	$(KUSTOMIZE) build config/manifests | operator-sdk generate bundle -q --overwrite --version $(VERSION) $(BUNDLE_METADATA_OPTS)
	$(YQ) eval -i ".metadata.annotations.\"olm.skipRange\" = \"<$(VERSION)\"" bundle/manifests/function-mesh.clusterserviceversion.yaml
	$(YQ) eval -i ".metadata.annotations.createdAt = \"$(BUILD_DATETIME)\"" bundle/manifests/function-mesh.clusterserviceversion.yaml
	$(YQ) eval -i ".metadata.annotations.containerImage = \"$(OPERATOR_IMG)\"" bundle/manifests/function-mesh.clusterserviceversion.yaml
	operator-sdk bundle validate ./bundle

# Build the bundle image.
.PHONY: bundle-build
bundle-build:
	docker build --platform linux/amd64 -f bundle.Dockerfile -t $(BUNDLE_IMG) .

.PHONY: bundle-push
bundle-push: ## Push the bundle image.
	echo $(BUNDLE_IMG)
	$(MAKE) image-push IMG=$(BUNDLE_IMG)

crd: manifests
	$(KUSTOMIZE) build config/crd > manifests/crd.yaml

rbac: manifests
	$(KUSTOMIZE) build config/rbac > manifests/rbac.yaml

release: manifests kustomize crd rbac manager operator-docker-image helm-crds

operator-docker-image: manager test
	docker build --platform linux/amd64 -f operator.Dockerfile -t $(OPERATOR_IMG) .
	docker tag $(OPERATOR_IMG) $(OPERATOR_IMG_LATEST)

docker-push:
	docker push $(OPERATOR_IMG)
	docker push $(OPERATOR_IMG_LATEST)

.PHONY: opm
OPM = ./bin/opm
opm: ## Download opm locally if necessary.
ifeq (,$(wildcard $(OPM)))
ifeq (,$(shell which opm 2>/dev/null))
	@{ \
	set -e ;\
	mkdir -p $(dir $(OPM)) ;\
	OS=$(shell go env GOOS) && ARCH=$(shell go env GOARCH) && \
	curl -sSLo $(OPM) https://github.com/operator-framework/operator-registry/releases/download/v1.15.3/$${OS}-$${ARCH}-opm ;\
	chmod +x $(OPM) ;\
	}
else
OPM = $(shell which opm)
endif
endif

# A comma-separated list of bundle images (e.g. make catalog-build BUNDLE_IMGS=example.com/operator-bundle:v0.1.0,example.com/operator-bundle:v0.2.0).
# These images MUST exist in a registry and be pull-able.
BUNDLE_IMGS ?= $(BUNDLE_IMG)

# The image tag given to the resulting catalog image (e.g. make catalog-build CATALOG_IMG=example.com/operator-catalog:v0.2.0).
CATALOG_IMG ?= $(IMAGE_TAG_BASE)-catalog:v$(VERSION)

ifneq ($(origin CATALOG_BRANCH_TAG), undefined)
CATALOG_BRANCH_IMG ?= $(IMAGE_TAG_BASE)-catalog:$(CATALOG_BRANCH_TAG)
endif

# Set CATALOG_BASE_IMG to an existing catalog image tag to add $BUNDLE_IMGS to that image.
ifneq ($(origin CATALOG_BASE_IMG), undefined)
FROM_INDEX_OPT := --from-index $(CATALOG_BASE_IMG)
endif

# Build a catalog image by adding bundle images to an empty catalog using the operator package manager tool, 'opm'.
# This recipe invokes 'opm' in 'semver' bundle add mode. For more information on add modes, see:
# https://github.com/operator-framework/community-operators/blob/7f1438c/docs/packaging-operator.md#updating-your-existing-operator
.PHONY: catalog-build
catalog-build: opm ## Build a catalog image.
	$(OPM) index add --container-tool docker --mode semver --tag $(CATALOG_IMG) --bundles $(BUNDLE_IMGS) $(FROM_INDEX_OPT)
ifneq ($(origin CATALOG_BRANCH_TAG), undefined)
	docker tag $(CATALOG_IMG) $(CATALOG_BRANCH_IMG)
endif

# Push the catalog image.
.PHONY: catalog-push
catalog-push: ## Push a catalog image.
	$(MAKE) image-push IMG=$(CATALOG_IMG)
ifneq ($(origin CATALOG_BRANCH_TAG), undefined)
	$(MAKE) image-push IMG=$(CATALOG_BRANCH_IMG)
endif

version:
	@echo ${VERSION}

operator-docker-image-name:
	@echo ${OPERATOR_IMG}

function-mesh-docker-image-name:
	@echo ${IMG}

# Build the docker image without tests
docker-build-skip-test:
	docker build --platform linux/amd64 . -t ${IMG}

e2e: skywalking-e2e yq
	$(E2E) run -c .ci/tests/integration/e2e.yaml

# go-get-tool will 'go get' any package $2 and install it to $1.
PROJECT_DIR := $(shell dirname $(abspath $(lastword $(MAKEFILE_LIST))))
define go-get-tool
	@[ -f $(1) ] || { \
	set -e ;\
	echo "Installing $(2)" ;\
	if [ "$(GO_MINOR_VERSION)" -ge "17" ]; then \
		GOBIN=$(PROJECT_DIR)/bin go install -v $(2) ;\
	else \
		CONTROLLER_GEN_TMP_DIR=$$(mktemp -d) ;\
		cd $$CONTROLLER_GEN_TMP_DIR ;\
		go mod init tmp ;\
		GOBIN=$(PROJECT_DIR)/bin go get -v $(2) ;\
		rm -rf $$CONTROLLER_GEN_TMP_DIR ;\
	fi ;\
}
endef

# Generate bundle manifests and metadata, then validate generated files.
.PHONY: redhat-certificated-bundle
redhat-certificated-bundle: yq kustomize manifests
	operator-sdk generate kustomize manifests -q
	cd config/manager && $(KUSTOMIZE) edit set image controller=$(shell docker inspect --format='{{json .RepoDigests}}' $(OPERATOR_IMG) | jq --arg IMAGE_TAG_BASE "$(IMAGE_TAG_BASE)" -c '.[] | select(index($$IMAGE_TAG_BASE))' -r)
	$(KUSTOMIZE) build config/manifests | operator-sdk generate bundle -q --overwrite --version $(VERSION) $(BUNDLE_METADATA_OPTS)
	$(YQ) eval -i ".metadata.annotations.\"olm.skipRange\" = \"<$(VERSION)\"" bundle/manifests/function-mesh.clusterserviceversion.yaml
	#$(YQ) eval -i ".metadata.annotations.\"olm.properties\" = ([{\"type\": \"olm.maxOpenShiftVersion\", \"value\": \"4.13\"}] | @json)" bundle/manifests/function-mesh.clusterserviceversion.yaml
	$(YQ) eval -i ".metadata.annotations.createdAt = \"$(BUILD_DATETIME)\"" bundle/manifests/function-mesh.clusterserviceversion.yaml
	$(YQ) eval -i ".metadata.annotations.containerImage = \"$(shell docker inspect --format='{{json .RepoDigests}}' $(OPERATOR_IMG) | jq --arg IMAGE_TAG_BASE "$(IMAGE_TAG_BASE)" -c '.[] | select(index($$IMAGE_TAG_BASE))' -r)\"" bundle/manifests/function-mesh.clusterserviceversion.yaml
	$(YQ) eval -i '.annotations += {"operators.operatorframework.io.bundle.channel.default.v1":"alpha"}' bundle/metadata/annotations.yaml
	IMG_DIGEST=$(shell docker inspect --format='{{json .RepoDigests}}' $(OPERATOR_IMG) | jq --arg IMAGE_TAG_BASE "$(IMAGE_TAG_BASE)" -c '.[] | select(index($$IMAGE_TAG_BASE))' -r) hack/postprocess-bundle.sh
	operator-sdk bundle validate ./bundle --select-optional name=operatorhub
	operator-sdk bundle validate ./bundle --select-optional suite=operatorframework

# Build the bundle image.
.PHONY: redhat-certificated-bundle-build
redhat-certificated-bundle-build:
	docker build --platform linux/amd64 -f bundle.Dockerfile -t $(BUNDLE_IMG) .

.PHONY: redhat-certificated-bundle-push
redhat-certificated-bundle-push: ## Push the bundle image.
	echo $(BUNDLE_IMG)
	$(MAKE) image-push IMG=$(BUNDLE_IMG)

# Build the bundle image.
.PHONY: redhat-certificated-image-build
redhat-certificated-image-build:
	docker build --platform linux/amd64 -f redhat.Dockerfile . -t ${OPERATOR_IMG} --build-arg VERSION=${VERSION} --no-cache

.PHONY: redhat-certificated-image-push
redhat-certificated-image-push: ## Push the bundle image.
	echo $(OPERATOR_IMG)
	$(MAKE) image-push IMG=$(OPERATOR_IMG)

##@ Generate the metrics documentation
.PHONY: generate-metricsdocs
generate-metricsdocs:
	mkdir -p $(shell pwd)/docs/monitoring
	go run -ldflags="${LDFLAGS}" ./pkg/monitoring/metricsdocs > docs/monitoring/metrics.md