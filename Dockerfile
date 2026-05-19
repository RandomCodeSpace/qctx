# syntax=docker/dockerfile:1.7
FROM golang:1.23-alpine AS builder
WORKDIR /src
RUN apk add --no-cache git make
COPY go.mod go.sum ./
RUN go mod download
COPY . .
ARG VERSION=dev
ARG COMMIT=none
ARG DATE=unknown
RUN CGO_ENABLED=0 go build -trimpath \
      -ldflags="-s -w \
        -X github.com/RandomCodeSpace/qctx/internal/version.Version=${VERSION} \
        -X github.com/RandomCodeSpace/qctx/internal/version.Commit=${COMMIT} \
        -X github.com/RandomCodeSpace/qctx/internal/version.Date=${DATE}" \
      -o /out/qctx ./cmd/qctx

FROM gcr.io/distroless/static:nonroot
COPY --from=builder /out/qctx /usr/local/bin/qctx
USER nonroot:nonroot
ENTRYPOINT ["/usr/local/bin/qctx"]
