package wallet

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strconv"
	"time"

	"github.com/sirupsen/logrus"
)

// WalletEncryption provides AES-256-GCM encryption for sensitive wallet data
// This is separate from the Vault encryption (which is client-side) - this is server-side
// for protecting wallet balances at rest in the database
type WalletEncryption struct {
	key        []byte
	enabled    bool
	keyVersion int
}

// EncryptedBalance represents an encrypted wallet balance
type EncryptedBalance struct {
	Ciphertext string `json:"ciphertext"` // base64 encoded
	IV         string `json:"iv"`         // base64 encoded
	Tag        string `json:"tag"`        // base64 encoded
	KeyVersion int    `json:"key_version"`
}

// AdminAdjustmentLimit represents limits for admin balance adjustments
type AdminAdjustmentLimit struct {
	// SingleOperationMax is the maximum amount for a single adjustment without secondary approval
	SingleOperationMax float64
	// DailyMax is the maximum total adjustments per admin per day
	DailyMax float64
	// RequiresSecondaryApprovalAbove is the threshold requiring a second admin approval
	RequiresSecondaryApprovalAbove float64
	// AlertThreshold is the amount that triggers monitoring alerts
	AlertThreshold float64
}

// Global configuration (loaded from environment)
var (
	adminAdjustmentLimits AdminAdjustmentLimit
	walletEncryption      *WalletEncryption
	auditHMACKey          = getEnvString("WALLET_AUDIT_HMAC_KEY", "default-audit-key-change-in-production")
)

func init() {
	// Load admin adjustment limits from environment
	adminAdjustmentLimits = AdminAdjustmentLimit{
		SingleOperationMax:             getEnvFloat64Safe("WALLET_ADMIN_ADJUSTMENT_SINGLE_MAX", 1000.0),
		DailyMax:                       getEnvFloat64Safe("WALLET_ADMIN_ADJUSTMENT_DAILY_MAX", 10000.0),
		RequiresSecondaryApprovalAbove: getEnvFloat64Safe("WALLET_ADMIN_ADJUSTMENT_SECONDARY_THRESHOLD", 5000.0),
		AlertThreshold:                 getEnvFloat64Safe("WALLET_ADMIN_ADJUSTMENT_ALERT_THRESHOLD", 1000.0),
	}

	// Initialize wallet encryption
	walletEncryption = initWalletEncryption()
}

// getEnvFloat64Safe parses a float64 from environment variable with default
func getEnvFloat64Safe(key string, defaultVal float64) float64 {
	val := os.Getenv(key)
	if val == "" {
		return defaultVal
	}
	f, err := strconv.ParseFloat(val, 64)
	if err != nil {
		return defaultVal
	}
	return f
}

// initWalletEncryption initializes the wallet encryption from environment
func initWalletEncryption() *WalletEncryption {
	keyStr := os.Getenv("WALLET_ENCRYPTION_KEY")
	if keyStr == "" {
		// Check if we're in production - encryption is mandatory in prod
		if os.Getenv("PRODUCTION") == "true" || os.Getenv("ENVIRONMENT") == "production" {
			logrus.Error("WALLET_ENCRYPTION_KEY not set in production - wallet encryption disabled but this is a security risk!")
		}
		logrus.Warn("WALLET_ENCRYPTION_KEY not set - wallet data will be stored unencrypted")
		return &WalletEncryption{enabled: false, keyVersion: 0}
	}

	// Decode base64 key
	key, err := base64.StdEncoding.DecodeString(keyStr)
	if err != nil {
		logrus.WithError(err).Error("Failed to decode WALLET_ENCRYPTION_KEY - wallet encryption disabled")
		return &WalletEncryption{enabled: false, keyVersion: 0}
	}

	// Must be 32 bytes for AES-256
	if len(key) != 32 {
		logrus.Errorf("WALLET_ENCRYPTION_KEY must be 32 bytes (got %d) - wallet encryption disabled", len(key))
		return &WalletEncryption{enabled: false, keyVersion: 0}
	}

	return &WalletEncryption{
		key:        key,
		enabled:    true,
		keyVersion: 1,
	}
}

// IsEnabled returns true if encryption is enabled
func (we *WalletEncryption) IsEnabled() bool {
	return we != nil && we.enabled
}

