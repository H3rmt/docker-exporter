FROM --platform=$BUILDPLATFORM golang:1.26-alpine@sha256:d4c4845f5d60c6a974c6000ce58ae079328d03ab7f721a0734277e69905473e5 AS builder
ARG TARGETOS
ARG TARGETARCH
ARG VERSION=main

RUN apk --no-cache add ca-certificates tzdata && update-ca-certificates
RUN echo "nobody:x:3000:3000:Nobody:/:" > /tmp/passwd
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

ENV CGO_ENABLED=0
RUN GOOS=$TARGETOS GOARCH=$TARGETARCH go build -ldflags "-X main.Version=${VERSION}" -o exporter ./cmd/main.go

FROM --platform=$TARGETOS/$TARGETARCH scratch
WORKDIR /tmp
COPY --from=builder /tmp/passwd /etc/passwd
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo

COPY --from=builder /app/exporter /exporter

# needed as only root can access docker socket
USER 0:0

ENTRYPOINT ["/exporter"]
