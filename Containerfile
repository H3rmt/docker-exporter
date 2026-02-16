FROM --platform=$BUILDPLATFORM golang:1.26-alpine@sha256:d4c4845f5d60c6a974c6000ce58ae079328d03ab7f721a0734277e69905473e5 AS builder
ARG TARGETOS
ARG TARGETARCH
ARG VERSION=main

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

ENV CGO_ENABLED=0
RUN GOOS=$TARGETOS GOARCH=$TARGETARCH go build -ldflags "-X main.Version=${VERSION}" -o exporter ./cmd/main.go

FROM --platform=$TARGETOS/$TARGETARCH ghcr.io/arca-consult/scratch:0.0.2@sha256:ea3be3c3643833df48d7883c3f0caa9b891087d3b88ff553e2f3a928d7c267bd
WORKDIR /
COPY --from=builder /app/exporter /exporter

# needed as only root can access docker socket
USER 0:0

ENTRYPOINT ["/exporter"]
