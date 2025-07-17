FROM --platform=$BUILDPLATFORM golang:1.24-alpine@sha256:48ee313931980110b5a91bbe04abdf640b9a67ca5dea3a620f01bacf50593396 AS builder
ARG TARGETOS
ARG TARGETARCH

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

ENV CGO_ENABLED=0
RUN GOOS=$TARGETOS GOARCH=$TARGETARCH go build -o exporter ./cmd/main.go

FROM --platform=$TARGETOS/$TARGETARCH ghcr.io/arca-consult/scratch:0.0.2@sha256:ea3be3c3643833df48d7883c3f0caa9b891087d3b88ff553e2f3a928d7c267bd
WORKDIR /
COPY --from=builder /app/exporter /exporter

# needed as only root can access docker socket
USER 0:0

ENTRYPOINT ["/exporter"]
