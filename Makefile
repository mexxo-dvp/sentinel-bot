# =========================
# kbot — unified Makefile
# =========================

# ---- App/Repo ----
APP_NAME ?= kbot
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
TAG := $(BASE_TAG)-$(GIT_SHA)            # e.g. v1.0.0-106879e
FULL_TAG := $(TAG)-$(OS)-$(ARCH)         # e.g. v1.0.0-106879e-linux-amd64

# ---- Build flags ----
LD_FLAGS := -X github.com/mexxo-dvp/kbot/cmd.appVersion=$(BASE_TAG)

# ---- Helm / tools ----
HELM_DIR ?= helm/$(APP_NAME)
YQ ?= yq

SHELL := /bin/bash

.PHONY: help print all \
        linux arm64 macos windows \
        image image-local push \
        helm-bump clean

# -------------------------
# Helpers
# -------------------------
help:
	@echo "Targets:"
	@echo "  linux        Build Go binary for linux/amd64 -> bin/$(APP_NAME)-linux-amd64"
	@echo "  arm64        Build Go binary for linux/arm64 -> bin/$(APP_NAME)-linux-arm64"
	@echo "  macos        Build Go binaries for darwin (amd64, arm64)"
	@echo "  windows      Build Go binary for windows/amd64"
	@echo "  image        Build+Push container to $(IMAGE):$(FULL_TAG) and :develop (buildx)"
	@echo "  image-local  Local docker build (no push), tag :local"
	@echo "  helm-bump    Update helm values (registry/repository/tag/os/arch) and Chart.appVersion, commit+push to develop"
	@echo "  print        Show resolved vars"
	@echo "  clean        Remove bin/ and optional local image"

print:
	@echo "OWNER      = $(OWNER)"
	@echo "IMAGE      = $(IMAGE)"
	@echo "BASE_TAG   = $(BASE_TAG)"
	@echo "GIT_SHA    = $(GIT_SHA)"
	@echo "TAG        = $(TAG)"
	@echo "FULL_TAG   = $(FULL_TAG)"
	@echo "OS/ARCH    = $(OS)/$(ARCH)"

all: image helm-bump

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
# Helm bump (values + Chart.appVersion) and commit to develop
# -------------------------
helm-bump:
	@if ! command -v $(YQ) >/dev/null 2>&1; then \
	  echo "ERROR: 'yq' is required. Install: https://github.com/mikefarah/yq"; \
	  exit 1; \
	fi
	$(YQ) -i '.image.registry = "$(REGISTRY)"' $(HELM_DIR)/values.yaml
	$(YQ) -i '.image.repository = "$(REPOSITORY)"' $(HELM_DIR)/values.yaml
	$(YQ) -i '.image.tag = "$(TAG)"' $(HELM_DIR)/values.yaml
	$(YQ) -i '.image.os = "$(OS)"' $(HELM_DIR)/values.yaml
	$(YQ) -i '.image.arch = "$(ARCH)"' $(HELM_DIR)/values.yaml
	$(YQ) -i '.appVersion = "$(BASE_TAG)"' $(HELM_DIR)/Chart.yaml
	git add $(HELM_DIR)/values.yaml $(HELM_DIR)/Chart.yaml || true
	if ! git diff --cached --quiet; then \
	  git -c user.name="github-actions[bot]" -c user.email="github-actions[bot]@users.noreply.github.com" \
	    commit -m "ci(helm): bump image to $(FULL_TAG) (appVersion=$(BASE_TAG))" && \
	  git push origin HEAD:develop; \
	else \
	  echo "No chart changes to commit."; \
	fi

# -------------------------
# Cleanup
# -------------------------
clean:
	rm -rf bin/
	-docker rmi $(IMAGE):local 2>/dev/null || true
