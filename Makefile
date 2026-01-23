.PHONY: build build-version major minor feature

VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_DATE ?= $(shell date -u +%Y-%m-%dT%H:%M:%SZ)

LDFLAGS = -s -w \
	-X github.com/aottr/compose-sync/internal/version.Version=$(VERSION) \
	-X github.com/aottr/compose-sync/internal/version.Commit=$(COMMIT) \
	-X github.com/aottr/compose-sync/internal/version.BuildDate=$(BUILD_DATE)

# Get current version tag (removes 'v' prefix if present, defaults to 0.0.0)
CURRENT_TAG := $(shell git describe --tags --abbrev=0 2>/dev/null | sed 's/^v//' || echo "0.0.0")
CURRENT_MAJOR := $(shell echo "$(CURRENT_TAG)" | awk -F. '{print $$1}' | grep -E '^[0-9]+$$' || echo "0")
CURRENT_MINOR := $(shell echo "$(CURRENT_TAG)" | awk -F. '{print $$2}' | grep -E '^[0-9]+$$' || echo "0")
CURRENT_PATCH := $(shell echo "$(CURRENT_TAG)" | awk -F. '{print $$3}' | grep -E '^[0-9]+$$' || echo "0")

build:
	go build -ldflags "$(LDFLAGS)" -o compose-sync ./cmd/compose-sync

build-version: build
	@./compose-sync -version

major:
	@CURRENT="$(CURRENT_TAG)"; \
	if [ -z "$$CURRENT" ] || [ "$$CURRENT" = "0.0.0" ]; then \
		NEW_VERSION="1.0.0"; \
	else \
		MAJOR=$$(echo "$$CURRENT" | awk -F. '{print $$1}'); \
		NEW_VERSION="$$(($$MAJOR + 1)).0.0"; \
	fi; \
	echo "Bumping major version: $$CURRENT -> $$NEW_VERSION"; \
	git tag -s "v$$NEW_VERSION" -m "Release v$$NEW_VERSION"; \
	echo "Created signed tag v$$NEW_VERSION. Push with: git push origin v$$NEW_VERSION"

minor:
	@CURRENT="$(CURRENT_TAG)"; \
	if [ -z "$$CURRENT" ] || [ "$$CURRENT" = "0.0.0" ]; then \
		NEW_VERSION="0.1.0"; \
	else \
		MAJOR=$$(echo "$$CURRENT" | awk -F. '{print $$1}'); \
		MINOR=$$(echo "$$CURRENT" | awk -F. '{print $$2}'); \
		NEW_VERSION="$$MAJOR.$$(($$MINOR + 1)).0"; \
	fi; \
	echo "Bumping minor version: $$CURRENT -> $$NEW_VERSION"; \
	git tag -s "v$$NEW_VERSION" -m "Release v$$NEW_VERSION"; \
	echo "Created signed tag v$$NEW_VERSION. Push with: git push origin v$$NEW_VERSION"

patch:
	@CURRENT="$(CURRENT_TAG)"; \
	if [ -z "$$CURRENT" ] || [ "$$CURRENT" = "0.0.0" ]; then \
		NEW_VERSION="0.0.1"; \
	else \
		MAJOR=$$(echo "$$CURRENT" | awk -F. '{print $$1}'); \
		MINOR=$$(echo "$$CURRENT" | awk -F. '{print $$2}'); \
		PATCH=$$(echo "$$CURRENT" | awk -F. '{print $$3}'); \
		NEW_VERSION="$$MAJOR.$$MINOR.$$(($$PATCH + 1))"; \
	fi; \
	echo "Bumping patch version: $$CURRENT -> $$NEW_VERSION"; \
	git tag -s "v$$NEW_VERSION" -m "Release v$$NEW_VERSION"; \
	echo "Created signed tag v$$NEW_VERSION. Push with: git push origin v$$NEW_VERSION"
