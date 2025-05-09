# Build stage
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git gcc musl-dev

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=1 GOOS=linux go build -o /app/bin/devtrackr ./cmd/devtrackr

# Final stage
FROM alpine:latest

WORKDIR /app

# Install runtime dependencies
RUN apk add --no-cache sqlite

# Copy the binary from builder
COPY --from=builder /app/bin/devtrackr /app/
COPY --from=builder /app/assets /app/assets

# Expose the application port
EXPOSE 8080

# Run the application
CMD ["/app/devtrackr", "serve"]