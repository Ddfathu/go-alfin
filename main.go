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

// 🧠 AI SMART RECOVERY DEFINITION
type ClientTracker struct {
	FailCount  int
	LastActive time.Time
}

var (
	trackerMutex sync.Mutex
	clientMap    = make(map[string]*ClientTracker)
)

// 🔄 ZERO-JITTER BUFFER POOL: Daur ulang memori agar Go gak usah "bersih-bersih" mendadak
var bufferPool = sync.Pool{
	New: func() interface{} {
		return make([]byte, 65536)
	},
}

func checkSmartRecovery(remoteAddr string, isFail bool) bool {
	trackerMutex.Lock()
	defer trackerMutex.Unlock()

	ip, _, _ := net.SplitHostPort(remoteAddr)
	
	tracker, exists := clientMap[ip]
	if !exists {
		tracker = &ClientTracker{FailCount: 0, LastActive: time.Now()}
		clientMap[ip] = tracker
	}

	if !isFail {
		tracker.FailCount = 0
		tracker.LastActive = time.Now()
		return false
	}

	tracker.FailCount++
	tracker.LastActive = time.Now()

	if tracker.FailCount >= 5 {
		log.Printf("\033[31m[⚠️ AI DETECT] Aplikasi HP terdeteksi STUCK/TIMEOUT (%d x)! Mengaktifkan Force Auto-Fresh...\033[0m\n", tracker.FailCount)
		tracker.FailCount = 0 
		return true           
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
	// 🔥 AMPUTASI AUTO-GC: Matikan fitur bersih-bersih otomatis bawaan Go agar tidak menahan jaringan
	debug.SetGCPercent(-1)

	// ⏰ PAWANG MEMORI OTOMATIS: Bersih-bersih berjadwal setiap 10 detik agar memori gak bengkak
	go func() {
		for {
			time.Sleep(10 * time.Second)
			runtime.GC() // Panggil pembersihan memori secara paksa di background goroutine
		}
	}()

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

	rawTitle := "⚡ GOLANG TUNNEL PRO: v6.0 ZERO-JITTER AI TURBO ACTIVE ⚡"
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

	// Jalur A: HP -> SSH Server
	go func() {
		defer wg.Done()
		
		// Pinjam buffer dari Pool biar gak bikin sampah memori baru
		buf := bufferPool.Get().([]byte)
		defer bufferPool.Put(buf) // Kembalikan setelah selesai koneksi
		
		sshHandshakeFound := false
		
		for {
			if isWS && !sshHandshakeFound {
				client.SetReadDeadline(time.Now().Add(5 * time.Second))
			} else {
				client.SetReadDeadline(time.Now().Add(120 * time.Second))
			}

			n, err := client.Read(buf)
			if err != nil {
				if isWS && !sshHandshakeFound {
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

	// Jalur B: SSH Server -> HP
	go func() {
		defer wg.Done()
		
		// Pinjam buffer dari Pool untuk download speed los maksimal
		buf := bufferPool.Get().([]byte)
		defer bufferPool.Put(buf)
		
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
