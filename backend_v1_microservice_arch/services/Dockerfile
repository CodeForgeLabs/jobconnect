# FROM golang:1.26.1-alpine AS builder

# ARG SERVICE
# ARG CMD_PATH

# WORKDIR /src
# COPY . .
# WORKDIR /src/services/${SERVICE}

# RUN go mod download
# RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -ldflags="-s -w" -o /out/service ${CMD_PATH}

# FROM alpine:3.20
# RUN apk add --no-cache ca-certificates
# WORKDIR /app
# COPY --from=builder /out/service /app/service
# EXPOSE 50051 50052 50053 50054 50055 50056 50059 50064
# ENTRYPOINT ["/app/service"]
FROM golang:1.26.1-alpine AS builder

ARG SERVICE
ARG CMD_PATH

WORKDIR /src

# 1. Create the directory structure for your services
# This prevents the "no such file or directory" error during mod download
RUN mkdir -p gateway api services/${SERVICE}

# 2. Copy ALL go.mod and go.sum files from the project
# We use a wildcard to grab them all at once. This ensures that 'replace' 
# directives pointing to other local modules (like 'user' or 'api') work.
COPY go.mod* go.sum* ./
COPY gateway/go.mod* gateway/go.sum* ./gateway/
COPY api/go.mod* api/go.sum* ./api/
# This copies the mod files for every service so internal replacements are satisfied
COPY services/ ./services/
# Clean up: remove everything in services EXCEPT the go.mod/sum files 
# to keep the cache clean (optional but recommended)
RUN find services -type f ! -name 'go.mod' ! -name 'go.sum' -delete

# 3. Now download dependencies
WORKDIR /src/services/${SERVICE}
RUN go mod download

# 4. Copy the actual source code
WORKDIR /src
COPY . .

# 4. Build with Mount Caches
WORKDIR /src/services/${SERVICE}
RUN --mount=type=cache,target=/root/.cache/go-build \
    --mount=type=cache,target=/go/pkg/mod \
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -ldflags="-s -w" -o /out/service ${CMD_PATH}

# ...

FROM alpine:3.20
RUN apk add --no-cache ca-certificates
WORKDIR /app
COPY --from=builder /out/service /app/service
EXPOSE 50051 50052 50053 50054 50055 50056 50059 50064
ENTRYPOINT ["/app/service"]