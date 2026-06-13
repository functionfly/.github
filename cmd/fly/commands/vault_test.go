package commands

import (
	"testing"
	"time"
)

func TestIsUUID(t *testing.T) {
	cases := []struct {
		in   string
		want bool
	}{
		{"550e8400-e29b-41d4-a716-446655440000", true},
		{"550E8400-E29B-41D4-A716-446655440000", true},
		{"", false},
		{"not-a-uuid", false},
		{"550e8400-e29b-41d4-a716", false},          // too short
		{"550e8400-e29b-41d4-a716-44665544000g", false}, // bad char
		{"550e8400e29b41d4a716446655440000", false},    // no dashes
	}
	for _, c := range cases {
		if got := isUUID(c.in); got != c.want {
			t.Errorf("isUUID(%q)=%v, want %v", c.in, got, c.want)
		}
	}
}

func TestParseExpiryHours(t *testing.T) {
	cases := []struct {
		in      string
		want    int
		wantErr bool
	}{
		{"1h", 1, false},
		{"24h", 24, false},
		{"720h", 720, false},
		{"8760h", 8760, false},
		{"30m", 1, false},    // round up to 1h
		{"0h", 0, true},
		{"-1h", 0, true},
		{"9000h", 0, true},  // > 1 year
		{"bogus", 0, true},
	}
	for _, c := range cases {
		got, err := parseExpiryHours(c.in)
		if (err != nil) != c.wantErr {
			t.Errorf("parseExpiryHours(%q) err=%v, wantErr=%v", c.in, err, c.wantErr)
			continue
		}
		if !c.wantErr && got != c.want {
			t.Errorf("parseExpiryHours(%q)=%d, want %d", c.in, got, c.want)
		}
	}
}

func TestParseExpiryHours_Recent(t *testing.T) {
	got, err := parseExpiryHours("24h")
	if err != nil {
		t.Fatal(err)
	}
	if time.Duration(got)*time.Hour != 24*time.Hour {
		t.Fatalf("24h mismatch: got %d hours", got)
	}
}
