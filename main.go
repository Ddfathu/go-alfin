package main

import (
	"bytes"
	"crypto/sha1"
	"encoding/base64"
	"fmt"
	"log"
	"net"
	"os"
	"runtime"
	"runtime/debug"
	"strings"
	"sync"
	"time"
)

const (
	WS_MAGIC           = "258EAFA5-E914-47DA-95CA-C5AB0DC85B11"
	TLS_HANDSHAKE_BYTE = 0x16
)

var bufferPool = sync.Pool{
	New: func() interface{} {
		return make([]byte, 65536)
	},
}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}

func turboTune(c net.Conn) {
	if tcp, ok := c.(*net.TCPConn); ok {
		_ = tcp.SetNoDelay(true)
		_ = tcp.SetKeepAlive(true)
		_ = tcp.SetKeepAlivePeriod(30 * time.Second)
		_ = tcp.SetReadBuffer(524288)  // Diperbesar ke 512KB untuk kestabilan speedtest
		_ = tcp.SetWriteBuffer(524288) 
	}
}

func main() {
	debug.SetGCPercent(-1)
	go func() {
		for {
			time.Sleep(10 * time.Second)
			runtime.GC() 
		}
	}()

	listenPort := getEnv("PORT", "8080")
	sslTargetHost := getEnv("SSL_TARGET_HOST", "127.0.0.1")
	sslTargetPort := getEnv("SSL_TARGET_PORT", "2443")
	wsTargetHost := getEnv("WS_TARGET_HOST", "127.0.0.1")
	wsTargetPort := getEnv("WS_TARGET_PORT", "22")

	fmt.Printf("[monster-mux-go] PERFECT HYBRID v8.5 ACTIVE on Port: %s 🚀\n", listenPort)

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
		go handlePureTurbo(conn, sslTargetHost, sslTargetPort, wsTargetHost, wsTargetPort)
	}
}

func handlePureTurbo(c net.Conn, sslHost, sslPort, wsHost, wsPort string) {
	turboTune(c) 
	defer c.Close()

	buf := make([]byte, 65536)
	n, err := c.Read(buf)
	if err != nil || n == 0 {
		return
	}
	rawPayload := buf[:n]

	// Jalur SSL Bypass
	if rawPayload[0] == TLS_HANDSHAKE_BYTE {
		target, err := net.DialTimeout("tcp", sslHost+":"+sslPort, 4*time.Second)
		if err != nil {
			return
		}
		turboTune(target)
		defer target.Close()
		_, _ = target.Write(rawPayload)
		pipePure(c, target, false)
		return
	}

	// Jalur WebSocket (Enhanced Payload Handler)
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

	// Hubungkan ke SSH Internal (Dropbear)
	sshTarget, err := net.DialTimeout("tcp", wsHost+":"+wsPort, 4*time.Second)
	if err != nil {
		return
	}
	turboTune(sshTarget)
	defer sshTarget.Close()

	// 🔥 LOGIKA PEMBERSIH UTUH: Kita bersihkan ampas teks HTTP tanpa memotong isi binary data jabat tangan
	cleanPayload := rawPayload
	if strings.Contains(reqStr, "PATCH") || strings.Contains(reqStr, "HTTP/") || strings.Contains(reqStr, "BMOVE") || strings.Contains(reqStr, "GET ") {
		// Hapus trigger teks enhanced method proxy tiruan agar Dropbear tidak bingung
		cleanPayload = bytes.ReplaceAll(cleanPayload, []byte("BMOVE / HTTP/1.0\r\n"), []byte(""))
		cleanPayload = bytes.ReplaceAll(cleanPayload, []byte("PATCH / HTTP/1.1\r\n"), []byte(""))
		// Hapus sisa baris Host jika masih menempel akibat payload Custom
		if idx := bytes.Index(cleanPayload, []byte("Host:")); idx != -1 {
			if endIdx := bytes.Index(cleanPayload[idx:], []byte("\r\n")); endIdx != -1 {
				cleanPayload = bytes.NewBuffer(append(cleanPayload[:idx], cleanPayload[idx+endIdx+2:]...)).Bytes()
			}
		}
	}

	if len(cleanPayload) > 0 {
		_, _ = sshTarget.Write(cleanPayload)
	}

	pipePure(c, sshTarget, true)
}

func pipePure(client, target net.Conn, isWS bool) {
	var once sync.Once
	closeAll := func() {
		_ = client.Close()
		_ = target.Close()
	}

	var wg sync.WaitGroup
	wg.Add(2)

	// Jalur A: HP -> SSH Server (Upload Speedtest Kebal Total)
	go func() {
		defer wg.Done()
		buf := bufferPool.Get().([]byte)
		defer bufferPool.Put(buf) 
		
		for {
			_ = client.SetReadDeadline(time.Now().Add(120 * time.Second))
			n, err := client.Read(buf)
			if err != nil {
				break
			}
			
			// 🔥 PENTING: Paket susulan saat speedtest upload langsung dioper mentah 100% 
			// Tanpa saringan string/continue yang merusak urutan paket cipher
			_, err = target.Write(buf[:n])
			if err != nil {
				break
			}
		}
		once.Do(closeAll)
	}()

	// Jalur B: SSH Server -> HP (Download Speed Los Maksimal)
	go func() {
		defer wg.Done()
		buf := bufferPool.Get().([]byte)
		defer bufferPool.Put(buf)
		
		for {
			_ = target.SetReadDeadline(time.Now().Add(60 * time.Second))
			n, err := target.Read(buf)
			if err != nil {
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
