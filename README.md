# GoScan TCP Port Scanner

A fast, concurrent TCP port scanner built in Go with a browser-based GUI featuring live results, filtering, and search functionality.

## Features

- **Concurrent scanning** — uses Go goroutines and a worker pool to scan ports in parallel
- **Live results** — ports stream into the table in real time as they are discovered
- **Filtering** — toggle between All, Open, and Closed ports instantly
- **Search** — search by port number or service name (e.g. "HTTP", "SSH")
- **Service detection** — automatically identifies common services (FTP, SSH, HTTP, HTTPS, RDP, and more)
- **Progress tracking** — live progress bar and open/closed port counts while scanning
- **Auto-launch** — opens the browser automatically on startup, no manual navigation needed
- **Auto-shutdown** — server shuts itself down when the browser tab is closed

## Usage

Download the binary for your platform from the releases section and run it directly no dependencies or installation required.

### Windows
```
PortScanner-windows.exe
```

### Linux
```
PortScanner-linux
```

### Mac
```
PortScanner-mac
```

The app will open automatically in your browser at `http://localhost:8080`.

Enter a target IP, set your port range, and click **Scan**.

## Building from Source

Requires [Go](https://go.dev/dl/) 1.16 or later.

```bash
git clone https://github.com/JoshuaMFrench/goscan.git
cd goscan
go build -o PortScanner .
```

**Cross-compile for all platforms (from Windows PowerShell):**
```bash
$env:GOOS="windows"; $env:GOARCH="amd64"; go build -ldflags="-s -w" -o PortScanner-windows.exe .
$env:GOOS="linux";   $env:GOARCH="amd64"; go build -ldflags="-s -w" -o PortScanner-linux .
$env:GOOS="darwin";  $env:GOARCH="amd64"; go build -ldflags="-s -w" -o PortScanner-mac .
```

## Tech Stack

- **Backend**  Go (net, net/http, sync, embed)
- **Concurrency**  goroutine worker pool with channels
- **Streaming**  Server-Sent Events (SSE) for live results
- **Frontend**  HTML, CSS, JavaScript (GUI designed with generative AI)
