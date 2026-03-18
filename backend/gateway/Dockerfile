FROM golang:1.25.1-alpine AS build
WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/gatewayd ./cmd/gatewayd

FROM alpine:3.22
WORKDIR /app
COPY --from=build /out/gatewayd /app/gatewayd
EXPOSE 8080
ENTRYPOINT ["/app/gatewayd"]
