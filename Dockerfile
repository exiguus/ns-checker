FROM golang:1.21-alpine AS builder

WORKDIR /app
COPY . .
RUN go build -o ns-checker

FROM alpine:latest

RUN apk add --no-cache tini libcap curl sudo

WORKDIR /app
COPY --from=builder /app/ns-checker /app/ns-checker

# Set proper permissions and capabilities
RUN chmod +x /app/ns-checker && \
    setcap cap_net_bind_service=+ep /app/ns-checker

ENV IN_CONTAINER=true

ENTRYPOINT ["/sbin/tini", "--"]
CMD ["/app/ns-checker", "listen"]
