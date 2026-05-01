FROM golang:1.26.1-alpine AS builder

ARG SERVICE
ARG CMD_PATH

WORKDIR /src

RUN mkdir -p gateway api services

# Copy only module manifests first to maximize cache hits.
COPY gateway/go.mod gateway/go.sum ./gateway/
COPY api/go.mod api/go.sum ./api/
COPY services/*/go.mod services/*/go.sum ./services/

WORKDIR /src/services/${SERVICE}
RUN --mount=type=cache,target=/go/pkg/mod go mod download

# Copy source only after dependency layer.
WORKDIR /src
COPY . .

WORKDIR /src/services/${SERVICE}
RUN --mount=type=cache,target=/root/.cache/go-build \
    --mount=type=cache,target=/go/pkg/mod \
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -ldflags="-s -w" -o /out/service ${CMD_PATH}

FROM alpine:3.20
RUN apk add --no-cache ca-certificates
WORKDIR /app
COPY --from=builder /out/service /app/service
EXPOSE 50051 50052 50053 50054 50055 50056 50059 50064
ENTRYPOINT ["/app/service"]
