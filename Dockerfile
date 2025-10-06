# --- Build Stage ---
# Use a specific version of Go for reproducibility
FROM golang:1.22-alpine AS builder

# Set the working directory inside the container
WORKDIR /app

# Copy go.mod and go.sum to leverage Docker cache
COPY go.mod go.sum ./
# Download dependencies
RUN go mod download

# Copy the rest of the source code
COPY . .

# Build the Go application for a Linux environment
# CGO_ENABLED=0 creates a static binary
# -ldflags="-w -s" strips debug information, reducing binary size
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /phinex-blog-api ./main.go

# --- Final Stage ---
# Use a minimal base image
FROM alpine:latest

# Set the working directory inside the container
WORKDIR /app

# Copy the built binary from the builder stage
COPY --from=builder /phinex-blog-api .

# Copy swagger docs
COPY docs ./docs

# Expose the port the application will run on.
# This should match the PORT variable in the .env file.
EXPOSE 5000

# The command to run when the container starts
CMD ["/app/phinex-blog-api"]