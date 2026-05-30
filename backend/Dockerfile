# Debian 运行时 + CGO：使用系统 CA 证书连接 Atlas（Alpine 静态编译在 Render 上易 TLS 失败）
FROM golang:1.24-bookworm AS builder

WORKDIR /src
RUN apt-get update && apt-get install -y --no-install-recommends gcc libc6-dev ca-certificates \
    && rm -rf /var/lib/apt/lists/*

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=1 GOOS=linux go build -ldflags="-s -w" -o /server ./cmd/server

FROM debian:bookworm-slim

RUN apt-get update && apt-get install -y --no-install-recommends ca-certificates \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /app
COPY --from=builder /server ./server

ENV SSL_CERT_FILE=/etc/ssl/certs/ca-certificates.crt
ENV PORT=8080
EXPOSE 8080

CMD ["./server"]
