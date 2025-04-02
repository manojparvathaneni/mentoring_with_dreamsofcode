# Build stage
FROM golang:1.20-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git

# Set working directory
WORKDIR /app

# Copy go.mod and go.sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o counter-service ./cmd/server

# Final stage
FROM alpine:3.17

# Add non-root user
RUN addgroup -S app && adduser -S app -G app

# Install runtime dependencies
RUN apk add --no-cache ca-certificates tzdata

# Set working directory
WORKDIR /app

# Copy binary from builder stage
COPY --from=builder /app/counter-service .

# Copy configuration
COPY --from=builder /app/config.yaml .

# Create data directory and set permissions
RUN mkdir -p /app/data && chown -R app:app /app

# Use non-root user
USER app

# Expose port
EXPOSE 8090

# Set environment variables
ENV COUNTER_ENVIRONMENT=production

# Run the application
CMD ["./counter-service"]