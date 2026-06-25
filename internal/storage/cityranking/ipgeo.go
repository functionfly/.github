package cityranking

import (
	"context"
	"fmt"
	"net"
	"strings"
	"sync"
	"time"
)

// IPGeoResolver looks up the most-likely city for an IP address. It is the
// fallback path for "where am I?" when the user hasn't set a `Location` on
// their profile (plan §2: "IP-based geo is a fallback when `Location` is
// empty"). It does NOT replace self-reported city — it only kicks in when
// `Location` is missing.
//
// The resolver is two-stage:
//  1. IP → country (via a pluggable CountryLookup; the default wiring uses the
//     MaxMind GeoLite2 client that already exists in `internal/routing`).
//  2. country → city (largest active city in that country, from our seed).
//
// The country→city step is a single SQL query (cacheable in Redis for a day)
// so the cost per request is O(1) DB reads after warm-up.
type IPGeoResolver struct {
	repo           *Repository
	countryLookup  CountryLookup
	cache          *ipGeoCache
	unknownCountry string // city slug to return when the country has no known city
}

// CountryLookup is the IP→country primitive. Any function that takes an IP
// string and returns an ISO-3166-1 alpha-2 country code (or an error) satisfies
// it. The production wiring is routing.GeoIPClient.LookupCountry; tests pass
// a stub.
type CountryLookup func(ctx context.Context, ip string) (countryCode string, err error)

// NewIPGeoResolver wires a resolver. If `countryLookup` is nil, the resolver
// always returns "no IP geo available" so the caller can fall through to a
// 4xx / "set your location manually" UX.
func NewIPGeoResolver(repo *Repository, countryLookup CountryLookup) *IPGeoResolver {
	return &IPGeoResolver{
		repo:           repo,
		countryLookup:  countryLookup,
		cache:          newIPGeoCache(time.Hour),
		unknownCountry: "",
	}
}

// IPGeoResult is what we hand back from Resolve.
type IPGeoResult struct {
	City        *CityRanking `json:"city,omitempty"`
	CountryCode string       `json:"country_code"`
	Source      string       `json:"source"` // "ip" or "default"
	NotFound    bool         `json:"not_found"`
}

// Resolve returns the largest city in the country that the IP maps to. The
// resolver is purely best-effort: errors are mapped to NotFound so the
// caller can decide what to do (typically: prompt the user to type their
// location manually).
func (r *IPGeoResolver) Resolve(ctx context.Context, ip string) (*IPGeoResult, error) {
	if r == nil || r.countryLookup == nil {
		return &IPGeoResult{NotFound: true, Source: "default"}, nil
	}
	ip = strings.TrimSpace(ip)
	if net.ParseIP(ip) == nil {
		return nil, fmt.Errorf("invalid IP: %q", ip)
	}
	country, err := r.countryLookup(ctx, ip)
	if err != nil {
		return &IPGeoResult{NotFound: true, Source: "ip"}, nil
	}
	country = strings.ToUpper(strings.TrimSpace(country))
	if country == "" {
		return &IPGeoResult{NotFound: true, Source: "ip"}, nil
	}

	city, err := r.largestCityInCountry(ctx, country)
	if err != nil {
		return nil, err
	}
	if city == nil {
		return &IPGeoResult{CountryCode: country, NotFound: true, Source: "ip"}, nil
	}
	return &IPGeoResult{City: city, CountryCode: country, Source: "ip"}, nil
}

func (r *IPGeoResolver) largestCityInCountry(ctx context.Context, country string) (*CityRanking, error) {
	if cached, ok := r.cache.get(country); ok {
		return cached, nil
	}
	row := r.repo.Pool().QueryRow(ctx, `
		SELECT c.id, c.slug, c.name, c.state_code, c.state_name, c.country_code, c.country_name,
			COALESCE(c.population, 0),
			m.slug, m.name
		FROM cities c
		LEFT JOIN metro_areas m ON m.id = c.metro_area_id
		WHERE c.is_active = TRUE AND c.country_code = $1
		ORDER BY c.population DESC NULLS LAST
		LIMIT 1
	`, country)
	var city CityRanking
	var metroSlug, metroName *string
	if err := row.Scan(
		&city.CityID, &city.CitySlug, &city.CityName,
		&city.StateCode, &city.StateName, &city.CountryCode, &city.CountryName,
		&city.Population, &metroSlug, &metroName,
	); err != nil {
		if err.Error() == "no rows in result set" {
			r.cache.set(country, nil)
			return nil, nil
		}
		return nil, err
	}
	city.MetroSlug = metroSlug
	city.MetroName = metroName
	r.cache.set(country, &city)
	return &city, nil
}

// ── TTL cache ─────────────────────────────────────────────────────────────

type ipGeoCache struct {
	ttl  time.Duration
	mu   sync.RWMutex
	data map[string]ipGeoEntry
}

type ipGeoEntry struct {
	city   *CityRanking
	expiry time.Time
}

func newIPGeoCache(ttl time.Duration) *ipGeoCache {
	return &ipGeoCache{ttl: ttl, data: map[string]ipGeoEntry{}}
}

func (c *ipGeoCache) get(country string) (*CityRanking, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	e, ok := c.data[country]
	if !ok || time.Now().After(e.expiry) {
		return nil, false
	}
	return e.city, true
}

func (c *ipGeoCache) set(country string, city *CityRanking) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.data[country] = ipGeoEntry{city: city, expiry: time.Now().Add(c.ttl)}
}
