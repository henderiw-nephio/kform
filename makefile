# Get the currently used golang install path (in GOPATH/bin, unless GOBIN is set)
ifeq (,$(shell go env GOBIN))
GOBIN=$(shell go env GOPATH)/bin
else
GOBIN=$(shell go env GOBIN)
endif

# Setting SHELL to bash allows bash commands to be executed by recipes.
# This is a requirement for 'setup-envtest.sh' in the test target.
# Options are set to exit when a recipe line exits non-zero or a piped command fails.
SHELL = /usr/bin/env bash -o pipefail
.SHELLFLAGS = -ec


.PHONY: generate
generate: mdtogo controller-gen
	rm -rf internal/docs/generated
	mkdir -p internal/docs/generated
	GOBIN=$(LOCALBIN) go generate ./...
	go fmt ./internal/docs/generated/...

.PHONY: manifests
manifests: controller-gen ## Generate WebhookConfiguration, ClusterRole and CustomResourceDefinition objects.
	$(CONTROLLER_GEN) rbac:roleName=manager-role crd webhook paths="./providers/provider-kubernetes/..." output:crd:artifacts:config=./providers/provider-kubernetes/crd
	$(CONTROLLER_GEN) rbac:roleName=manager-role crd webhook paths="./providers/provider-resourcebackend/..." output:crd:artifacts:config=./providers/provider-resourcebackend/crd

.PHONY: fmt
fmt: ## Run go fmt against code.
	go fmt ./...

.PHONY: vet
vet: ## Run go vet against code.
	go vet ./...

.PHONY: all
all: generate fmt vet ## Build manager binary.
	go build -ldflags "-X github.com/henderiw-nephio/k8sform/tools/cmd/kform/commands.version=${GIT_COMMIT}" -o $(LOCALBIN)/kform -v tools/cmd/kform/main.go
	go build -o $(LOCALBIN)/provider-kubernetes -v providers/provider-kubernetes/main.go
	go build -o $(LOCALBIN)/provider-resourcebackend -v providers/provider-resourcebackend/main.go
	##$(CONTROLLER_GEN) object:headerFile="hack/boilerplate.go.txt" paths="./..."

##@ Build Dependencies

## Location to install dependencies to
LOCALBIN ?= $(shell pwd)/bin
$(LOCALBIN):
	mkdir -p $(LOCALBIN)

## Tool Binaries
MDTOGO ?= $(LOCALBIN)/mdtogo
CONTROLLER_GEN ?= $(LOCALBIN)/controller-gen

CONTROLLER_TOOLS_VERSION ?= v0.13.0

.PHONY: controller-gen
controller-gen: $(CONTROLLER_GEN) ## Download controller-gen locally if necessary.
$(CONTROLLER_GEN): $(LOCALBIN)
	test -s $(LOCALBIN)/controller-gen || GOBIN=$(LOCALBIN) go install sigs.k8s.io/controller-tools/cmd/controller-gen@$(CONTROLLER_TOOLS_VERSION)

.PHONY: mdtogo
mdtogo: $(MDTOGO) ## Install mdtgo locallt
$(MDTOGO): $(LOCALBIN)
	GOBIN=$(LOCALBIN) go install ./mdtogo