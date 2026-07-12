#!/bin/bash

USER_NAME="${SSH_USER:-dd}"
USER_PASS="${SSH_PASSWORD:-dd}"
PUBLIC_PORT="${PORT:-8080}"
SSL_INTERNAL_PORT="${SSL_INTERNAL_PORT:-2443}"
WS_INTERNAL_PORT="${WS_INTERNAL_PORT:-8880}"

# =====================================================================
# 🔥 SYSCTL ULTRA HIGH SPEED TWEAK (MAKAN RAM SPEK BADAK)
# =====================================================================
echo "[*] Menyuntikkan Tweak Network Buffer High Speed..."

# Aktifkan BBR Congestion Control jika didukung kernel Alpine
modprobe tcp_bbr 2>/dev/null
echo "net.core.default_qdisc=fq" >> /etc/sysctl.conf
echo "net.ipv4.tcp_congestion_control=bbr" >> /etc/sysctl.conf

# Alokasi Maksimal Buffer Jaringan (Makan RAM demi Speed Mentok Kanan)
echo "net.core.rmem_max=67108864" >> /etc/sysctl.conf
echo "net.core.wmem_max=67108864" >> /etc/sysctl.conf
echo "net.core.rmem_default=33554432" >> /etc/sysctl.conf
echo "net.core.wmem_default=33554432" >> /etc/sysctl.conf

# TCP Window Tuning & Buffer Memory (Min, Default, Max dalam Bytes)
echo "net.ipv4.tcp_rmem=4096 87380 67108864" >> /etc/sysctl.conf
echo "net.ipv4.tcp_wmem=4096 65536 67108864" >> /etc/sysctl.conf

# Naikkan Batas Antrean Paket (Anti-Drop & Bebas Los saat Tethering Brutal)
echo "net.core.netdev_max_backlog=10000" >> /etc/sysctl.conf
echo "net.ipv4.tcp_max_syn_backlog=8192" >> /etc/sysctl.conf
echo "net.ipv4.tcp_tw_reuse=1" >> /etc/sysctl.conf
echo "net.ipv4.tcp_fin_timeout=15" >> /etc/sysctl.conf
echo "net.ipv4.tcp_keepalive_time=60" >> /etc/sysctl.conf

# Terapkan konfigurasi sysctl baru
sysctl -p /etc/sysctl.conf 2>/dev/null

# =====================================================================
# 🔥 SETUP OPENSSH: Pahat Host Keys & Buka Parameter Enkripsi Ringan CPU
# =====================================================================
echo "[*] Membuat Host Keys OpenSSH..."
ssh-keygen -A

# 🎨 BANNER WARNA-WARNI & CENTER LOGIC UNTUK LOGIN TULISAN
echo "[*] Mengonfigurasi Banner SSH..."
cat << 'EOF' > /etc/ssh_banner
=================================================
                  SELAMAT MENIKMATI
      👑 PREMIUM SSH SERVER OPENSSH goalfin 👑   
=================================================
       Dilarang Torrent / DDOS / Hacking! 
          👑 PRIVATE TUNNEL BY: DEDEFATHU 👑
=================================================
EOF

echo "[*] Mengonfigurasi Respon Server (Pasca-Login)..."
mkdir -p /etc/profile.d
cat << 'EOF' > /etc/profile.d/99-respon-server.sh
#!/bin/bash
clear
echo -e "\e[1;36m=================================================\e[0m"
echo -e "\e[1;32m       [✓] BERHASIL TERHUBUNG KE SERVER!         \e[0m"
echo -e "\e[1;36m=================================================\e[0m"
echo -e "\e[1;37m Username     : \e[1;33m$USER\e[0m"
echo -e "\e[1;37m Waktu Server : \e[1;33m$(date)\e[0m"
echo -e "\e[1;37m OS           : \e[1;33mAlpine Linux (OpenSSH Turbo)\e[0m"
echo -e "\e[1;36m=================================================\e[0m"
echo -e "\e[1;31m   TETAP PATUHI RULES SERVER AGAR TIDAK BANNED   \e[0m"
echo -e "\e[1;36m=================================================\e[0m"
EOF
chmod +x /etc/profile.d/99-respon-server.sh

# =====================================================================
# 🛠️ RACIKAN SSHD TUNING: SPEK BADAK ANTI-EOF + ULTRA THROUGHPUT
# =====================================================================
echo "[*] Membuat Konfigurasi sshd_config Turbo..."
cat << 'EOF' > /etc/ssh/sshd_config
Port 22
ListenAddress 127.0.0.1
PermitRootLogin yes
PasswordAuthentication yes
PermitEmptyPasswords no
ChallengeResponseAuthentication no
UsePAM yes
PrintMotd no
Banner /etc/ssh_banner
AcceptEnv LANG LC_*
Subsystem sftp /usr/lib/ssh/sftp-server

