FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY main.go .
RUN go build -o turbo-proxy main.go

FROM alpine:latest
# Instal bash, dropbear, stunnel, openssl, cloudflared, PLUS paket 'shadow' biar perintah manajemen user lebih mirip Ubuntu
RUN apk add --no-cache bash dropbear stunnel openssl cloudflared shadow

# Buat folder konfigurasi yang dibutuhkan
RUN mkdir -p /etc/dropbear /etc/stunnel /var/run /usr/bin

# 🔥 PROSES COPY MENU: Salin semua file skrip menu lu ke biner sistem
# Pastikan folder 'menu' berisi file addssh, delssh, dll ada di github lu
COPY menu/* /usr/bin/
RUN chmod +x /usr/bin/*

COPY --from=builder /app/turbo-proxy /usr/local/bin/turbo-proxy
COPY entrypoint.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh

ENTRYPOINT ["/entrypoint.sh"]
