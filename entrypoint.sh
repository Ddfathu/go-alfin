#!/bin/bash

USER_NAME="${SSH_USER:-dd}"
USER_PASS="${SSH_PASSWORD:-dd}"
PUBLIC_PORT="${PORT:-8080}"
SSL_INTERNAL_PORT="${SSL_INTERNAL_PORT:-2443}"
WS_INTERNAL_PORT="${WS_INTERNAL_PORT:-8880}"

# =====================================================================
# 🔥 FIX RESEP ALPINE: Pahat Host Keys Dropbear Manual (Biar Mau Bangun)
# =====================================================================
echo "[*] Memeriksa dan Membuat Host Keys Dropbear..."
if [ ! -f /etc/dropbear/dropbear_rsa_host_key ]; then
    dropbearkey -t rsa -f /etc/dropbear/dropbear_rsa_host_key -s 2048
fi
if [ ! -f /etc/dropbear/dropbear_ed25519_host_key ]; then
    dropbearkey -t ed25519 -f /etc/dropbear/dropbear_ed25519_host_key
fi

echo "[*] Mengonfigurasi Server Message Dropbear (Banner)..."
cat << 'EOF' > /etc/dropbear_banner
=================================================
                  SELAMAT MENIKMATI
             PREMIUM SSH SERVER DROPBEAR modssh        
=================================================
       Dilarang Torrent / DDOS / Hacking! 
                 Powered By: dedefathu
=================================================
EOF

echo "[*] Mengonfigurasi Respon Server (Pasca-Login)..."
# Di Alpine, file profile ditaruh di /etc/profile.d/
mkdir -p /etc/profile.d
cat << 'EOF' > /etc/profile.d/99-respon-server.sh
#!/bin/bash
clear
echo -e "\e[1;36m=================================================\e[0m"
echo -e "\e[1;32m       [✓] BERHASIL TERHUBUNG KE SERVER!         \e[0m"
echo -e "\e[1;36m=================================================\e[0m"
echo -e "\e[1;37m Username     : \e[1;33m$USER\e[0m"
echo -e "\e[1;37m Waktu Server : \e[1;33m$(date)\e[0m"
echo -e "\e[1;37m OS           : \e[1;33mAlpine Linux (Turbo Mode)\e[0m"
echo -e "\e[1;36m=================================================\e[0m"
echo -e "\e[1;31m   TETAP PATUHI RULES SERVER AGAR TIDAK BANNED   \e[0m"
echo -e "\e[1;36m=================================================\e[0m"
EOF
chmod +x /etc/profile.d/99-respon-server.sh

echo "[*] Mengonfigurasi User SSH..."
# Perintah adduser di Alpine agak beda dengan Ubuntu, kita pakai flags Alpine murni
if ! id "$USER_NAME" &>/dev/null; then
    adduser -D -s /bin/bash "$USER_NAME"
fi
echo "$USER_NAME:$USER_PASS" | chpasswd

echo "[*] Memulai Dropbear Server di Port Lokal 22..."
/usr/sbin/dropbear -p 127.0.0.1:22 -b /etc/dropbear_banner -W 65536

# 🔥 TAMBAHAN KESELAMATAN: Buat Sertifikat SSL Stunnel
echo "[*] Membuat Sertifikat SSL Stunnel..."
openssl req -new -newkey rsa:2048 -days 365 -nodes -x509 \
    -subj "/C=ID/ST=Jakarta/L=Jakarta/O=RailwaySSH/CN=localhost" \
    -keyout /etc/stunnel/stunnel.pem -out /etc/stunnel/stunnel.pem

echo "[*] Mengonfigurasi Stunnel internal di Port $SSL_INTERNAL_PORT..."
cat <<EOF > /etc/stunnel/stunnel.conf
pid = /var/run/stunnel.pid
foreground = yes
debug = 4

[ssh-ssl]
accept = 127.0.0.1:$SSL_INTERNAL_PORT
connect = 127.0.0.1:22
cert = /etc/stunnel/stunnel.pem
EOF

echo "[*] Menambahkan alias ke .bashrc..."
cat <<'EOF'>> ~/.bashrc
clear
alias c='clear'
alias x='exit'
alias +x='chmod +x'
alias cls='clear;ls'
EOF

# Di Alpine, namanya murni 'stunnel', bukan 'stunnel4'
echo "[*] Memulai Stunnel (internal, port $SSL_INTERNAL_PORT)..."
stunnel /etc/stunnel/stunnel.conf &

if [ -n "$CF_TUNNEL_TOKEN" ]; then
    echo "[*] Menjalankan Cloudflare Tunnel (Argo)..."
    cloudflared tunnel run --token "$CF_TUNNEL_TOKEN" &
else
    echo "[!] CF_TUNNEL_TOKEN kosong -> Cloudflare Tunnel dilewati."
fi

echo "[*] Memulai GOLANG TURBO TUNNEL ENGINE di Port PUBLIK $PUBLIC_PORT..."
exec env \
    PORT="$PUBLIC_PORT" \
    SSL_TARGET_HOST="127.0.0.1" \
    SSL_TARGET_PORT="$SSL_INTERNAL_PORT" \
    WS_TARGET_HOST="127.0.0.1" \
    WS_TARGET_PORT="22" \
    turbo-proxy
