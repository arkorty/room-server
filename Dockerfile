# Stage 1: Build the Go binary
FROM golang:alpine AS builder

# Set up dependencies and install SQLite development files
RUN apk add --no-cache build-base git sqlite-dev

# Set the working directory
WORKDIR /server

# Copy the Go module files and download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy the source code
COPY . .

# Build the server binary with CGO enabled
RUN CGO_ENABLED=1 go build -o server server.go

# Stage 2: Run the server
FROM alpine:latest

# Copy the server binary from the builder stage
COPY --from=builder /server/server /server

# Expose the port
EXPOSE 8080

# Run the server
CMD ["/server"]
