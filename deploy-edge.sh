#!/bin/bash
# FunctionFly Edge Deployment Script

# Install Go if not present
if ! command -v go &> /dev/null; then
    echo "Installing Go..."
    apt-get update && apt-get install -y golang-go
fi

# Clone the repo
cd /opt
git clone https://github.com/functionfly/functionfly.git
cd functionfly/edge-targets/functionfly-edge

# Build
go build -o functionfly-edge .

# Create systemd service
cat > /etc/systemd/system/functionfly-edge.service << EOF
[Unit]
Description=FunctionFly Edge Proxy
After=network.target

[Service]
Type=simple
User=root
WorkingDirectory=/opt/functionfly/edge-targets/functionfly-edge
Environment=FFLY_SHARED_SECRET=596a0b1247133a345c00ce7dc7509ab982612ea16d637c7f18f16321ebb29afa
Environment=BACKEND_URL=https://api.functionfly.com
Environment=PORT=8080
ExecStart=/opt/functionfly/edge-targets/functionfly-edge/functionfly-edge
Restart=always
RestartSec=5

[Install]
WantedBy=multi-user.target
EOF

# Start the service
systemctl daemon-reload
systemctl enable functionfly-edge
systemctl start functionfly-edge

# Check status
systemctl status functionfly-edge
