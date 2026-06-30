package main

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"runtime"
	"strconv"
	"sync"
	"syscall"
	"time"
)

//go:embed index.html
var indexHTML []byte //embed index.html

// creates stuct and JSON for scan results
type PortResult struct {
	Port    int    `json:"port"`
	Status  string `json:"status"`
	Service string `json:"service"`
}

// Common port service names
var commonServices = map[int]string{
	21:    "FTP",
	22:    "SSH",
	23:    "Telnet",
	25:    "SMTP",
	53:    "DNS",
	80:    "HTTP",
	110:   "POP3",
	135:   "RPC",
	139:   "NetBIOS",
	143:   "IMAP",
	443:   "HTTPS",
	445:   "SMB",
	3306:  "MySQL",
	3389:  "RDP",
	5432:  "PostgreSQL",
	6379:  "Redis",
	8080:  "HTTP-Alt",
	8443:  "HTTPS-Alt",
	27017: "MongoDB",
}

// returns service name
func getService(port int) string {
	if service, ok := commonServices[port]; ok {
		return service
	}
	return "Unknown"
}

// checks to see if port is open or closed
func checkPort(ip string, port int) PortResult {
	address := fmt.Sprintf("%s:%d", ip, port)
	conn, err := net.DialTimeout("tcp", address, 1*time.Second)
	if err != nil {
		return PortResult{Port: port, Status: "closed", Service: getService(port)}
	}
	conn.Close()
	//returns port number status and service
	return PortResult{Port: port, Status: "open", Service: getService(port)}
}

// handles UI
func scanHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	ip := r.URL.Query().Get("ip")
	startPort, _ := strconv.Atoi(r.URL.Query().Get("start"))
	endPort, _ := strconv.Atoi(r.URL.Query().Get("end"))

	if ip == "" {
		ip = "127.0.0.1"
	}
	if startPort == 0 {
		startPort = 1
	}
	if endPort == 0 {
		endPort = 1024
	}

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming not supported", http.StatusInternalServerError)
		return
	}

	ports := make(chan int, 100)
	results := make(chan PortResult, 100)
	var wg sync.WaitGroup

	// Launch 150 workers
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for port := range ports {
				results <- checkPort(ip, port)
			}
		}()
	}

	// Feed ports
	go func() {
		for port := startPort; port <= endPort; port++ {
			ports <- port
		}
		close(ports)
	}()

	// Close results when all workers done
	go func() {
		wg.Wait()
		close(results)
	}()

	// Stream results to client
	for result := range results {
		data, _ := json.Marshal(result)
		fmt.Fprintf(w, "data: %s\n\n", data)
		flusher.Flush()
	}

	fmt.Fprintf(w, "data: {\"done\": true}\n\n")
	flusher.Flush()
}

func openBrowser(url string) {
	var cmd *exec.Cmd

	switch runtime.GOOS { //opens browser diffrently based on OS
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	case "darwin": // Mac
		cmd = exec.Command("open", url)
	case "linux":
		cmd = exec.Command("xdg-open", url)
	}

	if cmd != nil {
		cmd.Start()
	}
}

func main() {
	// Declare quit channel FIRST before anything uses it
	quit := make(chan os.Signal, 1)
	signal.Notify(quit,
		syscall.SIGINT,
		syscall.SIGTERM,
	)

	// Track last heartbeat time
	lastHeartbeat := time.Now()

	// Heartbeat handler  browser pings this every 3 seconds
	http.HandleFunc("/heartbeat", func(w http.ResponseWriter, r *http.Request) {
		lastHeartbeat = time.Now()
		w.Write([]byte("ok"))
	})

	// Watchdog goroutine  shuts down if heartbeat stops
	go func() {
		for {
			time.Sleep(5 * time.Second)
			if time.Since(lastHeartbeat) > 10*time.Second {
				fmt.Println("Browser disconnected - shutting down")
				quit <- syscall.SIGTERM
				return
			}
		}
	}()

	// Quit route
	http.HandleFunc("/quit", func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("Quit signal received!")
		w.Write([]byte("Shutting down..."))
		go func() {
			time.Sleep(100 * time.Millisecond)
			quit <- syscall.SIGTERM
		}()
	})

	// Scan handler
	http.HandleFunc("/scan", scanHandler)

	// Serve embedded HTML
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.Write(indexHTML)
	})

	// Create server
	server := &http.Server{Addr: ":8080"}

	// Start server in goroutine
	go func() {
		fmt.Println("Port Scanner running at http://localhost:8080")
		server.ListenAndServe()
	}()

	// Open browser after short delay
	go func() {
		time.Sleep(500 * time.Millisecond)
		openBrowser("http://localhost:8080")
	}()

	// Block until quit signal received
	<-quit
	fmt.Println("Shutting down...")
	server.Close()
	os.Exit(0)
}
