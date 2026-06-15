BINARY  := imgup
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
LDFLAGS := -s -w -X github.com/dengshu2/imgup/cmd.version=$(VERSION)

.PHONY: build install test vet fmt clean help

build: ## Build the binary into ./bin
	go build -ldflags "$(LDFLAGS)" -o bin/$(BINARY) .

install: ## Install into GOBIN with version stamped in
	go install -ldflags "$(LDFLAGS)" .

test: ## Run tests with the race detector
	go test -race ./...

vet: ## Run go vet
	go vet ./...

fmt: ## Report files that are not gofmt'd
	gofmt -l .

clean: ## Remove build artifacts
	rm -rf bin dist

help: ## List targets
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "  %-10s %s\n", $$1, $$2}'
