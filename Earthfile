VERSION 0.6

helm-test:
    ARG ct_args=''
    FROM quay.io/helmpack/chart-testing:v3.7.1

    ARG KUBECONFORM_VERSION="0.5.0"
    ARG HELM_UNITTEST_VERSION="0.2.8"

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
