FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY main.go .
# 🔥 AMAN: Inisialisasi modul Go biar library crypto & encoding-nya ke-compile sempurna
RUN go mod init turbo-tunnel && go build -o turbo-proxy main.go

FROM alpine:latest
# Instal bash, dropbear, stunnel, openssl, cloudflared, dan shadow untuk menu
RUN apk add --no-cache bash dropbear stunnel openssl cloudflared shadow

# Buat folder sistem yang diperlukan
RUN mkdir -p /etc/dropbear /etc/stunnel /var/run /usr/bin

# Salin skrip menu dari folder github lu ke sistem biner Alpine
# *Catatan: pastikan folder 'menu' berisi file addssh dll ada di repo lu*
COPY menu/* /usr/bin/
RUN chmod +x /usr/bin/*

COPY --from=builder /app/turbo-proxy /usr/local/bin/turbo-proxy
COPY entrypoint.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh

ENTRYPOINT ["/entrypoint.sh"]