# 🚀 RACIKAN ULTRA SAKTI ANTI-EOF (Siap dihajar spam koneksi tethering brutal)
MaxStartups 100:30:500
MaxSessions 100
MaxAuthTries 10

# 🔥 SUNTIKAN SAKTI ANTI-REKONEK
ClientAliveInterval 30
ClientAliveCountMax 99999
TCPKeepAlive yes
LoginGraceTime 30

# 🚀 CIPHERS OPTIMIZED FOR SPEED: Hanya menyisakan yang enteng di CPU agar bandwidth plong
Ciphers chacha20-poly1305@openssh.com,aes128-gcm@openssh.com,aes256-gcm@openssh.com
KexAlgorithms curve25519-sha256,curve25519-sha256@libssh.org
MACs umac-64-etm@openssh.com,umac-128-etm@openssh.com,hmac-sha2-256-etm@openssh.com
EOF
# =====================================================================

echo "[*] Mengonfigurasi User SSH..."
if ! id "$USER_NAME" &>/dev/null; then
    useradd -m -s /bin/bash "$USER_NAME"
fi
echo "$USER_NAME:$USER_PASS" | chpasswd

echo "[*] Memulai OpenSSH Server di Port Lokal 22..."
/usr/sbin/sshd

# 🔥 TAMBAHAN SSL: Buat Sertifikat SSL Stunnel
echo "[*] Membuat Sertifikat SSL Stunnel..."
openssl req -new -newkey rsa:2048 -days 365 -nodes -x509 \
    -subj "/C=ID/ST=Jakarta/L=Jakarta/O=RailwaySSH/CN=localhost" \
    -keyout /etc/stunnel/stunnel.pem -out /etc/stunnel/stunnel.pem

# 🚀 OPTIMASI STUNNEL: Matikan debug log & perkecil timeout agar RAM fokus ke speed data
echo "[*] Mengonfigurasi Stunnel internal di Port $SSL_INTERNAL_PORT..."
cat <<EOF > /etc/stunnel/stunnel.conf
pid = /var/run/stunnel.pid
foreground = yes
debug = 0

[ssh-ssl]
accept = 127.0.0.1:$SSL_INTERNAL_PORT
connect = 127.0.0.1:22
cert = /etc/stunnel/stunnel.pem
options = NO_SSLv2
options = NO_SSLv3
TIMEOUTclose = 0
EOF

echo "[*] Menambahkan alias dan auto-start menu ke .bashrc..."
cat <<'EOF'>> ~/.bashrc
clear
alias c='clear'
alias x='exit'
alias +x='chmod +x'
alias cls='clear;ls'

menu
EOF

cat <<'EOF'>> /etc/skel/.bashrc
clear
menu
EOF

echo "[*] Memulai Stunnel (internal, port $SSL_INTERNAL_PORT)..."
stunnel /etc/stunnel/stunnel.conf &

if [ -n "$CF_TUNNEL_TOKEN" ]; then
    echo "[*] Menjalankan Cloudflare Tunnel (Argo)..."
    cloudflared tunnel run --token "$CF_TUNNEL_TOKEN" &
else
    echo "[!] CF_TUNNEL_TOKEN kosong -> Cloudflare Tunnel dilewati."
fi

# 🎨 BANNER DITENGAH & WARNA-WARNI UNTUK TAMPILAN STARTUP LOG RAILWAY
cyan="\e[1;36m"
yellow="\e[1;33m"
magenta="\e[1;35m"
green="\e[1;32m"
reset="\e[0m"

rawTitle="⚡ GOLANG TUNNEL PRO: FIXED DPI DESTROYER v5.1 FULL SPEED ACTIVE ⚡"
rawOwner="👑 PRIVATE TUNNEL BY: DEDEFATHU 👑"

paddingTitle=$(( (66 - ${#rawTitle}) / 2 ))
paddingOwner=$(( (66 - ${#rawOwner}) / 2 ))

centerTitle=$(printf "%${paddingTitle}s" "")$rawTitle
centerOwner=$(printf "%${paddingOwner}s" "")$rawOwner

echo -e "${cyan}==================================================================${reset}"
echo -e "${yellow}${centerTitle}${reset}"
echo -e "${magenta}${centerOwner}${reset}"
echo -e "${green}==================================================================${reset}"
echo -e "${green}[*] Engine listening smoothly on port: ${PUBLIC_PORT}${reset}"
echo -e "${cyan}==================================================================${reset}"

exec env \
    PORT="$PUBLIC_PORT" \
    SSL_TARGET_HOST="127.0.0.1" \
    SSL_TARGET_PORT="$SSL_INTERNAL_PORT" \
    WS_TARGET_HOST="127.0.0.1" \
    WS_TARGET_PORT="22" \
    turbo-proxy
