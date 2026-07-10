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

func main() {
	listenPort := getEnv("PORT", "8080")
	sslTargetHost := getEnv("SSL_TARGET_HOST", "127.0.0.1")
	sslTargetPort := getEnv("SSL_TARGET_PORT", "2443")
	wsTargetHost := getEnv("WS_TARGET_HOST", "127.0.0.1")
	wsTargetPort := getEnv("WS_TARGET_PORT", "22")

	log.Println("==================================================================")
	log.Println("🔥 GOLANG ENTERPRISE TUNNEL ACTIVE: DPI DESTROYER ENGINE v5.0 🔥")
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
		go handleEnterprise(conn, sslTargetHost, sslTargetPort, wsTargetHost, wsTargetPort)
	}
}

// Fungsi helper untuk generate angka acak yang aman secara kriptografi
func secureRandom(max int64) int64 {
	nBig, err := rand.Int(rand.Reader, big.NewInt(max))
	if err != nil {
		return 0
	}
	return nBig.Int64()
}

func handleEnterprise(c net.Conn, sslHost, sslPort, wsHost, wsPort string) {
	if tcp, ok := c.(*net.TCPConn); ok {
		_ = tcp.SetNoDelay(true)
		_ = tcp.SetKeepAlive(true)
		_ = tcp.SetKeepAlivePeriod(10 * time.Second) // Sangat agresif agar tidak mati di BTS
	}
	defer c.Close()

	// Dialokasikan buffer besar per user demi performa manipulasi byte yang leluasa
	buf := make([]byte, 131072) // 128 KB Buffer raksasa

	c.SetReadDeadline(time.Now().Add(4 * time.Second))
	n, err := c.Read(buf)
	if err != nil || n == 0 {
		return
	}
	c.SetReadDeadline(time.Time{})
	rawPayload := buf[:n]

	// 🛡️ DETEKSI JALUR SSL / SNI
	if rawPayload[0] == TLS_HANDSHAKE_BYTE {
		target, err := net.DialTimeout("tcp", sslHost+":"+sslPort, 4*time.Second)
		if err != nil {
			return
		}
		defer target.Close()
		_, _ = target.Write(rawPayload)
		pipeEnterprise(c, target, false)
		return
	}

	// 🌐 MODE ENHANCED PAYLOAD OBFUSCATION (WEBSOCKET)
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

	// Konek ke SSH Server Internal
	sshTarget, err := net.DialTimeout("tcp", wsHost+":"+wsPort, 4*time.Second)
	if err != nil {
		return
	}
	defer sshTarget.Close()

	// Ekstrak data SSH asli dari timbunan payload kotor
	if idx := bytes.Index(rawPayload, []byte("SSH-")); idx != -1 {
		_, _ = sshTarget.Write(rawPayload[idx:])
	}

	pipeEnterprise(c, sshTarget, true)
}

func pipeEnterprise(client, target net.Conn, isWS bool) {
	var once sync.Once
	closeAll := func() {
		_ = client.Close()
		_ = target.Close()
	}

	if tcp, ok := target.(*net.TCPConn); ok {
		_ = tcp.SetNoDelay(true)
		_ = tcp.SetKeepAlive(true)
		_ = tcp.SetKeepAlivePeriod(10 * time.Second)
	}

	// 🟢 JALUR A: HP (Client) -> Server SSH (Dengan Anti-DPI Timing & Scrambling)
	go func() {
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

			// 🧠 ANTI-DPI LOGIC 1: Micro-Jitter & Time Delay Simulation
			// Jeda acak 2-8 milidetik disuntikkan untuk mengacaukan AI Operator yang membaca pola ketukan paket
			jitter := secureRandom(7) + 2
			time.Sleep(time.Duration(jitter) * time.Millisecond)

			_, err = target.Write(data)
			if err != nil {
				break
			}
		}
		once.Do(closeAll)
	}()

	// 🔴 JALUR B: Server SSH -> HP (Dengan Dynamic Padding & Fake Traffic/Chaffing)
	buf := make([]byte, 65536)
	for {
		// Set waktu tunggu 15 detik. Jika lewat, kita suntik fake traffic.
		target.SetReadDeadline(time.Now().Add(15 * time.Second))
		n, err := target.Read(buf)
		
		if err != nil {
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				if isWS {
					// 🧠 ANTI-DPI LOGIC 2: Chaffing Engine (Fake Traffic Injection)
					// Selain kirim opcode ping murni (0x89), kita kirim payload 'kosong' dengan ukuran acak (10-60 byte)
					// Langkah ini membuat DPI mengira ada traffic web interaktif yang sedang mengalir
					fakeSize := secureRandom(50) + 10
					fakePayload := make([]byte, fakeSize)
					_, _ = rand.Read(fakePayload) // Isi dengan byte acak tingkat entropi tinggi

					// Bungkus dalam format frame teks WebSocket palsu (Opcode 0x01 atau Ping 0x89)
					_, err = client.Write([]byte{0x89, byte(fakeSize)})
					if err == nil {
						_, err = client.Write(fakePayload)
					}
					
					if err != nil {
						break
					}
					continue
				}
			}
			break
		}
		
		if n > 0 {
			dataToSend := buf[:n]

			// 🧠 ANTI-DPI LOGIC 3: Dynamic Packet Padding (Morphing Size)
			// Jika mendeteksi paket berukuran kecil (rentan dianalisis DPI), kita tambah padding acak di akhir frame
			if isWS && n < 200 {
				paddingSize := secureRandom(64) + 16 // Tambah 16-80 byte acak
				padding := make([]byte, paddingSize)
				_, _ = rand.Read(padding)
				
				// Kirim data asli
				_, err = client.Write(dataToSend)
				if err != nil {
					break
				}
				// Kirim data padding sebagai frame berkelanjutan (Continuation Frame - Opcode 0x00)
				_, err = client.Write([]byte{0x00, byte(paddingSize)})
				if err == nil {
					_, err = client.Write(padding)
				}
				if err != nil {
					break
				}
			} else {
				// Jalur normal untuk paket besar (seperti streaming/download)
				_, err = client.Write(dataToSend)
				if err != nil {
					break
				}
			}
		}
	}
	once.Do(closeAll)
}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}
