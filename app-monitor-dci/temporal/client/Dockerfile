FROM golang:1.25 AS builder
WORKDIR /app

COPY go.mod go.sum ./

# Clean module cache to ensure fresh download
RUN go clean -modcache

# Copy only required directories for client
COPY shared/ ./shared/
COPY client/ ./client/

RUN go mod tidy

# 2. Format check
RUN go fmt ./client/... ./shared/...

# 3. Vet check (with -mod=mod to auto-download missing dependencies)
RUN go vet -mod=mod ./client/... ./shared/...

# 4. Build (with -mod=mod to auto-download missing dependencies)
RUN go build -mod=mod -o client-app ./client

FROM gcr.io/distroless/base
WORKDIR /app
COPY --from=builder /app/client-app /app/client

ENTRYPOINT ["/app/client"]
