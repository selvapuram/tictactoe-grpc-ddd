# Dockerfile
FROM golang:1.21-alpine AS builder

# Install protobuf compiler
RUN apk add --no-cache protobuf protobuf-dev

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Generate protobuf code and build
RUN go install google.golang.org/protobuf/cmd/protoc-gen-go@latest && \
    go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
RUN make proto
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o tictactoe-server ./cmd/server

# Final stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates
WORKDIR /root/

# Copy the binary from builder stage
COPY --from=builder /app/tictactoe-server .

# Expose port
EXPOSE 8080

# Command to run
CMD ["./tictactoe-server"]

