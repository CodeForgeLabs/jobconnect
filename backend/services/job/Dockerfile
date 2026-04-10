FROM golang:1.25-alpine AS builder

ARG SERVICE
ARG CMD_PATH

WORKDIR /src
# We are copying from the backend root (context: .)
COPY . .
WORKDIR /src/services/job

RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -ldflags="-s -w" -o /out/service ${CMD_PATH}

FROM alpine:3.20
RUN apk add --no-cache ca-certificates
WORKDIR /app
COPY --from=builder /out/service /app/service
EXPOSE 50053
ENTRYPOINT ["/app/service"]
