# Build stage
FROM golang:1.21-alpine AS builder

# Install git and build dependencies
RUN apk add --no-cache git ca-certificates tzdata

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod ./
COPY go.sum* ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main .

# Final stage
FROM alpine:latest

# Install ca-certificates for HTTPS requests
RUN apk --no-cache add ca-certificates tzdata

# Create non-root user
RUN adduser -D -s /bin/sh cloudbox

# Set working directory
WORKDIR /app

# Copy binary and wait script from builder stage
COPY --from=builder /app/main .
COPY --from=builder /app/wait-for-services.sh .
RUN chmod +x wait-for-services.sh

# Create uploads directory
RUN mkdir -p uploads && chown cloudbox:cloudbox uploads

# Change ownership to non-root user
RUN chown cloudbox:cloudbox /app/main

# Switch to non-root user
USER cloudbox

# Expose port
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

# Run the application directly (skip wait script for now)
CMD ["./main"]