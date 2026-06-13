package commands

import (
	"encoding/base64"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/spf13/cobra"
)

// ============================================================================
// ff vault — Secrets Vault management subcommands
// ============================================================================
//
// The vault uses zero-knowledge client-side encryption, so most write
// operations accept pre-encrypted payloads. The CLI can either:
//   1. Take a base64 ciphertext+iv+salt+tag produced by the dashboard /
//      SDK and pass it through verbatim (default), or
//   2. Encrypt a plaintext locally using the user's stored key
//      (NOT YET IMPLEMENTED — see docs/VAULT_OPERATIONS.md).
//
// The CLI mirrors the API surface in internal/api/handlers/vault/.

func NewVaultCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "vault",
		Short: "Manage zero-knowledge secrets in the FunctionFly vault",
		Long: `Manage secrets stored in the FunctionFly zero-knowledge vault.

The vault encrypts all secret values client-side before they reach the
server, so the server never sees plaintext. When you fetch a secret
the CLI returns the encrypted payload — you are responsible for
decrypting it locally with your passphrase.

Common workflows:

  # List secrets (metadata only)
  ff vault secrets list

  # Create a new secret (you supply the encrypted payload)
  ff vault secrets set MY_KEY \
      --ciphertext <base64> --iv <base64> --salt <base64> --tag <base64> \
      --secret-type api_key

  # Fetch a secret (prints the encrypted payload)
  ff vault secrets get MY_KEY

  # Rotate / delete / import / export
  ff vault secrets rotate MY_KEY --ciphertext ... --iv ... --salt ... --tag ...
  ff vault secrets delete MY_KEY
  ff vault secrets import --from .env.production
  ff vault secrets export --format env

  # Access tokens (for runtime use)
  ff vault tokens create --secret-id <UUID> --expiry 24h
  ff vault tokens list --secret-id <UUID>
  ff vault tokens revoke <TOKEN_ID>

  # Dynamic credentials
  ff vault dynamic-credentials generate <CREDENTIAL_ID>
  ff vault dynamic-credentials targets list

  # Audit
  ff vault audit log --since 24h`,
	}
	cmd.AddCommand(
		newVaultSecretsCmd(),
		newVaultTokensCmd(),
		newVaultDynamicCmd(),
		newVaultAuditCmd(),
	)
	return cmd
}

// --- vault secrets ---------------------------------------------------------

func newVaultSecretsCmd() *cobra.Command {
	cmd := &cobra.Command{Use: "secrets", Short: "Manage vault secrets"}
	cmd.AddCommand(
		newVaultSecretListCmd(),
		newVaultSecretSetCmd(),
		newVaultSecretGetCmd(),
		newVaultSecretRotateCmd(),
		newVaultSecretDeleteCmd(),
		newVaultSecretImportCmd(),
		newVaultSecretExportCmd(),
	)
	return cmd
}

func newVaultSecretListCmd() *cobra.Command {
	var asJSON bool
	var secretType string
	cmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List secrets in the vault (metadata only)",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runVaultSecretList(asJSON, secretType)
		},
	}
	cmd.Flags().BoolVar(&asJSON, "json", false, "Output as JSON")
	cmd.Flags().StringVar(&secretType, "type", "", "Filter by secret type (api_key, oauth_token, password, certificate)")
	return cmd
}

func newVaultSecretSetCmd() *cobra.Command {
	var (
		ciphertext, iv, salt, tag string
		keyVersion                int
		secretType                string
		description               string
		fromFile                  string
	)
	cmd := &cobra.Command{
		Use:   "set NAME",
		Short: "Create a new secret from a pre-encrypted payload",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runVaultSecretSet(args[0], vaultSecretSetInput{
				Ciphertext:  ciphertext,
				IV:          iv,
				Salt:        salt,
				Tag:         tag,
				KeyVersion:  keyVersion,
				SecretType:  secretType,
				Description: description,
				FromFile:    fromFile,
			})
		},
	}
	cmd.Flags().StringVar(&ciphertext, "ciphertext", "", "Base64 AES-256-GCM ciphertext (required)")
	cmd.Flags().StringVar(&iv, "iv", "", "Base64 12-byte IV (required)")
	cmd.Flags().StringVar(&salt, "salt", "", "Base64 KDF salt (required)")
	cmd.Flags().StringVar(&tag, "tag", "", "Base64 GCM auth tag (required)")
	cmd.Flags().IntVar(&keyVersion, "key-version", 1, "KDF version: 1=PBKDF2, 2=Argon2id")
	cmd.Flags().StringVar(&secretType, "secret-type", "api_key", "api_key | oauth_token | password | certificate")
	cmd.Flags().StringVar(&description, "description", "", "Optional human description")
	cmd.Flags().StringVar(&fromFile, "from-file", "", "Read base64 ciphertext from file (alternative to --ciphertext)")
	return cmd
}

func newVaultSecretGetCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get SECRET_ID_OR_NAME",
		Short: "Fetch a secret (returns encrypted payload — decrypt locally)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runVaultSecretGet(args[0])
		},
	}
	return cmd
}

func newVaultSecretRotateCmd() *cobra.Command {
	var (
		ciphertext, iv, salt, tag string
		keyVersion                int
		reason                    string
	)
	cmd := &cobra.Command{
		Use:   "rotate SECRET_ID_OR_NAME",
		Short: "Rotate a secret's encrypted value",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runVaultSecretRotate(args[0], vaultSecretSetInput{
				Ciphertext: ciphertext,
				IV:         iv,
				Salt:       salt,
				Tag:        tag,
				KeyVersion: keyVersion,
			}, reason)
		},
	}
	cmd.Flags().StringVar(&ciphertext, "ciphertext", "", "Base64 ciphertext (required)")
	cmd.Flags().StringVar(&iv, "iv", "", "Base64 IV (required)")
	cmd.Flags().StringVar(&salt, "salt", "", "Base64 salt (required)")
	cmd.Flags().StringVar(&tag, "tag", "", "Base64 tag (required)")
	cmd.Flags().IntVar(&keyVersion, "key-version", 1, "KDF version")
	cmd.Flags().StringVar(&reason, "reason", "", "Optional reason recorded in audit log")
	return cmd
}

func newVaultSecretDeleteCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "delete SECRET_ID_OR_NAME",
		Aliases: []string{"rm"},
		Short:   "Soft-delete a secret and revoke its tokens",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runVaultSecretDelete(args[0])
		},
	}
	return cmd
}

func newVaultSecretImportCmd() *cobra.Command {
	var fromFile string
	cmd := &cobra.Command{
		Use:   "import",
		Short: "Import secrets from a .env file (NAME=VALUE per line; plaintext — encrypt first)",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runVaultSecretImport(fromFile)
		},
	}
	cmd.Flags().StringVar(&fromFile, "from", "", "Path to .env file (required)")
	_ = cmd.MarkFlagRequired("from")
	return cmd
}

func newVaultSecretExportCmd() *cobra.Command {
	var format string
	cmd := &cobra.Command{
		Use:   "export",
		Short: "Export secret metadata (no values)",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runVaultSecretExport(format)
		},
	}
	cmd.Flags().StringVar(&format, "format", "json", "json | env")
	return cmd
}

// --- vault tokens ----------------------------------------------------------

func newVaultTokensCmd() *cobra.Command {
	cmd := &cobra.Command{Use: "tokens", Short: "Manage vault access tokens"}
	cmd.AddCommand(
		newVaultTokenCreateCmd(),
		newVaultTokenListCmd(),
		newVaultTokenRevokeCmd(),
	)
	return cmd
}

func newVaultTokenCreateCmd() *cobra.Command {
	var (
		secretID string
		expiry   string
		scopes   string
		name     string
	)
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a runtime access token (plaintext shown once)",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runVaultTokenCreate(secretID, expiry, scopes, name)
		},
	}
	cmd.Flags().StringVar(&secretID, "secret-id", "", "Secret UUID (required)")
	cmd.Flags().StringVar(&expiry, "expiry", "24h", "Token lifetime (e.g. 1h, 24h, 720h)")
	cmd.Flags().StringVar(&scopes, "scopes", "", "Comma-separated scope list")
	cmd.Flags().StringVar(&name, "name", "", "Optional token label")
	_ = cmd.MarkFlagRequired("secret-id")
	return cmd
}

