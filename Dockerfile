# syntax=docker/dockerfile:1

### Build Stage ###
FROM golang:1.23-alpine AS builder

# Install Git and other necessary packages
RUN apk add --no-cache git

WORKDIR /app

# Cache dependency installation
COPY go.mod go.sum ./
RUN go mod tidy && go mod download

# Copy the full source code
COPY . .

# Build the Go binary (using CGO disabled for a static binary)
RUN CGO_ENABLED=0 GOOS=linux go build -o server ./cmd/server/main.go

### Final Stage ###
FROM alpine:latest

# Install CA certificates (for HTTPS, etc.)
RUN apk --no-cache add ca-certificates

WORKDIR /app

# Copy the built binary from the builder stage
COPY --from=builder /app/server .

COPY deployments/docker/.env /app/.env


# Expose the application port (adjust if your app listens on a different port)
EXPOSE 8080

# Use a non-root user in production (optional but recommended)
# RUN addgroup -S appgroup && adduser -S appuser -G appgroup
# USER appuser

# Start the application
ENTRYPOINT ["./server"]