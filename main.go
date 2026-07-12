package main

import (
	"bytes"
	"crypto/sha1"
	"encoding/base64"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

var (
	LISTEN_PORT      = getEnvInt("PORT", 8080)
	SSL_TARGET_HOST  = getEnvStr("SSL_TARGET_HOST", "127.0.0.1")
	SSL_TARGET_PORT  = getEnvInt("SSL_TARGET_PORT", 2443)
	SSH_TARGET_PORT  = getEnvInt("WS_TARGET_PORT", 22)
)

const (
	WS_MAGIC           = "258EAFA5-E914-47DA-95CA-C5AB0DC85B11"
	DEFAULT_RESPONSE   = "HTTP/1.1 101 Switching Protocols\r\n\r\n"
	TLS_HANDSHAKE_BYTE = 0x16
	BUFFER_SIZE        = 1024 * 1024 
)

func main() {
	listenAddr := fmt.Sprintf("0.0.0.0:%d", LISTEN_PORT)
	listener, err := net.Listen("tcp", listenAddr)
	if err != nil {
		log.Fatalf("Gagal menjalankan server: %v", err)
	}
	defer listener.Close()

	fmt.Printf("[monster-mux-go] ALL-IN-ONE FIXED ELITE v7.7 ACTIVE on Port: %d 🚀\n", LISTEN_PORT)

	for {
		clientConn, err := listener.Accept()
		if err != nil {
			continue
		}
		go handleConnection(clientConn)
	}
}

func handleConnection(clientConn net.Conn) {
	if tcpConn, ok := clientConn.(*net.TCPConn); ok {
		tcpConn.SetNoDelay(true)
		tcpConn.SetKeepAlive(true)
		tcpConn.SetKeepAlivePeriod(30 * time.Second) // 🔥 Sudah diperbaiki di sini
	}

	var targetConn net.Conn
	var err error
	var mu sync.Mutex

	isWsJalur := false
	firstPacketRead := false
	packetCounter := 0
	backendReady := false
	var queueBuffers [][]byte

	closeAll := func() {
		mu.Lock()
		defer mu.Unlock()
		if clientConn != nil {
			clientConn.Close()
		}
		if targetConn != nil {
			targetConn.Close()
		}
	}

	buffer := make([]byte, BUFFER_SIZE)

	for {
		n, errRead := clientConn.Read(buffer)
		if errRead != nil {
			closeAll()
			return
		}
		if n == 0 {
			continue
		}

		chunk := make([]byte, n)
		copy(chunk, buffer[:n])
		packetCounter++

		if !firstPacketRead {
			firstPacketRead = true

			if chunk[0] == TLS_HANDSHAKE_BYTE {
				isWsJalur = false
				targetAddr := fmt.Sprintf("%s:%d", SSL_TARGET_HOST, SSL_TARGET_PORT)
				targetConn, err = net.DialTimeout("tcp", targetAddr, 5*time.Second)
				if err != nil {
					closeAll()
					return
				}
				if tcpConn, ok := targetConn.(*net.TCPConn); ok {
					tcpConn.SetNoDelay(true)
				}

				_, err = targetConn.Write(chunk)
				if err != nil {
					closeAll()
					return
				}
				backendReady = true

			} else {
				isWsJalur = true
				headers := parseHeaders(chunk)
				rawTextLower := strings.ToLower(string(chunk))
				isWsUpgrade := strings.Contains(rawTextLower, "upgrade: websocket") || headers["upgrade"] == "websocket"

				if isWsUpgrade {
					wsKey := headers["sec-websocket-key"]
					if wsKey == "" && strings.Contains(rawTextLower, "sec-websocket-key:") {
						lines := strings.Split(string(chunk), "\r\n")
						for _, line := range lines {
							if strings.Contains(strings.ToLower(line), "sec-websocket-key") {
								parts := strings.Split(line, ":")
								if len(parts) > 1 {
									wsKey = strings.TrimSpace(parts[1])
									break
								}
							}
						}
					}
					if wsKey == "" {
						wsKey = base64.StdEncoding.EncodeToString([]byte("monster-mux-key-random"))
					}

					h := sha1.New()
					h.Write([]byte(wsKey + WS_MAGIC))
					acceptKey := base64.StdEncoding.EncodeToString(h.Sum(nil))

					response := fmt.Sprintf("HTTP/1.1 101 Switching Protocols\r\nUpgrade: websocket\r\nConnection: Upgrade\r\nSec-WebSocket-Accept: %s\r\n\r\n", acceptKey)
					clientConn.Write([]byte(response))
				} else {
					clientConn.Write([]byte(DEFAULT_RESPONSE))
				}

				sshAddr := fmt.Sprintf("127.0.0.1:%d", SSH_TARGET_PORT)
				targetConn, err = net.DialTimeout("tcp", sshAddr, 5*time.Second)
				if err != nil {
					closeAll()
					return
				}
				if tcpConn, ok := targetConn.(*net.TCPConn); ok {
					tcpConn.SetNoDelay(true)
				}

				backendReady = true

				if len(queueBuffers) > 0 {
					for _, qChunk := range queueBuffers {
						targetConn.Write(qChunk)
					}
					queueBuffers = nil
				}
			}

			// Jalur balik murni dari Dropbear ke HP client
			go func() {
				defer closeAll()
				resBuffer := make([]byte, BUFFER_SIZE)
				for {
					nRes, errRes := targetConn.Read(resBuffer)
					if errRes != nil {
						return
					}
					if nRes > 0 {
						_, errWrite := clientConn.Write(resBuffer[:nRes])
						if errWrite != nil {
							return
						}
					}
				}
			}()

			continue
		}

		if isWsJalur {
			cleanChunk := chunk

			if packetCounter <= 3 {
				chunkStr := string(chunk)
				if strings.Contains(chunkStr, "PATCH") || strings.Contains(chunkStr, "HTTP/") || strings.Contains(chunkStr, "BMOVE") || strings.Contains(chunkStr, "GET ") {
					if strings.Contains(chunkStr, "SSH-") {
						idx := strings.Index(chunkStr, "SSH-")
						cleanChunk = chunk[idx:]
					} else if bytes.Contains(chunk, []byte{0x53, 0x53, 0x48}) {
						idx := bytes.Index(chunk, []byte{0x53, 0x53, 0x48})
						cleanChunk = chunk[idx:]
					} else {
						continue 
					}
				}
			}

			if !backendReady {
				queueBuffers = append(queueBuffers, cleanChunk)
			} else {
				_, err = targetConn.Write(cleanChunk)
				if err != nil {
					closeAll()
					return
				}
			}
		} else {
			if !backendReady {
				queueBuffers = append(queueBuffers, chunk)
			} else {
				_, err = targetConn.Write(chunk)
				if err != nil {
					closeAll()
					return
				}
			}
		}
	}
}

func parseHeaders(rawBuffer []byte) map[string]string {
	headers := make(map[string]string)
	lines := strings.Split(string(rawBuffer), "\r\n")
	if len(lines) <= 1 {
		return headers
	}
	for i := 1; i < len(lines); i++ {
		line := lines[i]
		if strings.Contains(line, ":") {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				key := strings.TrimSpace(strings.ToLower(parts[0]))
				val := strings.TrimSpace(parts[1])
				headers[key] = val
			}
		}
	}
	return headers
}

func getEnvStr(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	if value, ok := os.LookupEnv(key); ok {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return fallback
}
