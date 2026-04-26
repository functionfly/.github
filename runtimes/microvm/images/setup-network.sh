#!/usr/bin/env bash
# setup-network.sh — Set up tap devices for Firecracker MicroVMs.
#
# Usage:
#   sudo ./setup-network.sh [--create] [--destroy] [--cleanup]
#
# This script creates the tap devices that Firecracker VMs connect to.
# It should be run once at host startup or deployment time.
#
# For Kubernetes: run as a DaemonSet initContainer before the orchestrator.
set -euo pipefail

ACTION="${1:-create}"
TAP_BASE="${TAP_BASE:-tap}"
TAP_COUNT="${TAP_COUNT:-10}"
NETWORK_BRIDGE="${NETWORK_BRIDGE:-br0}"
HOST_IP="${HOST_IP:-172.31.0.1}"

usage() {
    echo "Usage: $0 [--create|--destroy|--cleanup] [--tap-base tap] [--tap-count 10]"
    echo "  --create    Create tap devices (default)"
    echo "  --destroy    Remove all tap devices"
    echo "  --cleanup    Remove and recreate tap devices"
    echo ""
    echo "Environment variables:"
    echo "  TAP_BASE     Base name for tap devices (default: tap)"
    echo "  TAP_COUNT    Number of tap devices to create (default: 10)"
    echo "  NETWORK_BRIDGE  Bridge to attach taps to (default: br0)"
    echo "  HOST_IP      Host IP for the network (default: 172.31.0.1)"
    exit 1
}

create_tap() {
    local tap="$1"
    local ip="$2"

    if ip link show "${tap}" &>/dev/null; then
        echo "  ${tap}: already exists"
        return 0
    fi

    # Create tap device
    ip tuntap add "${tap}" mode tap user "$(id -u)" 2>/dev/null || \
    ip tuntap add "${tap}" mode tap 2>/dev/null || {
        echo "  ${tap}: failed to create (need root)"
        return 1
    }

    # Set device up
    ip link set "${tap}" up

    # Don't bridge by default - let firecracker handle it
    # Bridge setup should be done separately if needed
    echo "  ${tap}: created"
}

destroy_tap() {
    local tap="$1"

    if ip link show "${tap}" &>/dev/null; then
        ip link del "${tap}" 2>/dev/null || true
        echo "  ${tap}: removed"
    fi
}

setup_bridge() {
    local bridge="$1"
    local host_ip="$2"

    if ! ip link show "${bridge}" &>/dev/null; then
        echo "  Creating bridge ${bridge}..."
        ip link add name "${bridge}" type bridge
        ip addr add "${host_ip}/24" dev "${bridge}"
        ip link set "${bridge}" up
        echo "  ${bridge}: created with IP ${host_ip}/24"
    else
        echo "  ${bridge}: already exists"
    fi
}

case "$ACTION" in
    --create|-c|create)
        echo "==> Creating ${TAP_COUNT} tap devices..."
        for i in $(seq 0 $((TAP_COUNT - 1))); do
            create_tap "${TAP_BASE}${i}" "${HOST_IP}"
        done
        echo "==> Done. Tap devices ready for Firecracker."
        ;;

    --destroy|-d|destroy)
        echo "==> Destroying tap devices..."
        for i in $(seq 0 $((TAP_COUNT - 1))); do
            destroy_tap "${TAP_BASE}${i}"
        done
        echo "==> Done."
        ;;

    --cleanup|-C|cleanup)
        echo "==> Cleaning up tap devices..."
        for i in $(seq 0 $((TAP_COUNT - 1))); do
            destroy_tap "${TAP_BASE}${i}"
        done
        echo "==> Recreating tap devices..."
        for i in $(seq 0 $((TAP_COUNT - 1))); do
            create_tap "${TAP_BASE}${i}" "${HOST_IP}"
        done
        echo "==> Done."
        ;;

    --help|-h)
        usage
        ;;
    *)
        echo "Unknown action: $ACTION"
        usage
        ;;
esac