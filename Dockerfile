# Stage 1: Build the backend binary
FROM golang:1.24-alpine AS go-builder
WORKDIR /build

# Alpine dependencies required for some Go tools/CGO (like scrypt)
RUN apk add --no-cache gcc musl-dev

# Copy Go modules and install dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy source code and build
COPY . .
RUN go build -o spectre-updated

# Stage 2: Final lightweight image
FROM alpine:latest
WORKDIR /app

# Add required dependencies for the app runtime
RUN apk add --no-cache ca-certificates tzdata

# Copy the built Go binary from stage 1
COPY --from=go-builder /build/spectre-updated .
COPY --from=go-builder /build/public ./public
COPY --from=go-builder /build/templates ./templates

# Expose the application port
EXPOSE 8080

# Command to run the executable
CMD ["./spectre-updated", "-log_dir=logs", "-root=data", "--logtostderr=1"]
