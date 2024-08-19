# Use the official Golang image as the base image
FROM golang:1.22.6 AS builder

# Create a non-root user for running the application
RUN useradd -u 1001 nonroot

# Set the Current Working Directory inside the container
WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download all dependencies. Dependencies will be cached if the go.mod and go.sum files are not changed
RUN go mod download

# Copy the source code into the container
COPY . .

# Run tests
RUN go test ./... -v

# Build the Go app
RUN go build  \
    -ldflags="-linkmode external -extldflags -static" \
    -tags netgo \
    -o main ./cmd/indexer

# Start a new stage from scratch
FROM alpine:latest

# Copy the /etc/passwd file from the build stage to provide non-root user information
COPY --from=builder /etc/passwd /etc/passwd


# Install necessary packages and glibc compatibility
RUN apk --no-cache add ca-certificates \
    && apk --no-cache add libc6-compat

# Copy the Pre-built binary file from the previous stage
COPY --from=builder /app/main /main

# Use the non-root user
USER nonroot

# Expose port 8080 to the outside world
EXPOSE 8080

# Command to run the executable
CMD ["./main"]
