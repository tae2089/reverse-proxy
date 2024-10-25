# Start with the official Golang image to build the binary
FROM golang:1.20-alpine AS base
ARG VERSION
# Set the Current Working Directory inside the container
WORKDIR /app
# Use --mount to cache dependencies
RUN --mount=type=cache,target=/root/.cache/go-build \
    --mount=type=bind,source=.,target=/app \
    go mod download

# Build the Go app
FROM base AS builder
RUN --mount=type=bind,source=.,target=/app \
    go build -o reverse-proxy -ldflags "-s -w -X github.com/tae2089/reverse-proxy/internal/server.VERSION=${VERSION}" .

# Start a new stage from scratch
FROM scratch
# Copy the binary from the builder stage
COPY --from=builder /app/reverse-proxy /reverse-proxy
# Expose port 8080 to the outside world
EXPOSE 8080
# Command to run the executable
CMD ["/reverse-proxy"]