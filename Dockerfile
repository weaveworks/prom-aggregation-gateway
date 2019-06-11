FROM golang:1.12 AS builder
WORKDIR /go/src/github.com/moia-dev/prom-aggregation-gateway/
COPY . .
RUN make build

FROM alpine:3.9
COPY --from=builder /go/src/github.com/moia-dev/prom-aggregation-gateway/bin/server /server
ENTRYPOINT ./server