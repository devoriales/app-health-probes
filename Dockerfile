# Stage 1: Build the Go application
FROM golang:1.23.4 as builder

# Set environment variables for cross-compilation
ENV CGO_ENABLED=0
ENV GOOS=linux
ENV GOARCH=amd64

WORKDIR /app

# Initialize the Go module and download dependencies
COPY . .

# Check if go.mod exists; if not, initialize it
RUN [ ! -f go.mod ] && go mod init k8s-probe || echo "go.mod already exists"

# Download dependencies
RUN go mod tidy

# Build the Go application
RUN go build -o main .

# Stage 2: Create a lightweight image
FROM alpine:latest

WORKDIR /root/

# make dir for /tmp/startup-file
# RUN mkdir -p /tmp/startup-file

# Copy the binary from the builder stage
COPY --from=builder /app/main .

# Expose necessary ports
EXPOSE 8080

# Command to run the application
CMD ["./main"]
