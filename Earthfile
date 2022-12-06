VERSION 0.6

ARG version="dev"
ARG image_name="prom-aggregation-gateway"

ARG ALPINE_VERSION="3.17"
ARG CHART_RELEASER_VERSION="1.4.1"
ARG CHART_TESTING_VERSION="3.7.1"
ARG GITHUB_CLI_VERSION="2.20.2"
ARG GOLANG_VERSION="1.19.3"
ARG HELM_UNITTEST_VERSION="0.2.8"
ARG KUBECONFORM_VERSION="0.5.0"
ARG STATICCHECK_VERSION="0.3.3"

test:
    BUILD +lint-golang
    BUILD +test-golang
    BUILD +test-helm

build:
    BUILD +build-binary
    BUILD +build-docker
    BUILD +build-helm

release:
    BUILD +release-binary
    BUILD +build-docker

build-docker:
    FROM alpine:${ALPINE_VERSION}
    COPY +build-binary/prom-aggregation-gateway .
    ENTRYPOINT ["/prom-aggregation-gateway"]
    SAVE IMAGE --push ${image_name}:${version}

continuous-deploy:
    BUILD +release-helm

go-deps:
    FROM golang:${GOLANG_VERSION}-alpine${ALPINE_VERSION}

    WORKDIR /src
    COPY go.mod go.sum /src
    RUN go mod download

    SAVE ARTIFACT go.mod AS LOCAL go.mod
    SAVE ARTIFACT go.sum AS LOCAL go.sum

build-binary:
    FROM +go-deps

    WORKDIR /src
    COPY . /src
    RUN go build -o prom-aggregation-gateway .

    SAVE ARTIFACT ./prom-aggregation-gateway AS LOCAL ./dist/

release-binary:
    FROM +build-binary

    COPY +build-binary/prom-aggregation-gateway .

    # install github cli
    RUN FILE=ghcli.tgz \
        && URL=https://github.com/cli/cli/releases/download/v${GITHUB_CLI_VERSION}/gh_${GITHUB_CLI_VERSION}_linux_amd64.tar.gz \
        && wget ${URL} \
            --output-document ${FILE} \
        && tar \
            --extract \
            --verbose \
            --directory /usr \
            --strip-components=1 \
            --file ${FILE} \
        && gh version

    RUN --push gh release create ${version} ./prom-aggregation-gateway

lint-golang:
    FROM +go-deps

    # install staticcheck
    RUN FILE=staticcheck.tgz \
        && URL=https://github.com/dominikh/go-tools/releases/download/v${STATICCHECK_VERSION}/staticcheck_linux_amd64.tar.gz \
        && wget ${URL} \
            --output-document ${FILE} \
        && tar \
            --extract \
            --verbose \
            --directory /bin \
            --strip-components=1 \
            --file ${FILE} \
        && staticcheck -version

    ENV CGO_ENABLED=0
    COPY . /src
    RUN staticcheck ./...

test-golang:
    FROM +go-deps

    COPY . /src

    ENV CGO_ENABLED=0
    RUN go test .

test-helm:
    ARG ct_args=''
    FROM quay.io/helmpack/chart-testing:v${CHART_TESTING_VERSION}

    # install kubeconform
    RUN FILE=kubeconform.tgz \
        && URL=https://github.com/yannh/kubeconform/releases/download/v${KUBECONFORM_VERSION}/kubeconform-linux-amd64.tar.gz \
        && wget ${URL} \
            --output-document ${FILE} \
        && tar \
            --extract \
            --verbose \
            --directory /bin \
            --file ${FILE} \
        && kubeconform -v

    RUN apk add --no-cache bash git \
        && helm plugin install --version "${HELM_UNITTEST_VERSION}" https://github.com/quintush/helm-unittest \
        && helm unittest --help

    # actually lint the chart
    WORKDIR /src
    COPY . /src
    RUN ct --config ./.github/ct.yaml lint ./charts --all

build-helm:
    FROM quay.io/helmpack/chart-releaser:v${CHART_RELEASER_VERSION}

    WORKDIR /src
    COPY . /src

    RUN cr --config .github/cr.yaml package charts/*
    SAVE ARTIFACT .cr-release-packages/ AS LOCAL ./dist

release-helm:
    FROM quay.io/helmpack/chart-releaser:v${CHART_RELEASER_VERSION}

    ARG token

    WORKDIR /src
    COPY . /src

    RUN cr --config .github/cr.yaml package charts/*

    RUN mkdir -p .cr-index
    RUN git config --global user.email "opensource@zapier.com"
    RUN git config --global user.name "Open Source at Zapier"

    RUN --push cr --config .github/cr.yaml upload --token $token --skip-existing
    RUN --push cr --config .github/cr.yaml index --token $token --push
