.PHONY: help

IMAGE_NAME ?= "larktech/prom-aggregation-gateway"
IMAGE_TAG ?= "latest"

help:
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

test: ## Run Go tests
	go -v test ./...

build: ## Build docker image
	docker build . --tag=${IMAGE_NAME}:${IMAGE_TAG}

push: ## Push the image to docker
	docker push ${IMAGE_NAME}:${IMAGE_TAG}