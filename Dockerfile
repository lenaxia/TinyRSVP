FROM golang:1.24-alpine AS builder

WORKDIR /build

RUN apk add --no-cache git gcc musl-dev sqlite-dev

COPY go.mod go.sum ./
COPY vendor ./vendor

COPY . .

RUN CGO_ENABLED=1 GOOS=linux go build -a -installsuffix cgo \
    -ldflags="-w -s" \
    -mod=vendor \
    -o tinyrsvp ./cmd/server

FROM alpine:latest

RUN apk --no-cache add ca-certificates sqlite-libs wget

WORKDIR /app

COPY --from=builder /build/tinyrsvp .
COPY --from=builder /build/migrations ./migrations
COPY --from=builder /build/templates ./templates
COPY --from=builder /build/static ./static

RUN mkdir -p /data

EXPOSE 8080

HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

RUN addgroup -g 1000 tinyrsvp && \
    adduser -D -u 1000 -G tinyrsvp tinyrsvp && \
    chown -R tinyrsvp:tinyrsvp /app /data

USER tinyrsvp

CMD ["./tinyrsvp"]
