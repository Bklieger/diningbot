# Build stage
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the MCP server binary
RUN go build -o diningbot .

# Runtime stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata wget

WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/diningbot .

# Expose MCP Streamable HTTP port
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:8080/ || exit 1

# Run the MCP server in remote mode
# Set PORT environment variable to change the port (default: 8080)
# Set BIND_ADDR to 0.0.0.0 to allow external connections (required for Docker)
ENV PORT=8080
ENV BIND_ADDR=0.0.0.0
CMD ["./diningbot"]

