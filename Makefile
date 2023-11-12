## template Makefile:
## service example
#:

SHELL          = /bin/sh
PRG           ?= $(shell basename $$PWD)
GO            ?= go
SOURCES        = $(shell find . -maxdepth 3 -mindepth 1 -path ./var -prune -o -name '*.go')
TARGETOS      ?= linux
TARGETARCH    ?= amd64
LDFLAGS       := -s -w -extldflags '-static'

all: help

# ------------------------------------------------------------------------------
## Compile operations
#:

.PHONY: build

## Build app
build: $(PRG)

$(PRG): $(SOURCES) go.*
	@GOOS=${TARGETOS} GOARCH=${TARGETARCH} \
	  $(GO) build -v -o $@ .

## Build & run app
run: $(PRG)
	@./$(PRG)

.PHONY: fmt

## Format go sources
fmt:
	$(GO) fmt ./...

.PHONY: lint

## Run lint
lint:
	@which golint > /dev/null || $(GO) install golang.org/x/lint/golint@latest
	@golint ./...

.PHONY: ci-lint

## Run golangci-lint
ci-lint:
	@golangci-lint run ./...

.PHONY: vet

## Run vet
vet:
	@$(GO) vet ./...

## Run tests
test: lint vet coverage.out

.PHONY: test-race

test-race:
	$(GO) test -tags test -race -covermode=atomic ./...

# internal target
coverage.out: $(SOURCES)
	@#GIN_MODE=release $(GO) test -test.v -test.race -coverprofile=$@ -covermode=atomic ./...
	$(GO) test -tags test -covermode=atomic -coverprofile=$@ ./...

## Open coverage report in browser
cov-html: coverage.out
	$(GO) tool cover -html=$<

## Show code coverage per func
cov-func: coverage.out
	$(GO) tool cover -func coverage.out

## Show total code coverage
cov-total: coverage.out
	@$(GO) tool cover -func coverage.out | grep total: | awk '{print $$3}'

## Clean coverage report
cov-clean:
	rm -f coverage.*

## count LoC without generated code
cloc:
	@cloc --md --fullpath .

## Changes from last tag
changelog:
	@echo Changes since $(RELEASE)
	@echo
	@git log $(RELEASE)..@ --pretty=format:"* %s"

# ------------------------------------------------------------------------------
## Other
#:

# This code handles group header and target comment with one or two lines only
## list Makefile targets
## (this is default target)
help:
	@grep -A 1 -h "^## " $(MAKEFILE_LIST) \
  | sed -E 's/^--$$// ; /./{H;$$!d} ; x ; s/^\n## ([^\n]+)\n(## (.+)\n)*(.+):(.*)$$/"    " "\4" "\1" "\3"/' \
  | sed -E 's/^"    " "#" "(.+)" "(.*)"$$/"" "" "" ""\n"\1 \2" "" "" ""/' \
  | xargs printf "%s\033[36m%-15s\033[0m %s %s\n"
