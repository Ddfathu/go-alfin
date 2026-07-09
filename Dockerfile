FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY main.go .
RUN go mod init turbo-tunnel && go mod tidy && go build -o turbo-proxy main.go

FROM alpine:latest
# 🔥 INSTAL OPENSSH (Gantiin Dropbear yang kaku)
RUN apk add --no-cache bash openssh openssh-server-pam stunnel openssl shadow wget

# Download biner cloudflared murni (Udah terbukti lancar kilat)
RUN wget https://github.com/cloudflare/cloudflared/releases/latest/download/cloudflared-linux-amd64 -O /usr/local/bin/cloudflared && \
    chmod +x /usr/local/bin/cloudflared

# Buat folder sistem yang diperlukan OpenSSH dan Stunnel
RUN mkdir -p /var/run/sshd /etc/stunnel /var/run /usr/bin

# Salin semua file skrip menu lu yang berserakan di luar
COPY menu addssh delssh listssh /usr/bin/
RUN chmod +x /usr/bin/menu /usr/bin/addssh /usr/bin/delssh /usr/bin/listssh

COPY --from=builder /app/turbo-proxy /usr/local/bin/turbo-proxy
COPY entrypoint.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh

ENTRYPOINT ["/entrypoint.sh"]
