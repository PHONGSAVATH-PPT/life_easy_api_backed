# syntax=docker/dockerfile:1

FROM golang:1.26.2-alpine AS builder

WORKDIR /src

RUN apk add --no-cache ca-certificates git

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux \
    go build \
    -mod=readonly \
    -trimpath \
    -ldflags="-s -w" \
    -o /out/api \
    ./cmd/api


FROM alpine:3.22

RUN apk add --no-cache ca-certificates tzdata \
    && addgroup -g 10001 app \
    && adduser -D -u 10001 -G app app

WORKDIR /app

COPY --from=builder /out/api /app/api
COPY --from=builder /src/docs /app/docs

RUN mkdir -p /app/logs \
    && chown -R app:app /app

USER 10001:10001

ENV TZ=Asia/Vientiane

EXPOSE 8080

HEALTHCHECK \
    --interval=15s \
    --timeout=3s \
    --start-period=15s \
    --retries=3 \
    CMD ["/app/api", "healthcheck"]

ENTRYPOINT ["/app/api"]