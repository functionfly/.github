package privacy

import (
	"compress/gzip"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/oschwald/geoip2-golang"
	"github.com/sirupsen/logrus"
)

const (
	// GeoLite2CountryURL is the direct download URL for GeoLite2 Country (free tier)
	// Users need to sign up at https://www.maxmind.com/en/geolite2/signup for a free license key
	GeoLite2CountryURL = "https://download.maxmind.com/app/geoip_download?edition_id=GeoLite2-Country&suffix=tar.gz&license_key=%s"

	// GeoLite2CountryMD5URL is the MD5 checksum URL
	GeoLite2CountryMD5URL = "https://download.maxmind.com/app/geoip_download?edition_id=GeoLite2-Country&suffix=tar.gz.md5&license_key=%s"

	// DefaultDBPath is the default path to store the GeoLite2 database
	DefaultDBPath = "./data/GeoLite2-Country.mmdb"

	// UpdateInterval is how often to check for updates (7 days)
	UpdateInterval = 7 * 24 * time.Hour
)

// GeoIPService provides geo-IP lookup functionality using MaxMind GeoLite2
type GeoIPService struct {
	reader     *geoip2.Reader
	dbPath     string
	licenseKey string
	lastUpdate time.Time
	countryMap map[string]string // ISO code -> region mapping
}

// GeoIPConfig holds configuration for the GeoIP service
type GeoIPConfig struct {
	DBPath           string
	LicenseKey       string
	AutoUpdate       bool
	UpdateInterval   time.Duration
	EUCountries      []string // Countries considered part of EU for GDPR purposes
}

// DefaultGeoIPConfig returns a default configuration
func DefaultGeoIPConfig() *GeoIPConfig {
	return &GeoIPConfig{
		DBPath:         getEnvOrDefault("GEOLITE2_DB_PATH", DefaultDBPath),
		LicenseKey:     os.Getenv("MAXMIND_LICENSE_KEY"),
		AutoUpdate:     os.Getenv("GEOLITE2_AUTO_UPDATE") == "true",
		UpdateInterval: UpdateInterval,
		EUCountries: []string{
			"AT", "BE", "BG", "HR", "CY", "CZ", "DK", "EE", "FI", "FR",
			"DE", "GR", "HU", "IE", "IT", "LV", "LT", "LU", "MT", "NL",
			"PL", "PT", "RO", "SK", "SI", "ES", "SE", "GB", // EU member states
			"IS", "LI", "NO", "CH", // EEA/EFTA states
		},
	}
}

// NewGeoIPService creates a new GeoIP service
func NewGeoIPService(config *GeoIPConfig) (*GeoIPService, error) {
	if config == nil {
		config = DefaultGeoIPConfig()
	}

	service := &GeoIPService{
		dbPath:     config.DBPath,
		licenseKey: config.LicenseKey,
		countryMap: make(map[string]string),
	}

	// Build EU country map for quick lookup
	for _, code := range config.EUCountries {
		service.countryMap[code] = "EU"
	}

	// Ensure data directory exists
	if err := os.MkdirAll(filepath.Dir(config.DBPath), 0755); err != nil {
		return nil, fmt.Errorf("failed to create data directory: %w", err)
	}

	// Check if database exists, if not try to download
	if _, err := os.Stat(config.DBPath); os.IsNotExist(err) {
		if config.LicenseKey != "" {
			logrus.Info("GeoLite2 database not found, attempting download...")
			if err := service.DownloadDatabase(); err != nil {
				logrus.WithError(err).Warn("Failed to download GeoLite2 database, using fallback region detection")
				// Continue without database - will use fallback
			}
		} else {
			logrus.Warn("MAXMIND_LICENSE_KEY not set, GeoLite2 database download skipped. " +
				"Get a free license key at https://www.maxmind.com/en/geolite2/signup")
		}
	}

	// Try to open the database if it exists
	if _, err := os.Stat(config.DBPath); err == nil {
		reader, err := geoip2.Open(config.DBPath)
		if err != nil {
			return nil, fmt.Errorf("failed to open GeoLite2 database: %w", err)
		}
		service.reader = reader
		service.lastUpdate = time.Now()
		logrus.Info("GeoLite2 database loaded successfully")
	}

	return service, nil
}

