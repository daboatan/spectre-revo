FROM golang:1.24-alpine AS go-builder
WORKDIR /build

RUN apk add --no-cache gcc musl-dev

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /build/spectre-revo

FROM alpine:3.20
WORKDIR /app

RUN apk add --no-cache ca-certificates tzdata \
    && adduser -D -h /app appuser \
    && mkdir -p /app/data /app/public /app/templates /app/logs \
    && chown -R appuser:appuser /app

COPY --from=go-builder /build/spectre-revo /app/spectre-revo
COPY --from=go-builder /build/public /app/public
COPY --from=go-builder /build/templates /app/templates
COPY --from=go-builder /build/languages.yml /app/languages.yml

USER appuser

EXPOSE 8080

CMD ["./spectre-revo","-log_dir=/app/logs","-root=/app/data","--logtostderr=1"]