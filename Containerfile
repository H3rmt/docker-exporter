FROM --platform=$BUILDPLATFORM golang:1.21-alpine@sha256:2414035b086e3c42b99654c8b26e6f5b1b1598080d65fd03c7f499552ff4dc94 AS builder
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

ENTRYPOINT ["/exporter"]
