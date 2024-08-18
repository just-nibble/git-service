# Use the official Golang image as the base image
FROM golang:1.22.6 AS builder

# Set the Current Working Directory inside the container
WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download all dependencies. Dependencies will be cached if the go.mod and go.sum files are not changed
RUN go mod download

# Copy the source code into the container
COPY . .

# Build the Go app
RUN go build -o main ./cmd/indexer

# Start a new stage from scratch
FROM alpine:latest

# Install necessary packages and glibc compatibility
RUN apk --no-cache add ca-certificates \
    && apk --no-cache add libc6-compat

# Set the Current Working Directory inside the container
WORKDIR /root/

# Copy the Pre-built binary file from the previous stage
COPY --from=builder /app/main .

# Expose port 8080 to the outside world
EXPOSE 8080

# Command to run the executable
CMD ["./main"]
