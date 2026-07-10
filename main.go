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

// 🧠 AI SMART RECOVERY DEFINITION
type ClientTracker struct {
	FailCount  int
	LastActive time.Time
}

var (
	trackerMutex sync.Mutex
	clientMap    = make(map[string]*ClientTracker)
)

// Fungsi untuk mencatat dan memeriksa apakah aplikasi HP sedang stuck loop
func checkSmartRecovery(remoteAddr string, isFail bool) bool {
	trackerMutex.Lock()
	defer trackerMutex.Unlock()

	// Ambil IP Core-nya saja (tanpa port)
	ip, _, _ := net.SplitHostPort(remoteAddr)
	
	tracker, exists := clientMap[ip]
	if !exists {
		tracker = &ClientTracker{FailCount: 0, LastActive: time.Now()}
		clientMap[ip] = tracker
	}

	// Jika sukses konek sampai SSH, reset counter ke nol
	if !isFail {
		tracker.FailCount = 0
		tracker.LastActive = time.Now()
		return false
	}

	// Jika terdeteksi gagal (hanya kirim sampah enhanced)
	tracker.FailCount++
	tracker.LastActive = time.Now()

	// 🚨 DETEKSI CERDAS: Jika sudah 5 kali berturut-turut terdeteksi stuck/timeout
	if tracker.FailCount >= 5 {
		log.Printf("\033[31m[⚠️ AI DETECT] Aplikasi HP terdeteksi STUCK/TIMEOUT (%d x)! Mengaktifkan Force Auto-Fresh...\033[0m\n", tracker.FailCount)
		tracker.FailCount = 0 // Reset counter setelah tindakan diambil
		return true           // Trigger untuk putus instan/force refresh
	}

	return false
}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}

// 🚀 ENGINE TUNING MAX: Pipa buffer diperlebar maksimal untuk kecepatan loss
func turboTune(c net.Conn) {
	if tcp, ok := c.(*net.TCPConn); ok {
		_ = tcp.SetNoDelay(true)
		_ = tcp.SetKeepAlive(true)
		_ = tcp.SetKeepAlivePeriod(10 * time.Second)
		
		// Buffer 256KB agar tidak ada antrean paket di level OS Railway
		_ = tcp.SetReadBuffer(262144)  
		_ = tcp.SetWriteBuffer(262144) 
	}
}

func main() {
	listenPort := getEnv("PORT", "8080")
	sslTargetHost := getEnv("SSL_TARGET_HOST", "127.0.0.1")
	sslTargetPort := getEnv("SSL_TARGET_PORT", "2443")
	wsTargetHost := getEnv("WS_TARGET_HOST", "127.0.0.1")
	wsTargetPort := getEnv("WS_TARGET_PORT", "22")

	// 🎨 ANSI COLOR & CENTER BANNER
	reset := "\033[0m"
	cyan := "\033[36m"
	yellow := "\033[33m"
	magenta := "\033[35m"
	green := "\033[32m"

	rawTitle := "⚡ GOLANG TUNNEL PRO: FIXED ANTI-RTO v5.9 AI SMART TURBO ⚡"
	rawOwner := "👑 PRIVATE TUNNEL BY: DEDEFATHU 👑"
	
	paddingTitle := (66 - len(rawTitle)) / 2
	paddingOwner := (66 - len(rawOwner)) / 2
	
	centerTitle := strings.Repeat(" ", paddingTitle) + rawTitle
	centerOwner := strings.Repeat(" ", paddingOwner) + rawOwner

	log.Println(cyan + "==================================================================" + reset)
	log.Println(yellow + centerTitle + reset)
	log.Println(magenta + centerOwner + reset)
	log.Println(green + "==================================================================" + reset)
	log.Printf(green+"[*] Engine listening smoothly on port: %s\n"+reset, listenPort)
	log.Println(cyan + "==================================================================" + reset)

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

	buf := make([]byte, 131072)
	c.SetReadDeadline(time.Now().Add(4 * time.Second))
	n, err := c.Read(buf)
	if err != nil || n == 0 {
		return
	}
	c.SetReadDeadline(time.Time{})
	rawPayload := buf[:n]

	// Jalur SSL
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

	// Jalur WebSocket (Enhanced Payload)
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

	// Hubungkan ke SSH Internal
	sshTarget, err := net.DialTimeout("tcp", wsHost+":"+wsPort, 4*time.Second)
	if err != nil {
		return
	}
	turboTune(sshTarget)
	defer sshTarget.Close()

	// Pemotong sampah payload kotor lu tetep nangkring aman disini bos
	if idx := bytes.Index(rawPayload, []byte("SSH-")); idx != -1 {
		_, _ = sshTarget.Write(rawPayload[idx:])
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

	// Jalur A: HP -> SSH Server (AI DETECTOR + AUTOMATIC FRESH RECOVERY)
	go func() {
		defer wg.Done()
		buf := make([]byte, 65536)
		sshHandshakeFound := false
		
		for {
			if isWS && !sshHandshakeFound {
				client.SetReadDeadline(time.Now().Add(5 * time.Second))
			} else {
				client.SetReadDeadline(time.Now().Add(120 * time.Second))
			}

			n, err := client.Read(buf)
			if err != nil {
				// 🧠 Jika terjadi timeout / putus sebelum handshake SSH beres, laporkan ke AI Tracker
				if isWS && !sshHandshakeFound {
					// Jika sudah 5x berturut-turut stuck, paksa tidur 1 detik untuk membersihkan jalur
					if checkSmartRecovery(client.RemoteAddr().String(), true) {
						time.Sleep(1 * time.Second) 
					}
				}
				break
			}
			
			data := buf[:n]
			
			if isWS && !sshHandshakeFound {
				if idx := bytes.Index(data, []byte("SSH-")); idx != -1 {
					data = data[idx:]
					sshHandshakeFound = true
					client.SetReadDeadline(time.Time{})
					
					// 🧠 Berhasil konek bersih! Reset tracker kecerdasan ke 0
					checkSmartRecovery(client.RemoteAddr().String(), false)
				} else {
					continue 
				}
			}

			_, err = target.Write(data)
			if err != nil {
				break
			}
		}
		once.Do(closeAll)
	}()

	// Jalur B: SSH Server -> HP (Full Speed Download + Heartbeat Engine)
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
