# Build stage
FROM golang:1.21-alpine AS builder

# Install build dependencies
RUN apk add --no-cache gcc musl-dev sqlite-dev

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=1 GOOS=linux go build -a -installsuffix cgo -o tunetudo main.go

# Runtime stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates sqlite-libs

WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/tunetudo .

# Create storage directories
RUN mkdir -p storage/media/songs storage/media/uploads storage/images/profiles storage/images/covers

# Expose port
EXPOSE 2701

# Run the application
CMD ["./tunetudo"]