.PHONY: install-tools setup-tools install-go-tools format format-check test build static-analysis ci

export PATH := $(shell go env GOPATH)/bin:$(PATH)
export GOCACHE := $(CURDIR)/.cache/go-build
export GOLANGCI_LINT_CACHE := $(CURDIR)/.cache/golangci-lint

install-tools: setup-tools

setup-tools:
	go run ./cmd/kumite setup

install-go-tools:
	go install golang.org/x/tools/cmd/deadcode@latest
	go install github.com/roblaszczak/go-cleanarch@latest
	go install github.com/loov/goda@latest
	go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@latest

format:
	gofmt -w ./cmd ./internal

format-check:
	test -z "$$(gofmt -l ./cmd ./internal)"

test:
	go test ./...

build:
	mkdir -p .cache/bin
	go build -o .cache/bin/kumite ./cmd/kumite
	go build -o .cache/bin/kumite-installer ./cmd/kumite-installer

static-analysis:
	deadcode -test ./...
	golangci-lint run ./...
	go mod tidy
	git diff --exit-code -- go.mod go.sum
	go-cleanarch
	mkdir -p .cache
	goda graph "./..." | dot -Tsvg -o .cache/graph.svg

ci: format-check test static-analysis build
