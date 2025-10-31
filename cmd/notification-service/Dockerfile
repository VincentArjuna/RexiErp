# Build stage
FROM golang:1.23.1-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata

# Set the working directory
WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o notification-service ./cmd/notification-service

# Final stage
FROM alpine:latest

# Install runtime dependencies
RUN apk --no-cache add ca-certificates tzdata wget

# Set timezone to Asia/Jakarta for Indonesian business context
ENV TZ=Asia/Jakarta
RUN ln -snf /usr/share/zoneinfo/$TZ /etc/localtime && echo $TZ > /etc/timezone

# Create a non-root user
RUN addgroup -g 1001 -S rexi && \
    adduser -u 1001 -S rexi -G rexi

WORKDIR /app

# Copy the binary from builder stage
COPY --from=builder /app/notification-service .

# Copy configuration files if they exist
COPY --from=builder /app/configs ./configs

# Change ownership to rexi user
RUN chown -R rexi:rexi /app

# Switch to non-root user
USER rexi

# Expose port
EXPOSE 8006

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8006/health || exit 1

# Run the binary
CMD ["./notification-service"]