func newVaultTokenListCmd() *cobra.Command {
	var secretID string
	cmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List tokens for a secret",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runVaultTokenList(secretID)
		},
	}
	cmd.Flags().StringVar(&secretID, "secret-id", "", "Secret UUID (required)")
	_ = cmd.MarkFlagRequired("secret-id")
	return cmd
}

func newVaultTokenRevokeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "revoke TOKEN_ID",
		Short: "Revoke an access token",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runVaultTokenRevoke(args[0])
		},
	}
	return cmd
}

// --- vault dynamic credentials --------------------------------------------

func newVaultDynamicCmd() *cobra.Command {
	cmd := &cobra.Command{Use: "dynamic-credentials", Short: "Dynamic database credentials", Aliases: []string{"dyn"}}
	cmd.AddCommand(
		newVaultDynamicGenerateCmd(),
		newVaultDynamicTargetsCmd(),
	)
	return cmd
}

func newVaultDynamicGenerateCmd() *cobra.Command {
	var ttl int
	cmd := &cobra.Command{
		Use:   "generate CREDENTIAL_ID",
		Short: "Generate a fresh dynamic credential (password shown once)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runVaultDynamicGenerate(args[0], ttl)
		},
	}
	cmd.Flags().IntVar(&ttl, "ttl", 0, "Override TTL in seconds (must not exceed max_ttl_seconds)")
	return cmd
}

func newVaultDynamicTargetsCmd() *cobra.Command {
	cmd := &cobra.Command{Use: "targets", Short: "Manage dynamic-secret DB targets"}
	cmd.AddCommand(
		newVaultDynamicTargetsListCmd(),
	)
	return cmd
}

func newVaultDynamicTargetsListCmd() *cobra.Command {
	var asJSON bool
	cmd := &cobra.Command{
		Use: "list",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runVaultDynamicTargetsList(asJSON)
		},
	}
	cmd.Flags().BoolVar(&asJSON, "json", false, "Output as JSON")
	return cmd
}

// --- vault audit -----------------------------------------------------------

func newVaultAuditCmd() *cobra.Command {
	cmd := &cobra.Command{Use: "audit", Short: "Query vault audit log"}
	cmd.AddCommand(newVaultAuditLogCmd())
	return cmd
}

func newVaultAuditLogCmd() *cobra.Command {
	var (
		since  string
		asJSON bool
	)
	cmd := &cobra.Command{
		Use: "log",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runVaultAuditLog(since, asJSON)
		},
	}
	cmd.Flags().StringVar(&since, "since", "24h", "Look back this far (e.g. 1h, 24h, 7d)")
	cmd.Flags().BoolVar(&asJSON, "json", false, "Output as JSON")
	return cmd
}

// ============================================================================
// Runners
// ============================================================================

type vaultSecretSetInput struct {
	Ciphertext  string
	IV          string
	Salt        string
	Tag         string
	KeyVersion  int
	SecretType  string
	Description string
	FromFile    string
}

