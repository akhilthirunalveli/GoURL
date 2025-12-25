FROM golang:1.21-alpine AS builder

WORKDIR /app

# Copy dependencies first
COPY go.mod go.sum ./

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -mod=mod -o /app/bin/server ./cmd/server

# Runtime stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copy binary from builder
COPY --from=builder /app/bin/server .

EXPOSE 8080

CMD ["./server"]
