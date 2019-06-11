SHELL := /bin/bash -eu

AWS_REGION            := eu-central-1
BUILD_DIR             := bin
PACKAGES              := $(shell go list ./...)
LINT_TARGETS          := $(shell go list -f '{{.Dir}}' ./... | sed -e"s|${CURDIR}/\(.*\)\$$|\1/...|g" | grep -v ^node_modules/ )
DEPENDENCIES          := $(shell find . -type f -name '*.go')
BINARIES              := $(shell find ./cmd -name 'main.go' | grep -v -e node_modules -e vendor |awk -F/ '{print "bin/" $$3}')
SERVICE               := prometheus-aggregation-pushgateway
GITHUB_OWNER          := moia-dev
GITHUB_REPOSITORY     := $(SERVICE)
REPOSITORY            := $(GITHUB_OWNER)/$(GITHUB_REPOSITORY)
DOCKER_REGISTRY       := 614608043005.dkr.ecr.eu-central-1.amazonaws.com
SYSTEM                := $(shell uname -s | tr A-Z a-z)_$(shell uname -m | sed "s/x86_64/amd64/")
BUILD_TIME            := $(shell date +%FT%T%z)
GOLANGCI_LINT_VERSION := 1.15.0
GIT_HASH              := $(shell git describe --always --tags --dirty)
VERSION               := $(shell echo $$(date "+%Y%m%d")-$$(git rev-parse --short HEAD) )
SLACK_ICON            := :telescope:

all: lint build test

$(BUILD_DIR)/%: cmd/%/*.go $(DEPENDENCIES)
	CGO_ENABLED=0 go build -ldflags="-s -w -X main.BuildTime=$(BUILD_TIME) -X main.Version=$(VERSION)" -o $@ ./cmd/$(notdir $@)

.PHONY: build
build: $(BINARIES)

tools/golangci-lint:
	mkdir -p tools/
	curl -sSLf \
		https://github.com/golangci/golangci-lint/releases/download/v$(GOLANGCI_LINT_VERSION)/golangci-lint-$(GOLANGCI_LINT_VERSION)-$(shell echo $(SYSTEM) | tr '_' '-').tar.gz \
		| tar xzOf - golangci-lint-$(GOLANGCI_LINT_VERSION)-$(shell echo $(SYSTEM) | tr '_' '-')/golangci-lint > tools/golangci-lint && chmod +x tools/golangci-lint

.PHONY: codecov
codecov: codecov-report codecov-publish

.PHONY: codecov-report
codecov-report:
	curl --data-binary @codecov.yml https://codecov.io/validate
	go test -race -coverprofile=coverage.txt -covermode=atomic ./...

.PHONY: codecov-publish
codecov-publish: guard-CODECOV_TOKEN
	bash <(curl -s https://codecov.io/bash) -t $(CODECOV_TOKEN)

.PHONY: docker-build
docker-build:
	docker build -t $(DOCKER_REGISTRY)/$(SERVICE):$(VERSION) .

.PHONY: push-image
push-image: docker-build
	$$(aws ecr get-login --no-include-email --region $(AWS_REGION))
	docker push $(DOCKER_REGISTRY)/$(SERVICE):$(VERSION)

.PHONY: ecr
ecr: deployment/cloudformation/ecr.yml
	aws cloudformation deploy \
		--no-fail-on-empty-changeset \
		--template-file $< \
		--stack-name $(SERVICE)-ecr \
		--parameter-overrides RepositoryName=$(SERVICE) \
		--region $(AWS_REGION)

.PHONY: guard-%
guard-%:
	$(if $(value ${*}),,$(error "Variable ${*} not set!"))

.PHONY: print-%
print-%  : ; @echo $* = $($*)