type vaultSecretMeta struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description,omitempty"`
	SecretType  string    `json:"secret_type"`
	AccessCount int       `json:"access_count"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type vaultSecretListResp struct {
	Secrets []vaultSecretMeta `json:"secrets"`
	Total   int64             `json:"total"`
	Limit   int               `json:"limit"`
	Offset  int               `json:"offset"`
}

type vaultEncryptedData struct {
	Ciphertext string `json:"ciphertext"`
	IV         string `json:"iv"`
	Salt       string `json:"salt"`
	Tag        string `json:"tag"`
	KeyVersion int    `json:"key_version"`
}

type vaultSecretFullResp struct {
	ID            string             `json:"id"`
	Name          string             `json:"name"`
	Description   string             `json:"description,omitempty"`
	SecretType    string             `json:"secret_type"`
	EncryptedData vaultEncryptedData `json:"encrypted_data"`
	CreatedAt     time.Time          `json:"created_at"`
	UpdatedAt     time.Time          `json:"updated_at"`
}

func runVaultSecretList(asJSON bool, secretType string) error {
	client, err := NewAPIClient()
	if err != nil {
		return err
	}
	path := "/v1/vault/secrets?limit=100"
	if secretType != "" {
		path += "&secret_type=" + secretType
	}
	var out vaultSecretListResp
	if err := client.Get(path, &out); err != nil {
		return fmt.Errorf("could not list secrets: %w", err)
	}
	if asJSON {
		printJSON(out)
		return nil
	}
	if len(out.Secrets) == 0 {
		fmt.Println("No secrets in vault.")
		fmt.Println("   → Add one: ff vault secrets set NAME --ciphertext ...")
		return nil
	}
	tw := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(tw, "ID\tNAME\tTYPE\tACCESSES\tUPDATED")
	for _, s := range out.Secrets {
		fmt.Fprintf(tw, "%s\t%s\t%s\t%d\t%s\n", s.ID, s.Name, s.SecretType, s.AccessCount, s.UpdatedAt.Format(time.RFC3339))
	}
	tw.Flush()
	return nil
}

func runVaultSecretSet(name string, in vaultSecretSetInput) error {
	if in.FromFile != "" {
		raw, err := os.ReadFile(in.FromFile)
		if err != nil {
			return fmt.Errorf("read --from-file: %w", err)
		}
		// .env file expected as key=base64-blob per line; we treat
		// the entire file as a single base64 blob for simplicity.
		in.Ciphertext = strings.TrimSpace(string(raw))
	}
	// Validate required base64 inputs.
	for k, v := range map[string]string{
		"ciphertext": in.Ciphertext,
		"iv":         in.IV,
		"salt":       in.Salt,
		"tag":        in.Tag,
	} {
		if v == "" {
			return fmt.Errorf("--%s is required", k)
		}
		if _, err := base64.StdEncoding.DecodeString(v); err != nil {
			return fmt.Errorf("--%s is not valid base64: %w", k, err)
		}
	}
	if in.SecretType == "" {
		in.SecretType = "api_key"
	}
	client, err := NewAPIClient()
	if err != nil {
		return err
	}
	body := map[string]interface{}{
		"name":        name,
		"description": in.Description,
		"secret_type": in.SecretType,
		"encrypted_data": vaultEncryptedData{
			Ciphertext: in.Ciphertext,
			IV:         in.IV,
			Salt:       in.Salt,
			Tag:        in.Tag,
			KeyVersion: in.KeyVersion,
		},
	}
	var out vaultSecretFullResp
	if err := client.Post("/v1/vault/secrets", body, &out); err != nil {
		return fmt.Errorf("could not create secret: %w", err)
	}
	fmt.Printf("✅ Created secret %q (id=%s)\n", out.Name, out.ID)
	return nil
}

func runVaultSecretGet(identifier string) error {
	id, err := resolveVaultSecretID(identifier)
	if err != nil {
		return err
	}
	client, err := NewAPIClient()
	if err != nil {
		return err
	}
	var out vaultSecretFullResp
	if err := client.Get("/v1/vault/secrets/"+id, &out); err != nil {
		return fmt.Errorf("could not fetch secret: %w", err)
	}
	printJSON(out)
	fmt.Println("\n# Decrypt locally with your passphrase + KDF version", out.EncryptedData.KeyVersion)
	return nil
}

func runVaultSecretRotate(identifier string, in vaultSecretSetInput, reason string) error {
	id, err := resolveVaultSecretID(identifier)
	if err != nil {
		return err
	}
	for k, v := range map[string]string{
		"ciphertext": in.Ciphertext,
		"iv":         in.IV,
		"salt":       in.Salt,
		"tag":        in.Tag,
	} {
		if v == "" {
			return fmt.Errorf("--%s is required", k)
		}
	}
	client, err := NewAPIClient()
	if err != nil {
		return err
	}
	body := map[string]interface{}{
		"reason": reason,
		"encrypted_data": vaultEncryptedData{
			Ciphertext: in.Ciphertext,
			IV:         in.IV,
			Salt:       in.Salt,
			Tag:        in.Tag,
			KeyVersion: in.KeyVersion,
		},
	}
	if err := client.Patch("/v1/vault/secrets/"+id+"/rotate", body, nil); err != nil {
		return fmt.Errorf("could not rotate secret: %w", err)
	}
	fmt.Printf("✅ Rotated secret %s (reason: %q)\n", identifier, reason)
	return nil
}

func runVaultSecretDelete(identifier string) error {
	id, err := resolveVaultSecretID(identifier)
	if err != nil {
		return err
	}
	client, err := NewAPIClient()
	if err != nil {
		return err
	}
	if err := client.Delete("/v1/vault/secrets/"+id, nil); err != nil {
		return fmt.Errorf("could not delete secret: %w", err)
	}
	fmt.Printf("✅ Deleted secret %s\n", identifier)
	return nil
}

func runVaultSecretImport(fromFile string) error {
	raw, err := os.ReadFile(fromFile)
	if err != nil {
		return fmt.Errorf("read --from: %w", err)
	}
	fmt.Println("⚠️  Plaintext import is not yet implemented in the CLI.")
	fmt.Println("   The dashboard and SDK can encrypt-then-import via the API.")
	fmt.Printf("   Read %d bytes from %s\n", len(raw), fromFile)
	return nil
}

func runVaultSecretExport(format string) error {
	client, err := NewAPIClient()
	if err != nil {
		return err
	}
	var out struct {
		Secrets []vaultSecretMeta `json:"secrets"`
		Total   int               `json:"total"`
	}
	if err := client.Get("/v1/vault/secrets/export", &out); err != nil {
		return fmt.Errorf("could not export secrets: %w", err)
	}
	switch strings.ToLower(format) {
	case "env":
		for _, s := range out.Secrets {
			fmt.Printf("# %s (%s)\n", s.Name, s.SecretType)
		}
		fmt.Printf("\n# %d secret(s) exported (values hidden — vault is zero-knowledge)\n", out.Total)
	default:
		printJSON(out)
	}
	return nil
}

// --- token runners ---------------------------------------------------------

type vaultToken struct {
	ID            string     `json:"id"`
	SecretID      string     `json:"secret_id"`
	Name          string     `json:"name,omitempty"`
	ExpiresAt     time.Time  `json:"expires_at"`
	IsRevoked     bool       `json:"is_revoked"`
	RevokedAt     *time.Time `json:"revoked_at,omitempty"`
	LastUsedAt    *time.Time `json:"last_used_at,omitempty"`
	UseCount      int        `json:"use_count"`
	CreatedAt     time.Time  `json:"created_at"`
	RevokedReason string     `json:"revoked_reason,omitempty"`
}

type vaultTokenListResp struct {
	Tokens []vaultToken `json:"tokens"`
	Total  int64        `json:"total"`
}

type vaultTokenCreateResp struct {
	TokenID   string    `json:"token_id"`
	Token     string    `json:"token"`
	SecretID  string    `json:"secret_id"`
	Name      string    `json:"name,omitempty"`
	ExpiresAt time.Time `json:"expires_at"`
	CreatedAt time.Time `json:"created_at"`
}

func runVaultTokenCreate(secretID, expiry, scopes, name string) error {
	hours, err := parseExpiryHours(expiry)
	if err != nil {
		return err
	}
	scopeList := []string{}
	if strings.TrimSpace(scopes) != "" {
		for _, s := range strings.Split(scopes, ",") {
			s = strings.TrimSpace(s)
			if s != "" {
				scopeList = append(scopeList, s)
			}
		}
	}
	client, err := NewAPIClient()
	if err != nil {
		return err
	}
	body := map[string]interface{}{
		"secret_id":        secretID,
		"expires_in_hours": hours,
		"scopes":           scopeList,
		"name":             name,
	}
	var out vaultTokenCreateResp
	if err := client.Post("/v1/vault/secrets/"+secretID+"/tokens", body, &out); err != nil {
		return fmt.Errorf("could not create token: %w", err)
	}
	fmt.Printf("✅ Token created (id=%s, expires %s)\n", out.TokenID, out.ExpiresAt.Format(time.RFC3339))
	fmt.Println("\n🔑 PLAINTEXT TOKEN (shown only once — store it safely):")
	fmt.Println(out.Token)
	return nil
}

func runVaultTokenList(secretID string) error {
	client, err := NewAPIClient()
	if err != nil {
		return err
	}
	var out vaultTokenListResp
	if err := client.Get("/v1/vault/secrets/"+secretID+"/tokens", &out); err != nil {
		return fmt.Errorf("could not list tokens: %w", err)
	}
	if len(out.Tokens) == 0 {
		fmt.Println("No tokens for this secret.")
		return nil
	}
	tw := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(tw, "ID\tNAME\tSTATUS\tUSE_COUNT\tEXPIRES")
	for _, t := range out.Tokens {
		status := "active"
		if t.IsRevoked {
			status = "revoked"
		} else if time.Now().After(t.ExpiresAt) {
			status = "expired"
		}
		fmt.Fprintf(tw, "%s\t%s\t%s\t%d\t%s\n", t.ID, t.Name, status, t.UseCount, t.ExpiresAt.Format(time.RFC3339))
	}
	tw.Flush()
	return nil
}

func runVaultTokenRevoke(tokenID string) error {
	client, err := NewAPIClient()
	if err != nil {
		return err
	}
	if err := client.Delete("/v1/vault/tokens/"+tokenID, nil); err != nil {
		return fmt.Errorf("could not revoke token: %w", err)
	}
	fmt.Printf("✅ Revoked token %s\n", tokenID)
	return nil
}

// --- dynamic runners -------------------------------------------------------

type vaultDynamicCred struct {
	LeaseID    string    `json:"lease_id"`
	Username   string    `json:"username"`
	Password   string    `json:"password"`
	Host       string    `json:"host"`
	Port       int       `json:"port"`
	Database   string    `json:"database"`
	ExpiresAt  time.Time `json:"expires_at"`
	Credential struct {
		Name string `json:"name"`
	} `json:"credential"`
	Target struct {
		Name string `json:"name"`
	} `json:"target"`
}

type vaultDynamicTarget struct {
	ID                string   `json:"id"`
	Name              string   `json:"name"`
	DBType            string   `json:"db_type"`
	Host              string   `json:"host"`
	Port              int      `json:"port"`
	DatabaseName      string   `json:"database_name"`
	AdminUsername     string   `json:"admin_username"`
	DefaultTTLSeconds int      `json:"default_ttl_seconds"`
	MaxTTLSeconds     int      `json:"max_ttl_seconds"`
	Status            string   `json:"status"`
	AllowedRoles      []string `json:"allowed_roles,omitempty"`
}

type vaultDynamicTargetsListResp struct {
	Targets []vaultDynamicTarget `json:"targets"`
	Total   int                  `json:"total"`
}

func runVaultDynamicGenerate(credID string, ttl int) error {
	client, err := NewAPIClient()
	if err != nil {
		return err
	}
	body := map[string]interface{}{}
	if ttl > 0 {
		body["ttl_seconds"] = ttl
	}
	var out vaultDynamicCred
	path := "/v1/vault/dynamic-credentials/" + credID + "/generate"
	if err := client.Post(path, body, &out); err != nil {
		return fmt.Errorf("could not generate credential: %w", err)
	}
	fmt.Printf("✅ Generated %q on target %q\n", out.Credential.Name, out.Target.Name)
	fmt.Printf("   lease_id: %s\n", out.LeaseID)
	fmt.Printf("   host:     %s:%d\n", out.Host, out.Port)
	fmt.Printf("   database: %s\n", out.Database)
	fmt.Printf("   username: %s\n", out.Username)
	fmt.Printf("   expires:  %s\n", out.ExpiresAt.Format(time.RFC3339))
	fmt.Println("\n🔑 PASSWORD (shown only once):")
	fmt.Println(out.Password)
	return nil
}

func runVaultDynamicTargetsList(asJSON bool) error {
	client, err := NewAPIClient()
	if err != nil {
		return err
	}
	var out vaultDynamicTargetsListResp
	if err := client.Get("/v1/vault/dynamic-secret-targets", &out); err != nil {
		return fmt.Errorf("could not list targets: %w", err)
	}
	if asJSON {
		printJSON(out)
		return nil
	}
	if len(out.Targets) == 0 {
		fmt.Println("No dynamic-secret targets configured.")
		fmt.Println("   → Create one with the dashboard, or POST /v1/vault/dynamic-secret-targets")
		return nil
	}
	tw := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(tw, "ID\tNAME\tDB\tHOST:PORT\tDB_NAME\tSTATUS")
	for _, t := range out.Targets {
		fmt.Fprintf(tw, "%s\t%s\t%s\t%s:%d\t%s\t%s\n", t.ID, t.Name, t.DBType, t.Host, t.Port, t.DatabaseName, t.Status)
	}
	tw.Flush()
	return nil
}

// --- audit runner ----------------------------------------------------------

type vaultAuditEntry struct {
	ID        string                 `json:"id"`
	Action    string                 `json:"action"`
	ActorID   string                 `json:"actor_id"`
	ActorType string                 `json:"actor_type"`
	IPAddress string                 `json:"ip_address,omitempty"`
	Success   bool                   `json:"success"`
	Error     string                 `json:"error_message,omitempty"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt time.Time              `json:"created_at"`
}

