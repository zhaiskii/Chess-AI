# Chess AI Backend V3 - Production Dockerfile

# Build stage
FROM golang:1.22-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata

RUN apk add --no-cache curl

# Set working directory
WORKDIR /app

# Copy go mod files first for better Docker layer caching
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download && go mod verify

# Copy source code
COPY . .

# Build the application with optimizations
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags='-w -s -extldflags "-static"' \
    -a -installsuffix cgo \
    -o chess-ai .

# Production stage
FROM alpine:latest

# Install ca-certificates for HTTPS requests
RUN apk --no-cache add ca-certificates tzdata wget

# Create app directory
WORKDIR /app

# Create non-root user
RUN addgroup -g 1001 -S appgroup && \
    adduser -u 1001 -S appuser -G appgroup

# Copy the binary from builder stage
COPY --from=builder /app/chess-ai .

# Change ownership
RUN chown appuser:appgroup chess-ai

# Switch to non-root user
USER appuser

# Expose port
EXPOSE 8080

# Add health check
# HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
#     CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

# Set environment variables
ENV PORT=8080

# Run the application
CMD ["./chess-ai"]