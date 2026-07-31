BINARY  := gantry
CMD     := ./cmd/gantry
BIN     := bin/$(BINARY)
WEB     := web
MODULE  := github.com/3sarojbhattarai/gantry
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
LDFLAGS := -s -w -X $(MODULE)/internal/cli.version=$(VERSION)
GO      ?= go

.DEFAULT_GOAL := help

.PHONY: help build build-embed install run \
        test test-integration cover vet lint fmt fmt-check tidy check clean \
        web-install web-dev web-build web-test \
        docker snapshot release-check

help:
	@echo 'Targets:'
	@grep -oE '^[a-zA-Z0-9_-]+:' $(MAKEFILE_LIST) | sed 's/://' | sort -u | sed 's/^/  /'

build:
	$(GO) build -ldflags '$(LDFLAGS)' -o $(BIN) $(CMD)

build-embed: web-build
	$(GO) build -tags embed -ldflags '$(LDFLAGS)' -o $(BIN) $(CMD)

install:
	$(GO) install -ldflags '$(LDFLAGS)' $(CMD)

run:
	$(GO) run -ldflags '$(LDFLAGS)' $(CMD) $(ARGS)

test:
	$(GO) test ./...

test-integration:
	$(GO) test -tags integration ./...

cover:
	$(GO) test -coverprofile=coverage.out ./...
	$(GO) tool cover -func=coverage.out | tail -1

vet:
	$(GO) vet ./...

lint:
	golangci-lint run

fmt:
	gofmt -w .

fmt-check:
	@unformatted=$$(gofmt -l .); [ -z "$$unformatted" ] || { echo "$$unformatted"; exit 1; }

tidy:
	$(GO) mod tidy

check: fmt-check vet lint test

clean:
	rm -rf bin coverage.out

web-install:
	cd $(WEB) && npm install

web-dev:
	cd $(WEB) && npm run dev

web-build:
	cd $(WEB) && npm run build

web-test:
	cd $(WEB) && npm run test

docker:
	docker build --build-arg VERSION=$(VERSION) -t $(BINARY):$(VERSION) -t $(BINARY):latest .

snapshot:
	goreleaser release --snapshot --clean

release-check:
	goreleaser check
