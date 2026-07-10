package main

import (
	"bytes"
	"crypto/sha1"
	"encoding/base64"
	"fmt"
	"log"
	"net"
	"os"
	"strings"
	"sync"
	"time"
)

const (
	WS_MAGIC           = "258EAFA5-E914-47DA-95CA-C5AB0DC85B11"
	TLS_HANDSHAKE_BYTE = 0x16
)

func main() {
	listenPort := getEnv("PORT", "8080")
	sslTargetHost := getEnv("SSL_TARGET_HOST", "127.0.0.1")
	sslTargetPort := getEnv("SSL_TARGET_PORT", "2443")
	wsTargetHost := getEnv("WS_TARGET_HOST", "127.0.0.1")
	wsTargetPort := getEnv("WS_TARGET_PORT", "22")

	log.Println("================================================================")
	log.Printf("⚡ GOLANG HYBRID TUNNEL ACTIVE ON PORT %s ⚡\n", listenPort)
	log.Println("================================================================")

	listener, err := net.Listen("tcp", ":"+listenPort)
	if err != nil {
		log.Fatalf("[-] Listener gagal: %v", err)
	}
	defer listener.Close()

	for {
		conn, err := listener.Accept()
		if err != nil {
			continue
		}
		go func(c net.Conn) {
			if tcp, ok := c.(*net.TCPConn); ok {
				tcp.SetNoDelay(true)
				tcp.SetKeepAlive(true)
				tcp.SetKeepAlivePeriod(30 * time.Second)
			}
			defer c.Close()

			// 🕒 BUFFER RAKUS: Kasih waktu 3 detik buat nelan seluruh tumpukan payload kotor
			c.SetReadDeadline(time.Now().Add(3 * time.Second))
			buf := make([]byte, 65536)
			n, err := c.Read(buf)
			if err != nil || n == 0 {
				return
			}
			c.SetReadDeadline(time.Time{})
			rawPayload := buf[:n]

			// DETEKSI SSL MURNI
			if rawPayload[0] == TLS_HANDSHAKE_BYTE {
				target, err := net.DialTimeout("tcp", sslTargetHost+":"+sslTargetPort, 4*time.Second)
				if err != nil {
					return
				}
				defer targetConnClose(target)
				_, _ = target.Write(rawPayload)
				pipe(c, target, false)
				return
			}

			// DETEKSI WEBSOCKET / HTTP INJECTION
			reqStr := string(rawPayload)
			wsKey := ""
			for _, line := range strings.Split(reqStr, "\r\n") {
				if strings.Contains(strings.ToLower(line), "sec-websocket-key") {
					if parts := strings.Split(line, ":"); len(parts) > 1 {
						wsKey = strings.TrimSpace(parts[1])
						break
					}
				}
			}

			if wsKey == "" {
				wsKey = base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%d-salt", time.Now().UnixNano())))
			}

			h := sha1.New()
			h.Write([]byte(wsKey + WS_MAGIC))
			acceptKey := base64.StdEncoding.EncodeToString(h.Sum(nil))

			// Kirim balasan 101 murni setelah seluruh payload dump terbaca
			response := "HTTP/1.1 101 Switching Protocols\r\n" +
				"Upgrade: websocket\r\n" +
				"Connection: Upgrade\r\n" +
				"Sec-WebSocket-Accept: " + acceptKey + "\r\n\r\n"
			_, err = c.Write([]byte(response))
			if err != nil {
				return
			}

			// Hubungkan ke SSH Internal
			sshTarget, err := net.DialTimeout("tcp", wsTargetHost+":"+wsTargetPort, 4*time.Second)
			if err != nil {
				return
			}
			defer targetConnClose(sshTarget)

			// Cari apakah header SSH- sudah ikut terkirim di paket pertama
			idx := bytes.Index(rawPayload, []byte("SSH-"))
			if idx != -1 {
				_, _ = sshTarget.Write(rawPayload[idx:])
			}

			pipe(c, sshTarget, true)
		}(conn)
	}
}

func pipe(client, target net.Conn, isWS bool) {
	var once sync.Once
	closeAll := func() {
		client.Close()
		target.Close()
	}

	if tcp, ok := target.(*net.TCPConn); ok {
		tcp.SetNoDelay(true)
		tcp.SetKeepAlive(true)
		tcp.SetKeepAlivePeriod(30 * time.Second)
	}

	// Jalur HP -> SSH
	go func() {
		buf := make([]byte, 32768)
		first := true
		for {
			client.SetReadDeadline(time.Now().Add(90 * time.Second))
			n, err := client.Read(buf)
			if err != nil {
				break
			}
			data := buf[:n]
			if isWS && first {
				// Saring sisa sampah jika ada yang tertinggal pasca handshake
				idx := bytes.Index(data, []byte("SSH-"))
				if idx != -1 {
					data = data[idx:]
					first = false
				}
			}
			_, err = target.Write(data)
			if err != nil {
				break
			}
		}
		once.Do(closeAll)
	}()

	// Jalur SSH -> HP + Anti RTO Ping Engine
	buf := make([]byte, 32768)
	for {
		target.SetReadDeadline(time.Now().Add(25 * time.Second))
		n, err := target.Read(buf)
		if err != nil {
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				if isWS {
					// Kirim ping websocket ke aplikasi DarkTunnel biar gak bengong
					_, err = client.Write([]byte{0x89, 0x00})
					if err != nil {
						break
					}
					continue
				}
			}
			break
		}
		_, err = client.Write(buf[:n])
		if err != nil {
			break
		}
	}
	once.Do(closeAll)
}

func targetConnClose(c net.Conn) {
	if c != nil {
		c.Close()
	}
}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}
