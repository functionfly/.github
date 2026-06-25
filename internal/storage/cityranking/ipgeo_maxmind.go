package cityranking

import (
	"context"
	"fmt"
	"net"
	"os"
	"sync"

	"github.com/oschwald/geoip2-golang"
)

// MaxMindDBPath is the environment variable and default relative path for
// the MaxMind GeoLite2-Country database. The IP-geo fallback path (plan §2)
// is a no-op when this file is missing — the rest of the leaderboard keeps
// working from self-reported `Location`.
const MaxMindDBPath = "data/geoip/GeoLite2-Country.mmdb"

// OpenMaxMindDB opens the MaxMind GeoLite2-Country database if it exists.
// Returns (nil, nil) when the file is missing so callers can degrade
// gracefully.
func OpenMaxMindDB() (*geoip2.Reader, error) {
	path := os.Getenv("GEOIP_DATABASE_PATH")
	if path == "" {
		path = MaxMindDBPath
	}
	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("stat geoip db %s: %w", path, err)
	}
	r, err := geoip2.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open geoip db %s: %w", path, err)
	}
	return r, nil
}

// MaxMindCountryLookup returns a CountryLookup backed by a MaxMind reader.
// The returned function is safe for concurrent use; the reader is closed via
// the returned close function. If `r` is nil the lookup is a no-op
// (returns "", err) so the IP-geo resolver can still be wired in dev
// environments without a downloaded DB.
func MaxMindCountryLookup(r *geoip2.Reader) CountryLookup {
	if r == nil {
		return func(ctx context.Context, ip string) (string, error) {
			return "", fmt.Errorf("maxmind db not loaded")
		}
	}
	var mu sync.RWMutex
	return func(ctx context.Context, ip string) (string, error) {
		mu.RLock()
		defer mu.RUnlock()
		return lookupCountryFromReader(r, ip)
	}
}

func lookupCountryFromReader(r *geoip2.Reader, ip string) (string, error) {
	parsed := net.ParseIP(ip)
	if parsed == nil {
		return "", fmt.Errorf("invalid ip: %q", ip)
	}
	rec, err := r.Country(parsed)
	if err != nil {
		return "", fmt.Errorf("maxmind lookup: %w", err)
	}
	if rec == nil || rec.Country.IsoCode == "" {
		return "", fmt.Errorf("no country for ip")
	}
	return rec.Country.IsoCode, nil
}
