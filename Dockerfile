FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY main.go .
RUN go mod init turbo-tunnel && go mod tidy && go build -o turbo-proxy main.go

FROM alpine:latest
# 1. Instal semua paket yang VALIDE & ADA di Alpine (Tanpa cloudflared)
RUN apk add --no-cache bash dropbear stunnel openssl shadow wget

# 2. DOWNLOAD BINER CLOUDFLARED LANGSUNG SECARA MANUAL
# Kita ambil versi linux-amd64 resmi, lalu pindahkan ke /usr/local/bin
RUN wget https://github.com/cloudflare/cloudflared/releases/latest/download/cloudflared-linux-amd64 -O /usr/local/bin/cloudflared && \
    chmod +x /usr/local/bin/cloudflared

# Buat folder sistem yang diperlukan
RUN mkdir -p /etc/dropbear /etc/stunnel /var/run /usr/bin

# Salin skrip menu dari folder github lu ke sistem biner Alpine
COPY menu/* /usr/bin/
RUN chmod +x /usr/bin/*

COPY --from=builder /app/turbo-proxy /usr/local/bin/turbo-proxy
COPY entrypoint.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh

ENTRYPOINT ["/entrypoint.sh"]