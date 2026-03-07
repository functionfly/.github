# GoBetterAuth MFA Plugin (Phase 2)

Multi-Factor Authentication plugin for GoBetterAuth using TOTP (Time-based One-Time Password).

## Features

- **TOTP Support**: Compatible with Google Authenticator, Authy, Microsoft Authenticator, and other RFC 6238 compliant apps
- **Backup Codes**: 10 single-use recovery codes for account recovery
- **Clock Skew Tolerance**: ±1 period (±30 seconds by default) tolerance for time drift
- **Secure Storage**: TOTP secrets are encrypted, backup codes are bcrypt hashed
- **Easy Integration**: Simple API endpoints for setup and verification

## Installation

The MFA plugin is automatically included with GoBetterAuth. No additional installation required.

## Configuration

Add the following to your `.env` file:

```bash
# MFA Master Switch
GBA_MFA_ENABLED=true

# Require MFA during account setup (optional)
GBA_MFA_REQUIRE_ON_SETUP=false

# TOTP Configuration
TOTP_ISSUER=FunctionFly              # Display name in authenticator apps
TOTP_PERIOD=30                       # Code validity period in seconds
TOTP_DIGITS=6                        # Code length (6 or 8 digits)
TOTP_SKEW_PERIODS=1                  # Clock skew tolerance (±1 period)
```

## Usage

### Initialize the Plugin

```go
import (
    "github.com/functionfly/internal/auth/gba/plugins/mfa"
)

// Create MFA plugin
mfaConfig := &mfa.MFAConfig{
    Enabled:        true,
    RequireOnSetup: false,
    TOTPIssuer:     "FunctionFly",
    TOTPPeriod:     30,
    TOTPDigits:     6,
    SkewPeriods:    1,
}

mfaPlugin, err := mfa.New(db, mfaConfig, logger)
if err != nil {
    log.Fatal(err)
}
```

### Setup HTTP Handlers

```go
// Create handler
mfaHandler := mfa.NewHandler(mfaPlugin, logger)

// Register routes (typically under /v1/auth/mfa)
mux := http.NewServeMux()
mfaHandler.SetupRoutes(mux, "/v1/auth/mfa")
```

## API Endpoints

### Setup MFA
```http
POST /v1/auth/mfa/setup
Authorization: Bearer <token>

Response:
{
  "secret": "JBSWY3DPEHPK3PXP",
  "qr_code_url": "otpauth://totp/FunctionFly:user@example.com?secret=JBSWY3DPEHPK3PXP&issuer=FunctionFly",
  "backup_codes": ["ABCD-EFGH-IJKL", "MNOP-QRST-UVWX", ...],
  "message": "Scan the QR code with your authenticator app and verify a code to enable MFA"
}
```

### Verify and Enable MFA
```http
POST /v1/auth/mfa/verify
Authorization: Bearer <token>
Content-Type: application/json

{
  "code": "123456"
}

Response:
{
  "enabled": true,
  "message": "MFA enabled successfully"
}
```

### MFA Challenge (During Login)
```http
POST /v1/auth/mfa/challenge
Authorization: Bearer <token>
Content-Type: application/json

{
  "code": "123456"
  // OR use backup_code:
  // "backup_code": "ABCD-EFGH-IJKL"
}

Response:
{
  "valid": true,
  "message": "MFA verified successfully"
}
```

### Disable MFA
```http
POST /v1/auth/mfa/disable
Authorization: Bearer <token>
Content-Type: application/json

{
  "code": "123456"
  // OR use backup_code:
  // "backup_code": "ABCD-EFGH-IJKL"
}

Response:
{
  "disabled": true,
  "message": "MFA disabled successfully"
}
```

### Regenerate Backup Codes
```http
POST /v1/auth/mfa/backup-codes/regenerate
Authorization: Bearer <token>
Content-Type: application/json

{
  "code": "123456"
}

Response:
{
  "backup_codes": ["XXXX-XXXX-XXXX", "YYYY-YYYY-YYYY", ...]
}
```

### Get MFA Status
```http
GET /v1/auth/mfa/status
Authorization: Bearer <token>

Response:
{
  "enabled": true,
  "verified": true,
  "has_backup_codes": true,
  "backup_code_count": 10
}
```

## Integration with Login Flow

To require MFA during login:

```go
func (h *Handler) HandleLogin(w http.ResponseWriter, r *http.Request) {
    // ... validate credentials ...

    // Check if MFA is enabled
    enabled, err := mfaPlugin.IsEnabledForUser(ctx, userID)
    if err != nil {
        // Handle error
    }

    if enabled {
        // Return MFA required response
        // Frontend should prompt for MFA code
        respondJSON(w, http.StatusOK, map[string]interface{}{
            "mfa_required": true,
            "message": "Please provide MFA code",
        })
        return
    }

    // Complete login
}
```

## Backup Codes

- 10 single-use codes are generated during MFA setup
- Codes are in format: `XXXX-XXXX-XXXX` (base32 encoded)
- Each code can only be used once
- Codes are bcrypt hashed before storage (plaintext is only shown once during setup)
- Codes can be regenerated using the TOTP code

## Security Considerations

1. **Secret Storage**: TOTP secrets should be encrypted with a master key in production (consider using AES encryption)
2. **Backup Codes**: Store only bcrypt hashes, never plaintext
3. **Rate Limiting**: Implement rate limiting on MFA endpoints to prevent brute force attacks
4. **Session Management**: Mark sessions as MFA-verified after successful challenge
5. **Clock Skew**: Default ±30 seconds tolerance handles most time drift issues

## Database Schema

### gba_mfa_totp
| Column | Type | Description |
|--------|------|-------------|
| id | UUID | Primary key |
| user_id | UUID | Foreign key to gba_users |
| secret | TEXT | Encrypted TOTP secret |
| enabled | BOOLEAN | Whether MFA is enabled |
| verified | BOOLEAN | Whether setup is complete |
| created_at | TIMESTAMP | Creation time |
| updated_at | TIMESTAMP | Last update time |

### gba_mfa_backup_codes
| Column | Type | Description |
|--------|------|-------------|
| id | UUID | Primary key |
| user_id | UUID | Foreign key to gba_users |
| code_hash | TEXT | Bcrypt hash of backup code |
| used | BOOLEAN | Whether code has been used |
| used_at | TIMESTAMP | When code was used |
| created_at | TIMESTAMP | Creation time |

## Migration

Run the migration to create MFA tables:

```bash
# Using golang-migrate
migrate -path migrations -database "postgres://user:pass@localhost/dbname" up

# Or apply manually
psql -d functionfly -f migrations/20260307000003_add_gba_mfa_tables.up.sql
```

## Testing

Example test flow:

```go
func TestMFAFlow(t *testing.T) {
    ctx := context.Background()
    userID := uuid.MustParse("...")

    // 1. Setup MFA
    setup, err := mfaPlugin.GenerateTOTP(ctx, userID, "test@example.com")
    require.NoError(t, err)
    require.NotEmpty(t, setup.Secret)
    require.NotEmpty(t, setup.QRCodeURL)
    require.Len(t, setup.BackupCodes, 10)

    // 2. Generate TOTP code (using same secret)
    code, err := totp.GenerateCode(setup.Secret, time.Now())
    require.NoError(t, err)

    // 3. Verify and enable
    err = mfaPlugin.VerifyAndEnableTOTP(ctx, userID, code)
    require.NoError(t, err)

    // 4. Verify code works
    err = mfaPlugin.VerifyCode(ctx, userID, code)
    require.NoError(t, err)
}
```

## License

MIT License - Same as GoBetterAuth
