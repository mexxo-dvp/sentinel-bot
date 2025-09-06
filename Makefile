# =========================
# sentinel-bot — unified Makefile (OCI-ready)
# =========================

# ---- App/Repo ----
APP_NAME ?= sentinel-bot
OWNER ?= $(shell echo $${GITHUB_REPOSITORY_OWNER:-mexxo-dvp})
REGISTRY ?= ghcr.io
REPOSITORY ?= $(OWNER)/$(APP_NAME)
IMAGE ?= $(REGISTRY)/$(REPOSITORY)

# ---- Platform for container image ----
OS ?= linux
ARCH ?= amd64

# ---- Versioning (git tag + short sha) ----
GIT_SHA := $(shell git rev-parse --short=7 HEAD)
BASE_TAG ?= $(shell git describe --tags --abbrev=0 --match 'v*' 2>/dev/null)
ifeq ($(strip $(BASE_TAG)),)
BASE_TAG := v1.0.0
endif
TAG := $(BASE_TAG)-$(GIT_SHA)
FULL_TAG := $(TAG)-$(OS)-$(ARCH)

# Chart semver (без префікса 'v' — потрібно для Helm)
CHART_VER := $(patsubst v%,%,$(BASE_TAG))

# ---- Build flags ----
LD_FLAGS := -X github.com/mexxo-dvp/sentinel-bot/cmd.appVersion=$(BASE_TAG)

# ---- Helm / tools ----
HELM_DIR ?= helm/$(APP_NAME)
YQ ?= yq
HELM ?= helm

# ---- Helm OCI registry ----
HELM_REGISTRY ?= ghcr.io
HELM_CHARTS_REPO ?= $(HELM_REGISTRY)/$(OWNER)/charts
DIST_DIR ?= _dist
CHART_TGZ ?= $(DIST_DIR)/$(APP_NAME)-$(CHART_VER).tgz

SHELL := /bin/bash

.PHONY: help print all \
        linux arm64 macos windows \
        image image-local push \
        helm-bump helm-oci-login helm-package helm-push helm-release \
        helm-clean clean

# -------------------------
# Helpers
# -------------------------
help:
	@echo "Targets:"
	@echo "  linux         Build Go binary for linux/amd64 -> bin/$(APP_NAME)-linux-amd64"
	@echo "  arm64         Build Go binary for linux/arm64 -> bin/$(APP_NAME)-linux-arm64"
	@echo "  macos         Build Go binaries for darwin (amd64, arm64)"
	@echo "  windows       Build Go binary for windows/amd64"
	@echo "  image         Build+Push container to $(IMAGE):$(FULL_TAG) and :develop (buildx)"
	@echo "  image-local   Local docker build (no push), tag :local"
	@echo "  helm-bump     Update Helm values (image/top-level keys) + Chart.appVersion/version"
	@echo "  helm-package  helm package -> $(DIST_DIR)/$(APP_NAME)-$(CHART_VER).tgz"
	@echo "  helm-push     helm push OCI -> oci://$(HELM_CHARTS_REPO)"
	@echo "  helm-release  helm-package + helm-push"
	@echo "  print         Show resolved vars"
	@echo "  clean         Remove bin/ and optional local image"
	@echo "  helm-clean    Remove $(DIST_DIR)/"

print:
	@echo "OWNER      = $(OWNER)"
	@echo "IMAGE      = $(IMAGE)"
	@echo "BASE_TAG   = $(BASE_TAG)"
	@echo "CHART_VER  = $(CHART_VER)"
	@echo "GIT_SHA    = $(GIT_SHA)"
	@echo "TAG        = $(TAG)"
	@echo "FULL_TAG   = $(FULL_TAG)"
	@echo "OS/ARCH    = $(OS)/$(ARCH)"
	@echo "HELM_REPO  = oci://$(HELM_CHARTS_REPO)"

all: image helm-bump helm-release

# -------------------------
# Go binaries (retain your originals)
# -------------------------
linux:
	GOOS=linux GOARCH=amd64 go build -ldflags="$(LD_FLAGS)" -o bin/$(APP_NAME)-linux-amd64 main.go

arm64:
	GOOS=linux GOARCH=arm64 go build -ldflags="$(LD_FLAGS)" -o bin/$(APP_NAME)-linux-arm64 main.go

