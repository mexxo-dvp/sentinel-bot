ROOT := $(shell git rev-parse --show-toplevel 2>/dev/null || pwd)
HOOK_DIR := $(ROOT)/.git/hooks
APP := gitleaks-precommit
BIN := $(HOOK_DIR)/$(APP)
CONFIG := $(ROOT)/tools/gitleaks/config.toml

.PHONY: build install enable disable uninstall scan

build:
	@echo ">> build $(APP)"
	cd "$(ROOT)" && go build -o "$(BIN)" ./tools/gitleaks/cmd/$(APP)
	@chmod +x "$(BIN)"

install: build
	@echo ">> write pre-commit hook"
	@mkdir -p "$(HOOK_DIR)"
	@{ \
	  echo '#!/usr/bin/env bash'; \
	  echo 'set -euo pipefail'; \
	  echo '"$$(git rev-parse --git-path hooks)/$(APP)"'; \
	} > "$(HOOK_DIR)/pre-commit"
	@chmod +x "$(HOOK_DIR)/pre-commit"
	@echo "[OK] pre-commit installed"

enable:
	@git -C "$(ROOT)" config --local gitleaks.precommit.enable true
	@echo "[OK] enable"

disable:
	@git -C "$(ROOT)" config --local gitleaks.precommit.enable false
	@echo "[OK] disable"

uninstall:
	@rm -f "$(BIN)" "$(HOOK_DIR)/pre-commit"
	@echo "[OK] uninstalled"

scan:
	@command -v gitleaks >/dev/null 2>&1 || { echo "gitleaks not in PATH"; exit 1; }
	@cd "$(ROOT)" && gitleaks detect --redact --exit-code 1 --config "$(CONFIG)"
