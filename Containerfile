FROM --platform=$BUILDPLATFORM golang:1.27-alpine@sha256:70bd7ddea8b6cbe5d6943cde279d6b67feb10cc924fb9100d1d8757c1575b280 AS builder
ARG TARGETOS
ARG TARGETARCH
ARG VERSION=main

RUN apk --no-cache add ca-certificates tzdata && update-ca-certificates
RUN echo "root:x:0:0::/:/bin/false" > /tmp/passwd
RUN echo "NAME=\"scratch\"" >> /tmp/os-release
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

ENV CGO_ENABLED=0
RUN GOOS=$TARGETOS GOARCH=$TARGETARCH go build -ldflags "-X main.Version=${VERSION}" -o exporter ./cmd/main.go

FROM --platform=$TARGETOS/$TARGETARCH scratch
WORKDIR /tmp
COPY --from=builder /tmp/passwd /etc/passwd
COPY --from=builder /tmp/os-release /etc/os-release
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo

COPY --from=builder /app/exporter /exporter

# needed as only root can access docker socket
USER 0:0

ENTRYPOINT ["/exporter"]