// GetRegionFromIP returns a privacy-preserving region code for the given IP
func (s *GeoIPService) GetRegionFromIP(ip string) string {
	parsedIP := net.ParseIP(ip)
	if parsedIP == nil {
		return "unknown"
	}

	// Private IPs
	if IsPrivateIP(ip) {
		return "private"
	}

	// If we have the GeoLite2 database, use it
	if s.reader != nil {
		country, err := s.reader.Country(parsedIP)
		if err != nil {
			logrus.WithError(err).Debug("Failed to lookup IP in GeoLite2 database")
			return s.fallbackRegionDetection(ip)
		}

		isoCode := country.Country.IsoCode
		if isoCode == "" {
			return "unknown"
		}

		// Check if EU country
		if _, isEU := s.countryMap[isoCode]; isEU {
			return "EU"
		}

		// Map specific countries to regions
		return s.mapCountryToRegion(isoCode)
	}

	// Fallback to simple detection
	return s.fallbackRegionDetection(ip)
}

// mapCountryToRegion maps a country ISO code to a broader region
func (s *GeoIPService) mapCountryToRegion(isoCode string) string {
	// North America
	naCountries := []string{"US", "CA", "MX", "GL", "BM", "BS", "CU", "JM", "HT", "DO"}
	for _, c := range naCountries {
		if isoCode == c {
			return "NA"
		}
	}

	// Asia-Pacific
	apacCountries := []string{
		"CN", "JP", "IN", "KR", "ID", "TH", "VN", "MY", "PH", "SG",
		"AU", "NZ", "TW", "HK", "MO", "BD", "PK", "LK", "NP", "MM",
		"KH", "LA", "BN", "PG", "FJ", "NC", "VU", "SB", "KI", "TV",
		"NR", "PW", "MH", "FM", "AS", "GU", "MP", "PF", "WF", "CK",
		"NU", "TO", "WS", "TK", "PN", "WF",
	}
	for _, c := range apacCountries {
		if isoCode == c {
			return "APAC"
		}
	}

	// South America
	saCountries := []string{
		"BR", "AR", "CL", "CO", "PE", "VE", "EC", "BO", "PY", "UY",
		"GY", "SR", "GF", "FK", "GS",
	}
	for _, c := range saCountries {
		if isoCode == c {
			return "SA"
		}
	}

	// Africa
	afCountries := []string{
		"ZA", "NG", "EG", "KE", "ET", "GH", "DZ", "MA", "UG", "TZ",
		"MZ", "ZW", "BW", "ZM", "MW", "MG", "CM", "CI", "NE", "BF",
		"ML", "SN", "TD", "GN", "RW", "BI", "SL", "TG", "LR", "MR",
		"GM", "CV", "ST", "GQ", "GA", "CG", "CD", "AO", "NA", "SZ",
		"LS", "ER", "DJ", "SO", "SS", "SD", "CF", "TD", "LY", "TN",
		"EH", "YT", "RE", "SC", "MU", "KM", "MG",
	}
	for _, c := range afCountries {
		if isoCode == c {
			return "AF"
		}
	}

	// Middle East
	meCountries := []string{
		"SA", "AE", "IL", "TR", "IR", "IQ", "SY", "JO", "LB", "KW",
		"QA", "BH", "OM", "YE", "AZ", "AM", "GE", "CY", "PS", "JO",
	}
	for _, c := range meCountries {
		if isoCode == c {
			return "ME"
		}
	}

	// Russia and former Soviet states (excluding EU ones)
	if isoCode == "RU" || isoCode == "BY" || isoCode == "KZ" || isoCode == "UZ" ||
		isoCode == "TM" || isoCode == "KG" || isoCode == "TJ" || isoCode == "MN" {
		return "CIS"
	}

	// Default to the country code itself if no region mapping
	return isoCode
}

