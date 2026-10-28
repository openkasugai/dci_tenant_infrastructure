FROM golang:1.26.2 AS builder
WORKDIR /app

COPY go.mod go.sum ./

# Clean module cache to ensure fresh download
RUN go clean -modcache

# Copy only required directories for client
COPY shared/ ./shared/
COPY gateway/ ./gateway/

RUN go mod tidy

# 2. Format check
RUN go fmt ./gateway/... ./shared/...

# 3. Vet check (with -mod=mod to auto-download missing dependencies)
RUN go vet -mod=mod ./gateway/... ./shared/...

# 4. Build (with -mod=mod to auto-download missing dependencies)
RUN go build -mod=mod -o signal-gateway ./gateway

FROM gcr.io/distroless/base
WORKDIR /app
COPY --from=builder /app/signal-gateway /app/signal-gateway

ENTRYPOINT ["/app/signal-gateway"]
