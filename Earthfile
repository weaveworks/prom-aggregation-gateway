VERSION 0.6

ARG IMAGE_TAG="dev"

ARG GOLANG_VERSION="1.19.3"
ARG KUBECONFORM_VERSION="0.5.0"
ARG HELM_UNITTEST_VERSION="0.2.8"

go-deps:
    FROM golang:${GOLANG_VERSION}-alpine3.17

    WORKDIR /src
    COPY go.mod go.sum /src
    RUN go mod download

    SAVE ARTIFACT go.mod AS LOCAL go.mod
    SAVE ARTIFACT go.sum AS LOCAL go.sum

build-bin:
    FROM +go-deps

    WORKDIR /src
    COPY . /src
    RUN go build -o prom-aggregation-gateway .

    SAVE ARTIFACT ./prom-aggregation-gateway AS LOCAL ./dist/

test-bin:
    FROM +go-deps

    WORKDIR /src
    COPY . /src
    ENV CGO_ENABLED=0
    RUN go test .

    WORKDIR /src

build-docker:
    FROM alpine:3.17
    COPY +build-bin/prom-aggregation-gateway .
    ENTRYPOINT ["/prom-aggregation-gateway"]
    SAVE IMAGE ghcr.io/zapier/prom-aggregation-gateway:${IMAGE_TAG}

golang-test:
    FROM golang:${GOLANG_VERSION}-alpine3.17


helm-test:
    ARG ct_args=''
    FROM quay.io/helmpack/chart-testing:v3.7.1

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

chart-releaser:
    ARG CHART_RELEASER_VERSION="1.4.1"
    FROM quay.io/helmpack/chart-releaser:v${CHART_RELEASER_VERSION}

    WORKDIR /src
    COPY . /src
    RUN ls -al

    RUN cr --config .github/cr.yaml package charts/*
    SAVE ARTIFACT .cr-release-packages/ AS LOCAL ./dist
    RUN --push cr --config .github/cr.yaml upload --skip-existing --push
    RUN --push cr --config .github/cr.yaml index


