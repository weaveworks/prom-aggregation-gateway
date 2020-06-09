FROM golang:alpine as builder
ADD . /src/github.com/LarkTechnologies/prom-aggregation-gateway
WORKDIR /src/github.com/LarkTechnologies/prom-aggregation-gateway
RUN \
  CGO_ENABLED=0 GOOS=linux \
  go build -o /go/bin/ -a -installsuffix cgo -ldflags '-extldflags "-static"' \
  ./cmd/...

FROM scratch
COPY --from=builder /go/bin /app/
EXPOSE 80
WORKDIR /app
CMD ["./main"]
