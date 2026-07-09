FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY main.go .
# 🛠️ KUNCI SUKSES: Tambahkan 'go mod tidy' sebelum di-build sesuai kemauan sistemnya
RUN go mod init turbo-tunnel && go mod tidy && go build -o turbo-proxy main.go

FROM alpine:latest
# Instal bash, dropbear, stunnel, openssl, cloudflared, dan shadow untuk menu
RUN apk add --no-cache bash dropbear stunnel openssl cloudflared shadow

# Buat folder sistem yang diperlukan
RUN mkdir -p /etc/dropbear /etc/stunnel /var/run /usr/bin

# Salin skrip menu dari folder github lu ke sistem biner Alpine
COPY menu/* /usr/bin/
RUN chmod +x /usr/bin/*

COPY --from=builder /app/turbo-proxy /usr/local/bin/turbo-proxy
COPY entrypoint.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh

ENTRYPOINT ["/entrypoint.sh"]