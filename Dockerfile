# Stage 1: Build the backend binary
FROM golang:1.24-alpine AS go-builder
WORKDIR /build

# Only keep this if your app truly needs CGO
RUN apk add --no-cache gcc musl-dev

COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Safer for Alpine runtime
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /build/spectre-updated

# Stage 2: Final lightweight image
FROM alpine:3.20
WORKDIR /app

RUN apk add --no-cache ca-certificates tzdata \
    && adduser -D -h /app appuser \
    && mkdir -p /app/data /app/public /app/templates /app/logs \
    && chown -R appuser:appuser /app

COPY --from=go-builder /build/spectre-updated /app/spectre-updated
COPY --from=go-builder /build/public /app/public
COPY --from=go-builder /build/templates /app/templates
COPY --from=go-builder /build/languages.yml /app/languages.yml

USER appuser

EXPOSE 8080

CMD ["./spectre-updated","-log_dir=/app/logs","-root=/app/data","--logtostderr=1"]