// EncryptBalance encrypts a balance value for storage
func (we *WalletEncryption) EncryptBalance(balance float64) (*EncryptedBalance, error) {
	if !we.IsEnabled() {
		return nil, nil
	}

	// Convert balance to string for consistent encoding
	balanceStr := strconv.FormatFloat(balance, 'f', 4, 64)

	// Create AES-GCM cipher
	block, err := aes.NewCipher(we.key)
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM: %w", err)
	}

	// Generate random IV (nonce)
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, fmt.Errorf("failed to generate nonce: %w", err)
	}

	// Encrypt the data
	plaintext := []byte(balanceStr)
	ciphertext := gcm.Seal(nil, nonce, plaintext, nil)

	// Split ciphertext and tag (GCM appends tag to ciphertext)
	tagSize := 16 // GCM uses 16-byte tag
	if len(ciphertext) < tagSize {
		return nil, fmt.Errorf("ciphertext too short")
	}

	actualCiphertext := ciphertext[:len(ciphertext)-tagSize]
	tag := ciphertext[len(ciphertext)-tagSize:]

	return &EncryptedBalance{
		Ciphertext: base64.StdEncoding.EncodeToString(actualCiphertext),
		IV:         base64.StdEncoding.EncodeToString(nonce),
		Tag:        base64.StdEncoding.EncodeToString(tag),
		KeyVersion: we.keyVersion,
	}, nil
}

