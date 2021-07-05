FROM golang:1.16.5-alpine3.12 as builder
WORKDIR /go/src/github.com/mendersoftware/mender-reporting
COPY ./ .
RUN CGO_ENABLED=1 GOARCH=amd64 go build

FROM alpine:3.13.0
RUN apk add --no-cache ca-certificates xz
RUN mkdir -p /etc/mender-reporting
COPY ./config.yaml /etc/mender-reporting/
COPY --from=builder /go/src/github.com/mendersoftware/mender-reporting/reporting /usr/bin/
ENTRYPOINT ["/usr/bin/reporting", "--config", "/etc/mender-reporting/config.yaml"]

EXPOSE 8080
