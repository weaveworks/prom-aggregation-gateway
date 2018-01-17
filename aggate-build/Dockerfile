FROM golang:1.9.1-stretch
RUN apt-get update && apt-get install -y python-requests python-yaml file jq && \
	rm -rf /var/lib/apt/lists/* /tmp/* /var/tmp/*
RUN curl -fsSLo shfmt https://github.com/mvdan/sh/releases/download/v1.3.0/shfmt_v1.3.0_linux_amd64 && \
	echo "b1925c2c405458811f0c227266402cf1868b4de529f114722c2e3a5af4ac7bb2  shfmt" | sha256sum -c && \
	chmod +x shfmt && \
	mv shfmt /usr/bin
RUN go clean -i net && \
	go install -tags netgo std && \
	go install -race -tags netgo std
RUN go get -tags netgo \
		github.com/fzipp/gocyclo \
		github.com/golang/lint/golint \
		github.com/kisielk/errcheck \
		github.com/mjibson/esc \
		github.com/client9/misspell/cmd/misspell && \
	rm -rf /go/pkg/ /go/src/
COPY build.sh /
ENTRYPOINT ["/build.sh"]
