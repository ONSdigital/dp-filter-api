SHELL=bash
MAIN=dp-filter-api

BUILD=build
BIN_DIR?=.

BUILD_TIME=$(shell date +%s)
GIT_COMMIT=$(shell git rev-parse HEAD)
VERSION ?= $(shell git tag --points-at HEAD | grep ^v | head -n 1)
LDFLAGS=-ldflags "-w -s -X 'main.Version=${VERSION}' -X 'main.BuildTime=$(BUILD_TIME)' -X 'main.GitCommit=$(GIT_COMMIT)'"

export GRAPH_DRIVER_TYPE?=neptune
export GRAPH_ADDR?=ws://localhost:8182/gremlin

.PHONY: all
all: audit test build

.PHONY: audit
audit:
	nancy go.sum

.PHONY: build
build:
	@mkdir -p $(BUILD)/$(BIN_DIR)
	go build $(LDFLAGS) -o $(BUILD)/$(BIN_DIR)/dp-filter-api cmd/$(MAIN)/main.go

.PHONY: debug
debug:
	HUMAN_LOG=1 go run $(LDFLAGS) -race cmd/$(MAIN)/main.go

.PHONY: acceptance-publishing
acceptance-publishing:
	MONGODB_FILTERS_DATABASE=test HUMAN_LOG=1 go run $(LDFLAGS) -race cmd/$(MAIN)/main.go

.PHONY: acceptance-web
acceptance-web:
	ENABLE_PRIVATE_ENDPOINTS=false MONGODB_FILTERS_DATABASE=test HUMAN_LOG=1 go run $(LDFLAGS) -race cmd/$(MAIN)/main.go

.PHONY: test
test:
	go test -cover -race ./...
.PHONY: build debug acceptance test
