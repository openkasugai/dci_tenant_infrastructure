FROM docker.io/golang:1.25 AS builder
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod edit -go=1.25 && grep '^go ' go.mod

# Clean module cache to ensure fresh download
RUN go clean -modcache

# Copy only required directories for workflow
COPY shared/ ./shared/
COPY api/ ./api/
COPY workflow/ ./workflow/

# Generate go.sum for K8s 1.34
RUN go mod tidy

# 2. Format check
RUN go fmt ./workflow/... ./shared/... ./api/...

# 3. Vet check (with -mod=mod to auto-download missing dependencies)
RUN go vet -mod=mod ./workflow/... ./shared/... ./api/...

# 4. Build (with -mod=mod to auto-download missing dependencies)
RUN go build -mod=mod -o worker ./workflow

FROM gcr.io/distroless/base
WORKDIR /app
COPY --from=builder /app/worker /app/worker

ENTRYPOINT ["/app/worker"]
