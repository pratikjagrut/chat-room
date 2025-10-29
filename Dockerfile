# Stage 1: Build the Go application
FROM golang:1.25-alpine AS builder

# Set the working directory inside the container
WORKDIR /app

# Copy go.mod and go.sum to leverage Docker cache
COPY go.mod go.sum ./
# Download dependencies
RUN go mod download

# Copy the rest of the source code
COPY . .

# Build the Go app for a linux environment
# CGO_ENABLED=0 creates a static binary
# -ldflags="-w -s" strips debug information to reduce binary size
RUN CGO_ENABLED=0 GOOS=linux go build -a -ldflags="-w -s" -o /app/main .

# Stage 2: Create the final, lightweight image
FROM alpine:latest

# Set the working directory
WORKDIR /app

# Copy the static assets
COPY --chown=nobody:nogroup static ./static

# Copy the compiled binary from the builder stage
COPY --from=builder --chown=nobody:nogroup /app/main .

# Expose the port the app runs on
EXPOSE 8080

# Run as a non-root user for better security
USER nobody:nogroup

# Define the command to run the application
CMD ["./main"]
