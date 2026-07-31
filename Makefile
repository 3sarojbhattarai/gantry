# Gantry — developer tasks.
# Run `make` or `make help` for the list.

BINARY      := gantry
CMD         := ./cmd/gantry
BIN_DIR     := bin
VERSION     ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
LDFLAGS     := -s -w -X github.com/3sarojbhattarai/gantry/internal/cli.version=$(VERSION)
WEB_DIR     := web

GO          ?= go

.DEFAULT_GOAL := help

## ---- Go ---------------------------------------------------------------------

.PHONY: build
build: ## Build the gantry binary into bin/
	$(GO) build -ldflags "$(LDFLAGS)" -o $(BIN_DIR)/$(BINARY) $(CMD)

.PHONY: install
install: ## Install gantry into $GOBIN
	$(GO) install -ldflags "$(LDFLAGS)" $(CMD)

.PHONY: run
run: ## Run gantry (pass args with ARGS="version")
	$(GO) run -ldflags "$(LDFLAGS)" $(CMD) $(ARGS)

.PHONY: test
test: ## Run unit tests (fake-backed, no daemon)
	$(GO) test ./...

.PHONY: test-integration
test-integration: ## Run integration tests (needs a real Docker daemon)
	$(GO) test -tags integration ./...

.PHONY: cover
cover: ## Run unit tests with a coverage profile
	$(GO) test -coverprofile=coverage.out ./...
	$(GO) tool cover -func=coverage.out | tail -1

.PHONY: vet
vet: ## Run go vet
	$(GO) vet ./...

.PHONY: lint
lint: ## Run golangci-lint
	golangci-lint run

.PHONY: fmt
fmt: ## Format Go sources
	gofmt -w .

.PHONY: fmt-check
fmt-check: ## Fail if any Go source is unformatted
	@out=$$(gofmt -l .); if [ -n "$$out" ]; then echo "unformatted:"; echo "$$out"; exit 1; fi

.PHONY: tidy
tidy: ## Tidy go.mod/go.sum
	$(GO) mod tidy

.PHONY: check
check: fmt-check vet lint test ## Full pre-push gate (matches CI)

.PHONY: clean
clean: ## Remove build and coverage artifacts
	rm -rf $(BIN_DIR) coverage.out

## ---- Frontend (Phase 4+; React + Vite + Tailwind) ---------------------------

.PHONY: web-install
web-install: ## Install frontend dependencies
	cd $(WEB_DIR) && npm install

.PHONY: web-dev
web-dev: ## Run the Vite dev server
	cd $(WEB_DIR) && npm run dev

.PHONY: web-build
web-build: ## Build the frontend into web/dist (embedded by the binary)
	cd $(WEB_DIR) && npm run build

.PHONY: web-test
web-test: ## Run frontend unit tests (Vitest)
	cd $(WEB_DIR) && npm run test

## ---- Meta -------------------------------------------------------------------

.PHONY: help
help: ## Show this help
	@grep -hE '^[a-zA-Z0-9_-]+:.*?## ' $(MAKEFILE_LIST) \
		| awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-18s\033[0m %s\n", $$1, $$2}'
