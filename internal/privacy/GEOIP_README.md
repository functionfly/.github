# GeoIP Integration for Privacy-Preserving Region Detection

This module integrates MaxMind GeoLite2 for accurate, privacy-preserving IP-to-region mapping, which is used for:

- GDPR compliance (detecting EU users)
- Privacy-preserving execution logging (storing region instead of full IP)
- Regional data residency compliance

## Features

- **MaxMind GeoLite2 Country** - Free tier available (CC BY-SA 4.0 license)
- **Privacy-preserving** - Returns broad regions (US, EU, APAC, etc.) rather than specific locations
- **Automatic downloads** - Auto-downloads database on startup if license key is configured
- **Auto-update support** - Check for and download database updates periodically
- **Graceful fallback** - Falls back to simplified detection if database unavailable

## Setup

### 1. Get a Free MaxMind License Key

1. Sign up for a free account at: <https://www.maxmind.com/en/geolite2/signup>
2. Generate a license key in your account dashboard
3. Set the environment variable:

   ```bash
   export MAXMIND_LICENSE_KEY="your_license_key_here"
   ```

### 2. Configuration Options

| Environment Variable | Default | Description |
|---------------------|---------|-------------|
| `MAXMIND_LICENSE_KEY` | (none) | Your MaxMind license key (required for downloads) |
| `GEOLITE2_DB_PATH` | `./data/GeoLite2-Country.mmdb` | Path to store the database |
| `GEOLITE2_AUTO_UPDATE` | `false` | Enable automatic updates (checks every 7 days) |

### 3. First Run

On first startup with `MAXMIND_LICENSE_KEY` set, the service will:

1. Check if the database exists locally
2. If not, automatically download from MaxMind
3. Extract and load the database
4. Use it for accurate region detection

```bash
# Example startup output with license key:
INFO[0000] GeoIP service initialized for accurate region detection
INFO[0000] GeoLite2 database loaded successfully

# Without license key:
WARN[0000] MAXMIND_LICENSE_KEY not set, using simplified region detection.
      For accurate region detection, get a free key at https://www.maxmind.com/en/geolite2/signup
```

## Manual Database Download

If you prefer to manually manage the database:

1. Download from: <https://dev.maxmind.com/geoip/geolite2-free-geolocation-data>
2. Extract the `.mmdb` file
3. Place it at the path specified by `GEOLITE2_DB_PATH`
4. Restart the service

## Privacy Considerations

- **No precise geolocation** - Only broad regions (US, EU, APAC, etc.) are stored
- **No street-level data** - GeoLite2 Country only provides country-level data
- **GDPR compliant** - Country-level data is generally not considered PII
- **Regional aggregation** - For extra privacy, EU countries are grouped as "EU"

## API Usage

The GeoIP service is automatically used by the privacy service for region detection:

```go
// The privacy service automatically uses GeoIP if configured
privacyRepo := privacy.NewRepository(postgresDB)
privacyService := privacy.NewService(privacyRepo, salt)

// Region detection happens automatically in anonymization
region := privacyService.GetRegionFromIP("8.8.8.8") // Returns "US"
```

## Alternative: IP2Location LITE

If you prefer IP2Location:

1. Download from: <https://lite.ip2location.com> (free, CC BY-SA 4.0)
2. Use the DB1.LITE database (country-level only)
3. Modify `geoip_service.go` to use IP2Location's Go library instead

## License Compliance

MaxMind GeoLite2 is licensed under CC BY-SA 4.0. You must:

- Attribute MaxMind in your privacy policy or about page
- Link to <https://www.maxmind.com>
- Include the attribution: "This product includes GeoLite2 data created by MaxMind"

## Testing

The service gracefully degrades when the database is unavailable, using simplified IP range detection as a fallback. This ensures privacy features work even without the license key.