type vaultAuditListResp struct {
	Entries []vaultAuditEntry `json:"entries"`
	Total   int64             `json:"total"`
}

func runVaultAuditLog(since string, asJSON bool) error {
	dur, err := time.ParseDuration(since)
	if err != nil {
		return fmt.Errorf("--since: %w", err)
	}
	sinceT := time.Now().Add(-dur)
	client, err := NewAPIClient()
	if err != nil {
		return err
	}
	path := "/v1/vault/audit?limit=200"
	var out vaultAuditListResp
	if err := client.Get(path, &out); err != nil {
		return fmt.Errorf("could not fetch audit log: %w", err)
	}
	filtered := out.Entries[:0]
	for _, e := range out.Entries {
		if e.CreatedAt.After(sinceT) {
			filtered = append(filtered, e)
		}
	}
	if asJSON {
		printJSON(map[string]interface{}{"entries": filtered, "total": len(filtered)})
		return nil
	}
	if len(filtered) == 0 {
		fmt.Printf("No audit entries in the last %s.\n", since)
		return nil
	}
	tw := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(tw, "WHEN\tACTION\tACTOR\tIP\tOK\tERROR")
	for _, e := range filtered {
		ok := "✓"
		if !e.Success {
			ok = "✗"
		}
		fmt.Fprintf(tw, "%s\t%s\t%s\t%s\t%s\t%s\n",
			e.CreatedAt.Format(time.RFC3339), e.Action, e.ActorID, e.IPAddress, ok, e.Error)
	}
	tw.Flush()
	return nil
}

