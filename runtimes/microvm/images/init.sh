#!/bin/sh
# OpenRC init script for FunctionFly MicroVM
# This replaces systemd for Firecracker boot

# Mount essential filesystems
mount -t proc none /proc
mount -t sysfs none /sys
mount -t tmpfs none /dev
mount -t tmpfs none /run

# Create essential device nodes
mknod -m 666 /dev/null c 1 3
mknod -m 666 /dev/zero c 1 5
mknod -m 666 /dev/console c 5 1
mknod -m 666 /dev/ttyS0 c 4 64

# Set up hostname
hostname functionfly

# Set up networking via DHCP for eth0 (if not disabled)
if [ ! -f /etc/no_network ]; then
    udhcpc -b -i eth0 -s /etc/simple.script 2>/dev/null || true
fi

# Initialize Python path
export PYTHONPATH=/opt/functionfly-packages:/tmp/function-packages

# Run the function agent
exec /usr/local/bin/function-agent serve