// fallbackRegionDetection provides basic region detection without GeoLite2
// This is used as a fallback when the database is not available
func (s *GeoIPService) fallbackRegionDetection(ip string) string {
	parsedIP := net.ParseIP(ip)
	if parsedIP == nil {
		return "unknown"
	}

	// IPv6 detection
	if parsedIP.To4() == nil {
		// Check for common IPv6 prefixes
		ipStr := parsedIP.String()
		if strings.HasPrefix(ipStr, "2001:4860:") || // Google
			strings.HasPrefix(ipStr, "2600:") || // US AWS
			strings.HasPrefix(ipStr, "2601:") || // US Comcast
			strings.HasPrefix(ipStr, "2602:") || // US
			strings.HasPrefix(ipStr, "2603:") || // US
			strings.HasPrefix(ipStr, "2604:") || // US
			strings.HasPrefix(ipStr, "2605:") || // US
			strings.HasPrefix(ipStr, "2606:") || // US
			strings.HasPrefix(ipStr, "2607:") || // US
			strings.HasPrefix(ipStr, "2608:") || // US
			strings.HasPrefix(ipStr, "2609:") { // US
			return "US"
		}
		if strings.HasPrefix(ipStr, "2a02:") || // EU
			strings.HasPrefix(ipStr, "2a01:") || // EU
			strings.HasPrefix(ipStr, "2001:4c28:") { // EU
			return "EU"
		}
		if strings.HasPrefix(ipStr, "2400:") || // APAC
			strings.HasPrefix(ipStr, "2401:") ||
			strings.HasPrefix(ipStr, "2402:") ||
			strings.HasPrefix(ipStr, "2403:") ||
			strings.HasPrefix(ipStr, "2404:") ||
			strings.HasPrefix(ipStr, "2405:") ||
			strings.HasPrefix(ipStr, "2406:") ||
			strings.HasPrefix(ipStr, "2407:") ||
			strings.HasPrefix(ipStr, "2408:") ||
			strings.HasPrefix(ipStr, "2409:") ||
			strings.HasPrefix(ipStr, "240a:") ||
			strings.HasPrefix(ipStr, "240b:") ||
			strings.HasPrefix(ipStr, "240c:") ||
			strings.HasPrefix(ipStr, "240d:") ||
			strings.HasPrefix(ipStr, "240e:") ||
			strings.HasPrefix(ipStr, "240f:") {
			return "APAC"
		}
		return "unknown"
	}

	// IPv4 detection based on common ranges (very approximate)
	ipStr := parsedIP.String()

	// US ranges (simplified)
	if strings.HasPrefix(ipStr, "3.") || strings.HasPrefix(ipStr, "13.") ||
		strings.HasPrefix(ipStr, "34.") || strings.HasPrefix(ipStr, "35.") ||
		strings.HasPrefix(ipStr, "52.") || strings.HasPrefix(ipStr, "54.") ||
		strings.HasPrefix(ipStr, "18.") || strings.HasPrefix(ipStr, "50.") {
		return "US"
	}

	// EU ranges (simplified)
	if strings.HasPrefix(ipStr, "2.") || strings.HasPrefix(ipStr, "5.") ||
		strings.HasPrefix(ipStr, "31.") || strings.HasPrefix(ipStr, "46.") ||
		strings.HasPrefix(ipStr, "51.") || strings.HasPrefix(ipStr, "77.") ||
		strings.HasPrefix(ipStr, "78.") || strings.HasPrefix(ipStr, "79.") ||
		strings.HasPrefix(ipStr, "80.") || strings.HasPrefix(ipStr, "81.") ||
		strings.HasPrefix(ipStr, "82.") || strings.HasPrefix(ipStr, "83.") ||
		strings.HasPrefix(ipStr, "84.") || strings.HasPrefix(ipStr, "85.") ||
		strings.HasPrefix(ipStr, "86.") || strings.HasPrefix(ipStr, "87.") ||
		strings.HasPrefix(ipStr, "88.") || strings.HasPrefix(ipStr, "89.") ||
		strings.HasPrefix(ipStr, "90.") || strings.HasPrefix(ipStr, "91.") {
		return "EU"
	}

	// APAC ranges (simplified)
	if strings.HasPrefix(ipStr, "1.") || strings.HasPrefix(ipStr, "36.") ||
		strings.HasPrefix(ipStr, "42.") || strings.HasPrefix(ipStr, "43.") ||
		strings.HasPrefix(ipStr, "49.") || strings.HasPrefix(ipStr, "58.") ||
		strings.HasPrefix(ipStr, "59.") || strings.HasPrefix(ipStr, "60.") ||
		strings.HasPrefix(ipStr, "61.") || strings.HasPrefix(ipStr, "101.") ||
		strings.HasPrefix(ipStr, "106.") || strings.HasPrefix(ipStr, "110.") ||
		strings.HasPrefix(ipStr, "111.") || strings.HasPrefix(ipStr, "112.") ||
		strings.HasPrefix(ipStr, "113.") || strings.HasPrefix(ipStr, "114.") ||
		strings.HasPrefix(ipStr, "115.") || strings.HasPrefix(ipStr, "116.") ||
		strings.HasPrefix(ipStr, "117.") || strings.HasPrefix(ipStr, "118.") ||
		strings.HasPrefix(ipStr, "119.") || strings.HasPrefix(ipStr, "120.") ||
		strings.HasPrefix(ipStr, "121.") || strings.HasPrefix(ipStr, "122.") ||
		strings.HasPrefix(ipStr, "123.") || strings.HasPrefix(ipStr, "124.") ||
		strings.HasPrefix(ipStr, "125.") || strings.HasPrefix(ipStr, "126.") ||
		strings.HasPrefix(ipStr, "175.") || strings.HasPrefix(ipStr, "180.") ||
		strings.HasPrefix(ipStr, "182.") || strings.HasPrefix(ipStr, "183.") {
		return "APAC"
	}

	return "unknown"
}

