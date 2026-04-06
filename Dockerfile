FROM golang:1.24-bookworm AS builder

WORKDIR /build

RUN apt-get update && apt-get install -y --no-install-recommends \
    git gcc libsqlite3-dev \
    && rm -rf /var/lib/apt/lists/*

COPY go.mod go.sum ./
COPY vendor ./vendor

COPY . .

RUN CGO_ENABLED=1 GOOS=linux go build -a -installsuffix cgo \
    -ldflags="-w -s" \
    -mod=vendor \
    -o tinyrsvp ./cmd/server

FROM debian:bookworm-slim

RUN apt-get update && apt-get install -y --no-install-recommends \
    ca-certificates libsqlite3-0 wget tzdata \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /app

COPY --from=builder /build/tinyrsvp .
COPY --from=builder /build/migrations ./migrations
COPY --from=builder /build/templates ./templates
COPY --from=builder /build/static ./static

RUN mkdir -p /data

EXPOSE 8080

HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

RUN groupadd -g 1000 tinyrsvp && \
    useradd -D -u 1000 -g tinyrsvp tinyrsvp && \
    chown -R tinyrsvp:tinyrsvp /app /data

USER tinyrsvp

CMD ["./tinyrsvp"]
