# Use the official Golang image as a base image
FROM golang:latest AS builder

# Set the working directory inside the container
WORKDIR /build

# Copy the current directory contents into the container at /build
COPY . .

# Download and install any dependencies required by go.mod files in all subdirectories.
# Run `go mod tidy` in cmd/scanner/ to update go.mod if needed.
RUN go mod tidy
RUN go mod download

# Build the Go application
# Without CGO_ENABLED=0, running the binary will fail with "exec ./scanner: no such file or directory"
RUN CGO_ENABLED=0 GOOS=linux go build -o /bin/scanner ./cmd/scanner

# Start a new stage from scratch
FROM alpine:latest

# Set the working directory inside the container
WORKDIR /app

RUN apk add --no-cache nmap

# Copy the built executable from the previous stage
COPY --from=builder /bin/scanner ./scanner

# Run the scanner binary
ENTRYPOINT ["./scanner"]
