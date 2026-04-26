FROM golang:1.26.1-alpine AS build
WORKDIR /src/gateway

# Needed for replace jobconnect/auth => ../services/auth in gateway/go.mod.
COPY services/auth /src/services/auth
COPY api /src/api
COPY services/job /src/services/job
COPY services/contract /src/services/contract
COPY services/dispute /src/services/dispute
COPY services/proposal /src/services/proposal
COPY services/recommendation /src/services/recommendation
COPY services/user /src/services/user
COPY services/verification /src/services/verification
COPY services/chat /src/services/chat
COPY gateway/go.mod gateway/go.sum ./
RUN go mod download

COPY gateway/. .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/gatewayd ./cmd/gatewayd

FROM alpine:3.22
WORKDIR /app
COPY --from=build /out/gatewayd /app/gatewayd
EXPOSE 8080
ENTRYPOINT ["/app/gatewayd"]
