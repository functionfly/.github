#!/bin/bash
# FunctionFly Edge VPS Deployment Script
# Run this on each VPS server (217.160.124.206, 209.46.125.113)

set -e

echo "=== FunctionFly Edge Deployment ==="

# Configuration - UPDATE THIS with your shared secret
export FFLY_SHARED_SECRET="${FFLY_SHARED_SECRET:-YOUR_SECRET_HERE}"
export BACKEND_URL="${BACKEND_URL:-https://api.functionfly.com}"
export PORT=8080

echo "Backend URL: $BACKEND_URL"

# Check if Go is installed
if ! command -v go &> /dev/null; then
    echo "Installing Go..."
    apt-get update && apt-get install -y golang-go
fi

# Create application directory
mkdir -p /opt/functionfly-edge
cd /opt/functionfly-edge

# Download the edge proxy source (or clone from repo)
echo "Downloading FunctionFly Edge source..."
if command -v git &> /dev/null; then
    # Clone if git available
    git clone --depth 1 https://github.com/functionfly/functionfly.git /tmp/functionfly || true
    cp -r /tmp/functionfly/edge-targets/functionfly-edge/* . 2>/dev/null || true
fi

# If no source, create a minimal version
if [ ! -f main.go ]; then
    echo "Creating edge proxy from embedded source..."
    cat > main.go << 'GOEOF'
package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"time"
)

var (
	sharedSecret = os.Getenv("FFLY_SHARED_SECRET")
	backendURL   = os.Getenv("BACKEND_URL")
	port         = os.Getenv("PORT")
)

func main() {
	if sharedSecret == "" {
		log.Fatal("FFLY_SHARED_SECRET environment variable is required")
	}
	if backendURL == "" {
		backendURL = "https://api.functionfly.com"
	}
	if port == "" {
		port = "8080"
	}

	http.HandleFunc("/healthz", healthHandler)
	http.HandleFunc("/ping", pingHandler)
	http.HandleFunc("/", proxyHandler)

	log.Printf("Starting FunctionFly Edge on port %s", port)
	log.Printf("Backend URL: %s", backendURL)

	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal(err)
	}
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, "OK")
}

func pingHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, `{"status":"pong","timestamp":"%s","provider":"functionfly-edge"}`, time.Now().UTC().Format(time.RFC3339))
}

func proxyHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/healthz" || r.URL.Path == "/ping" {
		http.NotFound(w, r)
		return
	}

	if !verifyHMAC(r) {
		log.Printf("Invalid HMAC signature from %s", r.RemoteAddr)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Fprintf(w, `{"error":"Invalid signature","timestamp":"%s"}`, time.Now().UTC().Format(time.RFC3339))
		return
	}

	req, err := http.NewRequest(r.Method, backendURL+r.URL.Path, r.Body)
	if err != nil {
		log.Printf("Error creating request: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, `{"error":"Failed to create request","timestamp":"%s"}`, time.Now().UTC().Format(time.RFC3339))
		return
	}

	for key, values := range r.Header {
		for _, value := range values {
			if key == "X-FFLY-Timestamp" || key == "X-FFLY-Signature" {
				continue
			}
			req.Header.Add(key, value)
		}
	}
	req.Host = getHost(backendURL)

	client := &http.Client{Timeout: 30 * time.Second, CheckRedirect: func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	}}

	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Error proxying request: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadGateway)
		fmt.Fprintf(w, `{"error":"Proxy request failed","message":"%s","timestamp":"%s"}`, err.Error(), time.Now().UTC().Format(time.RFC3339))
		return
	}
	defer resp.Body.Close()

	for key, values := range resp.Header {
		for _, value := range values {
			w.Header().Add(key, value)
		}
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Error reading response body: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(resp.StatusCode)
	w.Write(body)
}

func verifyHMAC(r *http.Request) bool {
	timestamp := r.Header.Get("X-FFLY-Timestamp")
	signature := r.Header.Get("X-FFLY-Signature")

	if timestamp == "" || signature == "" {
		return false
	}

	ts, err := strconv.ParseInt(timestamp, 10, 64)
	if err != nil {
		return false
	}

	now := time.Now().Unix()
	if now-ts > 300 || now-ts < -300 {
		return false
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		return false
	}

	bodyHash := sha256.Sum256(body)
	bodyHashHex := hex.EncodeToString(bodyHash[:])

	sigString := fmt.Sprintf("%d%s%s%s", ts, r.Method, r.URL.Path, bodyHashHex)

	mac := hmac.New(sha256.New, []byte(sharedSecret))
	mac.Write([]byte(sigString))
	expected := hex.EncodeToString(mac.Sum(nil))

	return hmac.Equal([]byte(signature), []byte(expected))
}

func getHost(urlStr string) string {
	u, err := url.Parse(urlStr)
	if err != nil {
		return ""
	}
	return u.Host
}
GOEOF
fi

# Build the binary
echo "Building FunctionFly Edge..."
go build -o functionfly-edge .

# Create systemd service
echo "Creating systemd service..."
cat > /etc/systemd/system/functionfly-edge.service << EOF
[Unit]
Description=FunctionFly Edge Proxy
After=network.target

[Service]
Type=simple
User=root
WorkingDirectory=/opt/functionfly-edge
Environment=FFLY_SHARED_SECRET=${FFLY_SHARED_SECRET}
Environment=BACKEND_URL=${BACKEND_URL}
Environment=PORT=${PORT}
ExecStart=/opt/functionfly-edge/functionfly-edge
Restart=always
RestartSec=5

[Install]
WantedBy=multi-user.target
EOF

# Start the service
echo "Starting FunctionFly Edge..."
systemctl daemon-reload
systemctl enable functionfly-edge
systemctl start functionfly-edge

# Check status
echo ""
echo "=== Deployment Complete ==="
systemctl status functionfly-edge --no-pager

echo ""
echo "Testing health endpoint..."
sleep 2
curl -s http://localhost:8080/healthz || echo "Health check failed"
echo ""
curl -s http://localhost:8080/ping || echo "Ping failed"
