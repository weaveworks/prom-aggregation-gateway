# builder image
FROM golang:1.19-alpine3.16 as builder
RUN mkdir /build
ADD . /build/
WORKDIR /build
RUN CGO_ENABLED=0 GOOS=linux go build -a -o prom-aggregation-gateway .

# generate clean, final image for end users
FROM alpine:3.16
COPY --from=builder /build/prom-aggregation-gateway .

# executable
ENTRYPOINT [ "./prom-aggregation-gateway" ]
