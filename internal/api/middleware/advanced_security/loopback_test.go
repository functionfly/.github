package advanced_security

import "testing"

func TestIsLoopbackIP(t *testing.T) {
	tests := []struct {
		ip       string
		loopback bool
	}{
		{"[::1]:36126", true},
		{"[::1]:46098", true},
		{"127.0.0.1:8080", true},
		{"::1", true},
		{"127.0.0.1", true},
		{"192.168.1.1:443", false},
		{"10.0.0.1:80", false},
	}
	for _, tt := range tests {
		got := isLoopbackIP(tt.ip)
		if got != tt.loopback {
			t.Errorf("isLoopbackIP(%q) = %v, want %v", tt.ip, got, tt.loopback)
		}
	}
}
