package main

import (
	"bytes"
	"crypto/rand"
	"crypto/sha1"
	"encoding/base64"
	"fmt"
	"log"
	"math/big"
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

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}

func secureRandom(max int64) int64 {
	nBig, err := rand.Int(rand.Reader, big.NewInt(max))
	if err != nil {
		return 0
	}
	return nBig.Int64()
}

// 🚀 ENGINE TUNING: Optimasi Buffer Socket untuk Kcepatan Penuh
func turboTune(c net.Conn) {
	if tcp, ok := c.(*net.TCPConn); ok {
		_ = tcp.SetNoDelay(true)
		_ = tcp.SetKeepAlive(true)
		_ = tcp.SetKeepAlivePeriod(10 * time.Second)
		
		// Menaikkan ukuran buffer TCP Window OS secara agresif (Kunci Full Speed Download)
		_ = tcp.SetReadBuffer(262144)  // 256 KB Buffer
		_ = tcp.SetWriteBuffer(262144) // 256 KB Buffer
	}
}

func main() {
	listenPort := getEnv("PORT", "8080")
	sslTargetHost := getEnv("SSL_TARGET_HOST", "127.0.0.1")
	sslTargetPort := getEnv("SSL_TARGET_PORT", "2443")
	wsTargetHost := getEnv("WS_TARGET_HOST", "127.0.0.1")
	wsTargetPort := getEnv("WS_TARGET_PORT", "22")

	log.Println("==================================================================")
	log.Println("🚀 GOLANG TUNNEL PRO: FIXED DPI DESTROYER v5.1 FULL SPEED ACTIVE 🔥")
	log.Println("==================================================================")

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
		go handleFixedEnterprise(conn, sslTargetHost, sslTargetPort, wsTargetHost, wsTargetPort)
	}
}

func handleFixedEnterprise(c net.Conn, sslHost, sslPort, wsHost, wsPort string) {
	turboTune(c) // Optimasi koneksi client awal
	defer c.Close()

	buf := make([]byte, 131072)
	c.SetReadDeadline(time.Now().Add(4 * time.Second))
	n, err := c.Read(buf)
	if err != nil || n == 0 {
		return
	}
	c.SetReadDeadline(time.Time{})
	rawPayload := buf[:n]

	// SSL Detection
	if rawPayload[0] == TLS_HANDSHAKE_BYTE {
		target, err := net.DialTimeout("tcp", sslHost+":"+sslPort, 4*time.Second)
		if err != nil {
			return
		}
		turboTune(target)
		defer target.Close()
		_, _ = target.Write(rawPayload)
		pipeFixed(c, target, false)
		return
	}

	// Enhanced Payload Handshake
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
		wsKey = base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%d", time.Now().UnixNano())))
	}

	h := sha1.New()
	h.Write([]byte(wsKey + WS_MAGIC))
	acceptKey := base64.StdEncoding.EncodeToString(h.Sum(nil))

	response := "HTTP/1.1 101 Switching Protocols\r\n" +
		"Upgrade: websocket\r\n" +
		"Connection: Upgrade\r\n" +
		"Sec-WebSocket-Accept: " + acceptKey + "\r\n\r\n"
	_, err = c.Write([]byte(response))
	if err != nil {
		return
	}

	// Konek ke SSH
	sshTarget, err := net.DialTimeout("tcp", wsHost+":"+wsPort, 4*time.Second)
	if err != nil {
		return
	}
	turboTune(sshTarget)
	defer sshTarget.Close()

	if idx := bytes.Index(rawPayload, []byte("SSH-")); idx != -1 {
		_, _ = sshTarget.Write(rawPayload[idx:])
	}

	pipeFixed(c, sshTarget, true)
}

func pipeFixed(client, target net.Conn, isWS bool) {
	var once sync.Once
	closeAll := func() {
		_ = client.Close()
		_ = target.Close()
	}

	var wg sync.WaitGroup
	wg.Add(2)

	// Jalur A: HP -> SSH Server (Dengan Anti-DPI Jittering Super Halus)
	go func() {
		defer wg.Done()
		buf := make([]byte, 65536)
		first := true
		for {
			client.SetReadDeadline(time.Now().Add(120 * time.Second))
			n, err := client.Read(buf)
			if err != nil {
				break
			}
			
			data := buf[:n]
			if isWS && first {
				if idx := bytes.Index(data, []byte("SSH-")); idx != -1 {
					data = data[idx:]
					first = false
				}
			}

			// 🔥 FULL SPEED ADJUSTMENT: Jitter diperkecil ke 1-4ms agar upload secepat kilat tapi DPI tetap buta
			jitter := secureRandom(4) + 1 
			time.Sleep(time.Duration(jitter) * time.Millisecond)

			_, err = target.Write(data)
			if err != nil {
				break
			}
		}
		once.Do(closeAll)
	}()

	// Jalur B: SSH Server -> HP (Anti-RTO & Full Speed Download Tanpa Batas)
	go func() {
		defer wg.Done()
		buf := make([]byte, 65536)
		for {
			target.SetReadDeadline(time.Now().Add(20 * time.Second))
			n, err := target.Read(buf)
			
			if err != nil {
				if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
					if isWS {
						_, err = client.Write([]byte{0x89, 0x00})
						if err != nil {
							break
						}
						continue
					}
				}
				break
			}
			
			if n > 0 {
				_, err = client.Write(buf[:n])
				if err != nil {
					break
				}
			}
		}
		once.Do(closeAll)
	}()

	wg.Wait()
}