// DecryptBalance decrypts an encrypted balance value
func (we *WalletEncryption) DecryptBalance(encrypted *EncryptedBalance) (float64, error) {
	if !we.IsEnabled() || encrypted == nil {
		return 0, nil
	}

	// Decode base64 values
	ciphertext, err := base64.StdEncoding.DecodeString(encrypted.Ciphertext)
	if err != nil {
		return 0, fmt.Errorf("failed to decode ciphertext: %w", err)
	}

	nonce, err := base64.StdEncoding.DecodeString(encrypted.IV)
	if err != nil {
		return 0, fmt.Errorf("failed to decode IV: %w", err)
	}

	tag, err := base64.StdEncoding.DecodeString(encrypted.Tag)
	if err != nil {
		return 0, fmt.Errorf("failed to decode tag: %w", err)
	}

	// Combine ciphertext and tag for GCM
	ciphertextWithTag := append(ciphertext, tag...)

	// Create AES-GCM cipher
	block, err := aes.NewCipher(we.key)
	if err != nil {
		return 0, fmt.Errorf("failed to create cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return 0, fmt.Errorf("failed to create GCM: %w", err)
	}

	// Decrypt
	plaintext, err := gcm.Open(nil, nonce, ciphertextWithTag, nil)
	if err != nil {
		return 0, fmt.Errorf("failed to decrypt: %w", err)
	}

	// Parse balance
	balance, err := strconv.ParseFloat(string(plaintext), 64)
	if err != nil {
		return 0, fmt.Errorf("failed to parse decrypted balance: %w", err)
	}

	return balance, nil
}

// EncryptWalletBalances encrypts all balance fields in a wallet
// Returns JSON string of encrypted data to store in BalanceEncryptedJSON column
func (we *WalletEncryption) EncryptWalletBalances(wallet *Wallet) (string, error) {
	if !we.IsEnabled() || wallet == nil {
		return "", nil
	}

	// Encrypt main balance
	encryptedBalance, err := we.EncryptBalance(wallet.BalanceUSD)
	if err != nil {
		return "", fmt.Errorf("failed to encrypt balance: %w", err)
	}

	// Store encrypted balance as JSON string
	if encryptedBalance != nil {
		jsonData, err := json.Marshal(encryptedBalance)
		if err != nil {
			return "", fmt.Errorf("failed to marshal encrypted balance: %w", err)
		}
		return string(jsonData), nil
	}

	return "", nil
}

// DecryptWalletBalances decrypts balance from JSON string
func (we *WalletEncryption) DecryptWalletBalances(encryptedJSON string) (float64, error) {
	if !we.IsEnabled() || encryptedJSON == "" {
		return 0, nil
	}

	var encryptedBalance EncryptedBalance
	if err := json.Unmarshal([]byte(encryptedJSON), &encryptedBalance); err != nil {
		return 0, fmt.Errorf("failed to unmarshal encrypted balance: %w", err)
	}

	balance, err := we.DecryptBalance(&encryptedBalance)
	if err != nil {
		return 0, fmt.Errorf("failed to decrypt balance: %w", err)
	}

	return balance, nil
}

// GetAdminAdjustmentLimits returns the configured admin adjustment limits
func GetAdminAdjustmentLimits() AdminAdjustmentLimit {
	return adminAdjustmentLimits
}

// GetWalletEncryption returns the global wallet encryption instance
func GetWalletEncryption() *WalletEncryption {
	return walletEncryption
}

// CheckAdminAdjustmentAllowed checks if an admin adjustment is allowed based on limits
func CheckAdminAdjustmentAllowed(adminID string, amount float64, dailyTotal float64) (bool, string) {
	limits := GetAdminAdjustmentLimits()

	// Check single operation max
	if amount > limits.SingleOperationMax {
		// This requires secondary approval, not a hard fail
		if amount <= limits.RequiresSecondaryApprovalAbove {
			return true, "requires_secondary_approval"
		}
		return false, fmt.Sprintf("amount $%.2f exceeds maximum single operation limit of $%.2f",
			amount, limits.SingleOperationMax)
	}

	// Check daily max
	if dailyTotal+amount > limits.DailyMax {
		return false, fmt.Sprintf("daily adjustment limit of $%.2f would be exceeded (current: $%.2f, proposed: $%.2f)",
			limits.DailyMax, dailyTotal, amount)
	}

	// Check if secondary approval is required
	if amount > limits.RequiresSecondaryApprovalAbove {
		return true, "requires_secondary_approval"
	}

	return true, ""
}

// GenerateWalletEncryptionKey generates a new 32-byte encryption key for wallets
// This should be used to create the WALLET_ENCRYPTION_KEY environment variable
func GenerateWalletEncryptionKey() (string, error) {
	key := make([]byte, 32)
	if _, err := io.ReadFull(rand.Reader, key); err != nil {
		return "", fmt.Errorf("failed to generate key: %w", err)
	}
	return base64.StdEncoding.EncodeToString(key), nil
}

// SecureCompare performs a constant-time comparison of two strings
// to prevent timing attacks
func SecureCompare(a, b string) bool {
	return subtle.ConstantTimeCompare([]byte(a), []byte(b)) == 1
}

// HashForAudit creates a hash suitable for audit logging
// This is a one-way hash for integrity verification
func HashForAudit(data string) string {
	h := hmac.New(sha256.New, []byte(auditHMACKey))
	h.Write([]byte(data))
	return hex.EncodeToString(h.Sum(nil))
}

// WalletSecurityConfig holds all wallet security configuration
type WalletSecurityConfig struct {
	EncryptionEnabled              bool    `json:"encryption_enabled"`
	KeyVersion                     int     `json:"key_version"`
	SingleOperationMax             float64 `json:"single_operation_max"`
	DailyMax                       float64 `json:"daily_max"`
	RequiresSecondaryApprovalAbove float64 `json:"requires_secondary_approval_above"`
	AlertThreshold                 float64 `json:"alert_threshold"`
}

// GetSecurityConfig returns the current security configuration
func GetSecurityConfig() WalletSecurityConfig {
	limits := GetAdminAdjustmentLimits()
	enc := GetWalletEncryption()

	return WalletSecurityConfig{
		EncryptionEnabled:              enc != nil && enc.IsEnabled(),
		KeyVersion:                     enc.keyVersion,
		SingleOperationMax:             limits.SingleOperationMax,
		DailyMax:                       limits.DailyMax,
		RequiresSecondaryApprovalAbove: limits.RequiresSecondaryApprovalAbove,
		AlertThreshold:                 limits.AlertThreshold,
	}
}

// ValidateEncryptionSetup checks if encryption is properly configured
func ValidateEncryptionSetup() error {
	enc := GetWalletEncryption()

	if !enc.IsEnabled() {
		isProd := os.Getenv("PRODUCTION") == "true" || os.Getenv("ENVIRONMENT") == "production"
		if isProd {
			return fmt.Errorf("wallet encryption is not enabled in production - set WALLET_ENCRYPTION_KEY")
		}
		logrus.Warn("Wallet encryption is not enabled - balances stored in plaintext")
		return nil
	}

	logrus.Info("Wallet encryption is enabled with AES-256-GCM")
	return nil
}

// SecondaryApprovalRecord tracks pending secondary approvals
type SecondaryApprovalRecord struct {
	ID              string     `json:"id"`
	WalletID        string     `json:"wallet_id"`
	AdminID         string     `json:"admin_id"`
	AmountUSD       float64    `json:"amount_usd"`
	Reason          string     `json:"reason"`
	Reference       string     `json:"reference"`
	RequestedAt     time.Time  `json:"requested_at"`
	SecondAdminID   *string    `json:"second_admin_id,omitempty"`
	ApprovedAt      *time.Time `json:"approved_at,omitempty"`
	RejectedAt      *time.Time `json:"rejected_at,omitempty"`
	RejectionReason *string    `json:"rejection_reason,omitempty"`
}

// ToJSON serializes the record to JSON for storage
func (sar *SecondaryApprovalRecord) ToJSON() ([]byte, error) {
	return json.Marshal(sar)
}

// ParseSecondaryApprovalRecord deserializes from JSON
func ParseSecondaryApprovalRecord(data []byte) (*SecondaryApprovalRecord, error) {
	var record SecondaryApprovalRecord
	if err := json.Unmarshal(data, &record); err != nil {
		return nil, err
	}
	return &record, nil
}
