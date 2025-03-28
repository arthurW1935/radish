# Start with the official Go image
FROM golang:1.24.1 AS builder

# Set the working directory inside the container
WORKDIR /app

# Copy the Go module files and download dependencies
COPY go.mod ./
RUN go mod download

# Copy the application source code
COPY . .

# Build the application (creates a binary)
RUN go build -o server ./cmd/server

# Use a smaller final image (distroless or alpine for smaller size)
FROM gcr.io/distroless/base-debian12

# Set the working directory
WORKDIR /app

# Copy the built binary from the builder stage
COPY --from=builder /app/server .

# Expose the port (change this if needed)
EXPOSE 7171

# Run the server binary
CMD ["./server"]