macos:
	GOOS=darwin GOARCH=amd64 go build -ldflags="$(LD_FLAGS)" -o bin/$(APP_NAME)-darwin-amd64 main.go

windows:
	GOOS=windows GOARCH=amd64 go build -ldflags="$(LD_FLAGS)" -o bin/$(APP_NAME)-windows-amd64.exe main.go

# -------------------------
# Container image (GHCR)
# -------------------------
# Local test build (no push)
image-local:
	docker build \
		--build-arg VERSION=$(BASE_TAG) \
		-t $(IMAGE):local \
		.

# Buildx build + push (linux/amd64) with required tags
image:
	docker buildx build \
		--platform $(OS)/$(ARCH) \
		--build-arg VERSION=$(BASE_TAG) \
		--tag $(IMAGE):$(FULL_TAG) \
		--tag $(IMAGE):$(TAG) \
		--tag $(IMAGE):develop \
		--push \
		.

# Kept for compatibility; push is already done in `image`
push:
	@echo "Image already pushed in 'image' target."

# -------------------------
# Helm bump (values + Chart) — без commit/push (це робить CI)
# -------------------------
helm-bump:
	@if ! command -v $(YQ) >/dev/null 2>&1; then \
	  echo "ERROR: 'yq' is required. Install: https://github.com/mikefarah/yq"; \
	  exit 1; \
	fi
	# sync image.* in values.yaml
	$(YQ) -i '.image.registry = "$(REGISTRY)"'                $(HELM_DIR)/values.yaml || true
	$(YQ) -i '.image.repository = "$(REGISTRY)/$(REPOSITORY)"' $(HELM_DIR)/values.yaml
	$(YQ) -i '.image.tag = "$(TAG)"'                           $(HELM_DIR)/values.yaml
	$(YQ) -i '.image.os = "$(OS)"'                             $(HELM_DIR)/values.yaml
	$(YQ) -i '.image.arch = "$(ARCH)"'                         $(HELM_DIR)/values.yaml
	# sync top-level keys required by autotest (keep if present)
	$(YQ) -i '.repository = "$(REGISTRY)/$(REPOSITORY)"'       $(HELM_DIR)/values.yaml
	$(YQ) -i '.tag = "$(TAG)"'                                 $(HELM_DIR)/values.yaml
	$(YQ) -i '.os = "$(OS)"'                                   $(HELM_DIR)/values.yaml
	$(YQ) -i '.arch = "$(ARCH)"'                               $(HELM_DIR)/values.yaml
	# Chart.yaml: appVersion keeps 'v*', version is semver without 'v'
	$(YQ) -i '.appVersion = "$(BASE_TAG)"'                     $(HELM_DIR)/Chart.yaml
	$(YQ) -i '.version = "$(CHART_VER)"'                       $(HELM_DIR)/Chart.yaml
	@echo "Helm values & Chart updated locally."

# -------------------------
# Helm OCI publish (local dev) — login/package/push
# -------------------------
helm-oci-login:
	@if ! command -v $(HELM) >/dev/null 2>&1; then \
	  echo "ERROR: 'helm' is required."; exit 1; fi
	@echo "Login to OCI registry: $(HELM_REGISTRY)"
	@echo "$${GITHUB_TOKEN:-$${GH_TOKEN:-}} " | \
	$(HELM) registry login $(HELM_REGISTRY) --username $(OWNER) --password-stdin

helm-package:
	@if ! command -v $(HELM) >/dev/null 2>&1; then \
	  echo "ERROR: 'helm' is required."; exit 1; fi
	mkdir -p $(DIST_DIR)
	$(HELM) package $(HELM_DIR) --destination $(DIST_DIR)
	@ls -l $(DIST_DIR) | sed 's/^/  /'

helm-push: helm-oci-login helm-package
	$(HELM) push $(CHART_TGZ) oci://$(HELM_CHARTS_REPO)

helm-release: helm-bump helm-push
	@echo "✅ Helm chart published to oci://$(HELM_CHARTS_REPO) as $(APP_NAME)-$(CHART_VER).tgz"

# -------------------------
# Cleanup
# -------------------------
helm-clean:
	rm -rf $(DIST_DIR)

clean:
	rm -rf bin/
	-docker rmi $(IMAGE):local 2>/dev/null || true
