SHELL=bash
MAIN=dp-filter-api

BUILD=build
BIN_DIR?=.

BUILD_TIME=$(shell date +%s)
GIT_COMMIT=$(shell git rev-parse HEAD)
VERSION ?= $(shell git tag --points-at HEAD | grep ^v | head -n 1)
LDFLAGS=-ldflags "-w -s -X 'main.Version=${VERSION}' -X 'main.BuildTime=$(BUILD_TIME)' -X 'main.GitCommit=$(GIT_COMMIT)'"

export GRAPH_DRIVER_TYPE?=neo4j
export GRAPH_ADDR?=bolt://localhost:7687

.PHONY: all
all: audit test build

.PHONY: audit-go
audit-go:
	dis-vulncheck

.PHONY: audit-python
audit-python:
	$(MAKE) -C sdk/python audit

.PHONY: audit
audit: audit-go audit-python

.PHONY: build
build:
	@mkdir -p $(BUILD)/$(BIN_DIR)
	go build $(LDFLAGS) -o $(BUILD)/$(BIN_DIR)/dp-filter-api cmd/$(MAIN)/main.go

.PHONY: debug
debug:
	HUMAN_LOG=1 go run $(LDFLAGS) -race cmd/$(MAIN)/main.go

.PHONY: lint-go
lint-go:
	golangci-lint run ./...

.PHONY: lint-python
lint-python:
	$(MAKE) -C sdk/python lint

.PHONY: lint-python-types
lint-python-types:
	$(MAKE) -C sdk/python typecheck

.PHONY: lint
lint: lint-go lint-python lint-python-types

.PHONY: format-python
format-python:
	$(MAKE) -C sdk/python format

.PHONY: test-component
test-component:
	exit

.PHONY: acceptance-publishing
acceptance-publishing:
	MONGODB_FILTERS_DATABASE=test HUMAN_LOG=1 go run $(LDFLAGS) -race cmd/$(MAIN)/main.go

.PHONY: acceptance-web
acceptance-web:
	ENABLE_PRIVATE_ENDPOINTS=false MONGODB_FILTERS_DATABASE=test HUMAN_LOG=1 go run $(LDFLAGS) -race cmd/$(MAIN)/main.go

.PHONY: test-go
test-go:
	go test -cover -race ./...

.PHONY: test-python
test-python:
	$(MAKE) -C sdk/python test

.PHONY: test
test: test-go test-python