// DownloadDatabase downloads the GeoLite2 database from MaxMind
func (s *GeoIPService) DownloadDatabase() error {
	if s.licenseKey == "" {
		return fmt.Errorf("MAXMIND_LICENSE_KEY not set - get a free key at https://www.maxmind.com/en/geolite2/signup")
	}

	url := fmt.Sprintf(GeoLite2CountryURL, s.licenseKey)

	logrus.WithField("url", strings.ReplaceAll(url, s.licenseKey, "***")).
		Info("Downloading GeoLite2 database...")

	// Download the tarball
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("failed to download GeoLite2: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download failed with status: %s", resp.Status)
	}

	// Create temporary file
	tmpFile, err := os.CreateTemp("", "geolite2-*.tar.gz")
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	defer os.Remove(tmpFile.Name())

	// Copy to temp file
	if _, err := io.Copy(tmpFile, resp.Body); err != nil {
		tmpFile.Close()
		return fmt.Errorf("failed to save download: %w", err)
	}
	tmpFile.Close()

	// Extract the database
	if err := s.extractDatabase(tmpFile.Name()); err != nil {
		return fmt.Errorf("failed to extract database: %w", err)
	}

	logrus.Info("GeoLite2 database downloaded and extracted successfully")
	return nil
}

// extractDatabase extracts the mmdb file from the tarball
func (s *GeoIPService) extractDatabase(tarballPath string) error {
	file, err := os.Open(tarballPath)
	if err != nil {
		return err
	}
	defer file.Close()

	gzReader, err := gzip.NewReader(file)
	if err != nil {
		return fmt.Errorf("failed to create gzip reader: %w", err)
	}
	defer gzReader.Close()

	// Create destination file
	destFile, err := os.Create(s.dbPath)
	if err != nil {
		return fmt.Errorf("failed to create destination file: %w", err)
	}
	defer destFile.Close()

	// Copy decompressed data
	if _, err := io.Copy(destFile, gzReader); err != nil {
		return fmt.Errorf("failed to extract database: %w", err)
	}

	return nil
}

// CheckForUpdate checks if a new version is available
func (s *GeoIPService) CheckForUpdate() (bool, error) {
	if s.licenseKey == "" {
		return false, fmt.Errorf("MAXMIND_LICENSE_KEY not set")
	}

	// Get remote MD5
	md5URL := fmt.Sprintf(GeoLite2CountryMD5URL, s.licenseKey)
	resp, err := http.Get(md5URL)
	if err != nil {
		return false, fmt.Errorf("failed to fetch remote checksum: %w", err)
	}
	defer resp.Body.Close()

	remoteMD5, err := io.ReadAll(resp.Body)
	if err != nil {
		return false, fmt.Errorf("failed to read remote checksum: %w", err)
	}

	// Calculate local MD5
	localMD5, err := s.calculateLocalMD5()
	if err != nil {
		return false, err
	}

	// Compare
	return string(remoteMD5) != localMD5, nil
}

// calculateLocalMD5 calculates the MD5 checksum of the local database
func (s *GeoIPService) calculateLocalMD5() (string, error) {
	file, err := os.Open(s.dbPath)
	if err != nil {
		return "", fmt.Errorf("failed to open local database: %w", err)
	}
	defer file.Close()

	hash := md5.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", fmt.Errorf("failed to calculate checksum: %w", err)
	}

	return hex.EncodeToString(hash.Sum(nil)), nil
}

// Close closes the GeoIP database reader
func (s *GeoIPService) Close() error {
	if s.reader != nil {
		return s.reader.Close()
	}
	return nil
}

// getEnvOrDefault returns the value of an environment variable or a default
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