// --- helpers ---------------------------------------------------------------

// resolveVaultSecretID accepts either a UUID or a secret name and returns
// the canonical UUID. Names are looked up via list; UUIDs short-circuit.
func resolveVaultSecretID(identifier string) (string, error) {
	if isUUID(identifier) {
		return identifier, nil
	}
	client, err := NewAPIClient()
	if err != nil {
		return "", err
	}
	var out vaultSecretListResp
	if err := client.Get("/v1/vault/secrets?limit=200", &out); err != nil {
		return "", fmt.Errorf("could not look up secret by name: %w", err)
	}
	for _, s := range out.Secrets {
		if s.Name == identifier {
			return s.ID, nil
		}
	}
	return "", fmt.Errorf("secret not found: %q", identifier)
}

func isUUID(s string) bool {
	if len(s) != 36 {
		return false
	}
	for i, ch := range s {
		switch i {
		case 8, 13, 18, 23:
			if ch != '-' {
				return false
			}
		default:
			if !((ch >= '0' && ch <= '9') || (ch >= 'a' && ch <= 'f') || (ch >= 'A' && ch <= 'F')) {
				return false
			}
		}
	}
	return true
}

func parseExpiryHours(s string) (int, error) {
	d, err := time.ParseDuration(s)
	if err != nil {
		return 0, fmt.Errorf("--expiry: %w", err)
	}
	if d <= 0 {
		return 0, fmt.Errorf("--expiry must be positive")
	}
	hours := int(d.Hours())
	if hours < 1 {
		// round up to the nearest hour
		hours = 1
	}
	if hours > 8760 { // 1 year
		return 0, fmt.Errorf("--expiry cannot exceed 8760h (1 year)")
	}
	return hours, nil
}
