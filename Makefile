BINARY_NAME ?= mdtoc
BUILD_DIR ?= build
GO ?= go
INSTALL_DIR ?= $(HOME)/.local/bin
RELEASE_IDENTITY ?=
SOURCE_REF ?= local
SOURCE_REVISION ?= unknown

VERSION_LDFLAGS := \
	-X 'mdtoc/internal/version.ReleaseIdentity=$(RELEASE_IDENTITY)' \
	-X 'mdtoc/internal/version.SourceRef=$(SOURCE_REF)' \
	-X 'mdtoc/internal/version.SourceRevision=$(SOURCE_REVISION)'

.PHONY: build check-testhelpers install release-build sync-embedded test validate-distribution

test:
	go test ./...

validate-distribution:
	@GO="$(GO)" sh ./scripts/validate-distribution.sh

build:
	@command -v "$(GO)" >/dev/null 2>&1 || { printf 'Go is required for source builds: %s\n' "$(GO)" >&2; exit 1; }
	@mkdir -p "$(BUILD_DIR)"
	@$(GO) build \
		-ldflags "$(VERSION_LDFLAGS)" \
		-o "$(BUILD_DIR)/$(BINARY_NAME)" \
		./cmd/mdtoc || { printf 'source build failed\n' >&2; exit 1; }

release-build:
	@[ -n "$(RELEASE_IDENTITY)" ] || { printf 'RELEASE_IDENTITY is required for release builds\n' >&2; exit 1; }
	@$(MAKE) build RELEASE_IDENTITY="$(RELEASE_IDENTITY)"

install: build
	@dest="$(INSTALL_DIR)/$(BINARY_NAME)"; \
	mkdir -p "$(INSTALL_DIR)" || { printf 'INSTALL_DIR is not usable: %s\n' "$(INSTALL_DIR)" >&2; exit 1; }; \
	cp "$(BUILD_DIR)/$(BINARY_NAME)" "$$dest" || { printf 'INSTALL_DIR is not usable: %s\n' "$(INSTALL_DIR)" >&2; exit 1; }; \
	chmod 0755 "$$dest" || { printf 'INSTALL_DIR is not usable: %s\n' "$(INSTALL_DIR)" >&2; exit 1; }; \
	if [ ! -x "$$dest" ]; then \
		printf 'installed mdtoc is not executable: %s\n' "$$dest" >&2; \
		exit 1; \
	fi; \
	if ! version_output=$$("$$dest" --version 2>&1); then \
		printf 'installed mdtoc at %s failed --version\n' "$$dest" >&2; \
		exit 1; \
	fi; \
	case "$$version_output" in \
	*"source"* ) ;; \
	*) printf 'installed mdtoc at %s did not report source provenance\n' "$$dest" >&2; exit 1 ;; \
	esac; \
	case "$$version_output" in \
	*"$(SOURCE_REF)"*"$(SOURCE_REVISION)"* ) ;; \
	*) printf 'installed mdtoc at %s did not report requested source provenance\n' "$$dest" >&2; exit 1 ;; \
	esac; \
	printf 'Installed mdtoc source ref=%s revision=%s to %s\n' "$(SOURCE_REF)" "$(SOURCE_REVISION)" "$$dest"

sync-embedded:

check-testhelpers:
