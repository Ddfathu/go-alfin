package main

import (
	"bytes"
	"context"
	"crypto/sha1"
	"encoding/base64"
	"errors"
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
	MAX_JUNK_PARSE     = 256 * 1024 // Batas toleransi payload sampah (256 KB)
)

type TunnelEngine struct {
	ListenPort    string
	SSLTargetHost string
	SSLTargetPort string
	WSTargetHost  string
	WSTargetPort  string
}

func main() {
	engine := &TunnelEngine{
		ListenPort:    getEnv("PORT", "8080"),
		SSLTargetHost: getEnv("SSL_TARGET_HOST", "127.0.0.1"),
		SSLTargetPort: getEnv("SSL_TARGET_PORT", "2443"),
		WSTargetHost:  getEnv("WS_TARGET_HOST", "127.0.0.1"),
		WSTargetPort:  getEnv("WS_TARGET_PORT", "22"),
	}

	log.Println("================================================================")
	log.Println("⚡ GOLANG HYPER-TUNNEL ENGINE v3.0 (SUPER COMPLEX STATE) ACTIVE ⚡")
	log.Println("================================================================")

	listener, err := net.Listen("tcp", ":"+engine.ListenPort)
	if err != nil {
		log.Fatalf("[-] Fatal: Listener gagal: %v", err)
	}
	defer listener.Close()

	for {
		conn, err := listener.Accept()
		if err != nil {
			continue
		}
		go engine.orchestrate(conn)
	}
}

func (te *TunnelEngine) orchestrate(client net.Conn) {
	te.tuneSocket(client)
	defer client.Close()

	// Fase 1: Telan & Analisis Enhanced Payload
	client.SetReadDeadline(time.Now().Add(4 * time.Second))
	peekBuffer := make([]byte, 32768) // 32KB buffer awal untuk dump payload kotor
	n, err := client.Read(peekBuffer)
	if err != nil || n == 0 {
		return
	}
	client.SetReadDeadline(time.Time{})
	initialData := peekBuffer[:n]

	// 🛡️ DETEKSI JALUR SNI / SSL MURNI
	if initialData[0] == TLS_HANDSHAKE_BYTE {
		te.handleSSL(client, initialData)
		return
	}

	// 🌐 DETEKSI JALUR HTTP/WEBSOCKET (Enhanced Payload Mode)
	te.handleWebSocket(client, initialData)
}

func (te *TunnelEngine) handleSSL(client net.Conn, initialData []byte) {
	target, err := net.DialTimeout("tcp", te.SSLTargetHost+":"+te.SSLTargetPort, 4*time.Second)
	if err != nil {
		return
	}
	te.tuneSocket(target)
	_, _ = target.Write(initialData)

	te.biDirectionalPipe(client, target, false)
}

func (te *TunnelEngine) handleWebSocket(client net.Conn, initialData []byte) {
	reqStr := string(initialData)
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
		// Generate key palsu jika operator membuang header asli demi kestabilan handshake
		wsKey = base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%d-hyper-salt", time.Now().UnixNano())))
	}

	// Hitung Sec-WebSocket-Accept secara presisi
	h := sha1.New()
	h.Write([]byte(wsKey + WS_MAGIC))
	acceptKey := base64.StdEncoding.EncodeToString(h.Sum(nil))

	// Kirim Balasan Handshake Murni untuk mengunci koneksi di HP
	response := "HTTP/1.1 101 Switching Protocols\r\n" +
		"Upgrade: websocket\r\n" +
		"Connection: Upgrade\r\n" +
		"Sec-WebSocket-Accept: " + acceptKey + "\r\n\r\n"
	_, err := client.Write([]byte(response))
	if err != nil {
		return
	}

	// Konek ke SSH Server (Dropbear/OpenSSH)
	sshTarget := fmt.Sprintf("%s:%s", te.WSTargetHost, te.WSTargetPort)
	target, err := net.DialTimeout("tcp", sshTarget, 4*time.Second)
	if err != nil {
		return
	}
	te.tuneSocket(target)

	// Filter & Bersihkan Enhanced Payload sebelum dikirim ke SSH
	cleanedData, hasSSH := te.extractSSHHeader(initialData)
	if hasSSH && len(cleanedData) > 0 {
		_, _ = target.Write(cleanedData)
	} else if !hasSSH {
		// Jika SSH header tidak ditemukan di paket pertama karena tertimbun sampah payload, 
		// kita lakukan streaming filter tingkat lanjut (Advanced Stream Filtering)
		if err := te.streamFilterToTarget(client, target); err != nil {
			target.Close()
			return
		}
	}

	// Jalankan pipa data dua arah anti-RTO
	te.biDirectionalPipe(client, target, true)
}

// 🛡️ ADVANCED STATE MACHINE: Pengekstrak Header SSH dari Sampah Payload
func (te *TunnelEngine) extractSSHHeader(data []byte) ([]byte, bool) {
	idx := bytes.Index(data, []byte("SSH-"))
	if idx != -1 {
		return data[idx:], true
	}
	return nil, false
}

// ⏳ STREAM FILTERING: Jika payload kotornya terlalu besar & memotong paket data awal
func (te *TunnelEngine) streamFilterToTarget(client, target net.Conn) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	buf := make([]byte, 4096)
	var accumulated []byte

	for {
		select {
		case <-ctx.Done():
			return errors.New("timeout mencari header SSH di dalam payload")
		default:
			client.SetReadDeadline(time.Now().Add(2 * time.Second))
			n, err := client.Read(buf)
			if err != nil {
				return err
			}
			accumulated = append(accumulated, buf[:n]...)
			
			if idx := bytes.Index(accumulated, []byte("SSH-")); idx != -1 {
				client.SetReadDeadline(time.Time{})
				_, err = target.Write(accumulated[idx:])
				return err
			}

			if len(accumulated) > MAX_JUNK_PARSE {
				return errors.New("payload terlalu kotor melewati batas toleransi")
			}
		}
	}
}

// 🔄 BI-DIRECTIONAL PIPE & COMPLEX HEARTBEAT ENGINE
func (te *TunnelEngine) biDirectionalPipe(client, target net.Conn, isWS bool) {
	var wg sync.WaitGroup
	ctx, cancel := context.WithCancel(context.Background())
	
	wg.Add(2)

	// JALUR A: Client (HP) -> Target (SSH Server)
	go func() {
		defer wg.Done()
		defer cancel()
		buf := make([]byte, 32768) // Ukuran buffer optimal untuk menekan jitter
		for {
			select {
			case <-ctx.Done():
				return
			default:
				client.SetReadDeadline(time.Now().Add(90 * time.Second))
				n, err := client.Read(buf)
				if err != nil {
					return
				}
				if n > 0 {
					_, err = target.Write(buf[:n])
					if err != nil {
						return
					}
				}
			}
		}
	}()

	// JALUR B: Target (SSH Server) -> Client (HP) + WebSocket Heartbeat Engine
	go func() {
		defer wg.Done()
		defer cancel()
		buf := make([]byte, 32768)
		for {
			select {
			case <-ctx.Done():
				return
			default:
				// Ping dikirim setiap 20 detik jika tidak ada aktivitas data dari SSH
				target.SetReadDeadline(time.Now().Add(20 * time.Second))
				n, err := target.Read(buf)
				
				if err != nil {
					if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
						if isWS {
							// ⚡ INJEKSI FRAME WEBSOCKET PING (Murni Kebal RTO)
							// Mengirimkan ping frame 0x89 ke HP untuk menjaga kestabilan aplikasi VPN
							client.SetWriteDeadline(time.Now().Add(5 * time.Second))
							_, err = client.Write([]byte{0x89, 0x00})
							client.SetWriteDeadline(time.Time{})
							if err != nil {
								return
							}
							continue
						}
					}
					return
				}
				
				if n > 0 {
					_, err = client.Write(buf[:n])
					if err != nil {
						return
					}
				}
			}
		}
	}()

	wg.Wait()
}

// ⚙️ TWEAK KERNEL SOCKET (Optimasi Latensi & Ketahanan Sinyal)
func (te *TunnelEngine) tuneSocket(c net.Conn) {
	if tcp, ok := c.(*net.TCPConn); ok {
		_ = tcp.SetNoDelay(true)
		_ = tcp.SetKeepAlive(true)
		_ = tcp.SetKeepAlivePeriod(20 * time.Second)
		
		// Set Buffer Sistem Operasi secara agresif untuk kelancaran streaming/gaming
		_ = tcp.SetReadBuffer(65536)
		_ = tcp.SetWriteBuffer(65536)
	}
}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}
