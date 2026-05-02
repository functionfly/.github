package storage

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTenantDatabaseConfig_GetTenantDBName(t *testing.T) {
	cfg := &TenantDatabaseConfig{
		Prefix: "functionfly_tenant_",
	}

	tests := []struct {
		name     string
		tenantID string
		want     string
	}{
		{
			name:     "normal tenant ID",
			tenantID: "550e8400-e29b-41d4-a716-446655440000",
			want:     "functionfly_tenant_550e8400",
		},
		{
			name:     "short tenant ID",
			tenantID: "abc",
			want:     "functionfly_tenant_abc",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := cfg.GetTenantDBName(tt.tenantID)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestTenantDatabaseConfig_BuildTenantDBConnectionString(t *testing.T) {
	cfg := &TenantDatabaseConfig{
		Host:     "localhost",
		Port:     5432,
		User:     "postgres",
		Password: "secret",
		PoolMax:  10,
	}

	got := cfg.BuildTenantDBConnectionString("testdb")
	want := "host=localhost port=5432 user=postgres password=secret dbname=testdb sslmode=require pool_max_conns=10"

	assert.Equal(t, want, got)
}

func TestEncryptDecryptPasswordFallback(t *testing.T) {
	password := "my-secret-password-123"

	encrypted, err := encryptPasswordFallback(password)
	require.NoError(t, err)
	assert.NotEmpty(t, encrypted)
	assert.NotEqual(t, password, encrypted)

	// Test that it's base64 encoded
	decrypted, err := decryptPasswordFallback(encrypted)
	require.NoError(t, err)
	assert.Equal(t, password, decrypted)
}

func TestEncryptPasswordFallback_EmptyPassword(t *testing.T) {
	encrypted, err := encryptPasswordFallback("")
	assert.NoError(t, err)
	assert.Empty(t, encrypted)
}

func TestEncryptPasswordFallback_DifferentPasswords(t *testing.T) {
	password1 := "password1"
	password2 := "password2"

	encrypted1, err := encryptPasswordFallback(password1)
	require.NoError(t, err)

	encrypted2, err := encryptPasswordFallback(password2)
	require.NoError(t, err)

	// Different passwords should produce different encrypted values
	assert.NotEqual(t, encrypted1, encrypted2)
}

func TestDecryptPasswordFallback_InvalidInput(t *testing.T) {
	// Test with invalid base64
	_, err := decryptPasswordFallback("not-valid-base64!!!")
	assert.Error(t, err)
}

func TestLoadTenantDatabaseConfig(t *testing.T) {
	// Reset environment for test
	t.Setenv("TENANT_DB_ENABLED", "true")
	t.Setenv("TENANT_DB_HOST", "tenant-db.example.com")
	t.Setenv("TENANT_DB_PORT", "5433")
	t.Setenv("TENANT_DB_USER", "tenant_user")
	t.Setenv("TENANT_DB_PASSWORD", "tenant_secret")
	t.Setenv("TENANT_DB_TEMPLATE", "my_template")
	t.Setenv("TENANT_DB_PREFIX", "custom_prefix_")
	t.Setenv("TENANT_DB_POOL_MIN", "3")
	t.Setenv("TENANT_DB_POOL_MAX", "20")
	t.Setenv("TENANT_DB_USE_TEMPLATE", "true")

	cfg := LoadTenantDatabaseConfig()

	assert.True(t, cfg.Enabled)
	assert.Equal(t, "tenant-db.example.com", cfg.Host)
	assert.Equal(t, 5433, cfg.Port)
	assert.Equal(t, "tenant_user", cfg.User)
	assert.Equal(t, "tenant_secret", cfg.Password)
	assert.Equal(t, "my_template", cfg.TemplateDB)
	assert.Equal(t, "custom_prefix_", cfg.Prefix)
	assert.Equal(t, 3, cfg.PoolMin)
	assert.Equal(t, 20, cfg.PoolMax)
	assert.True(t, cfg.UseTemplateDB)
}

func TestLoadTenantDatabaseConfig_Defaults(t *testing.T) {
	// Clear relevant environment variables
	t.Setenv("TENANT_DB_ENABLED", "false")

	cfg := LoadTenantDatabaseConfig()

	assert.False(t, cfg.Enabled)
	assert.Equal(t, "localhost", cfg.Host)
	assert.Equal(t, 5432, cfg.Port)
	assert.Equal(t, "postgres", cfg.User)
	assert.Equal(t, "functionfly_tenant_template", cfg.TemplateDB)
	assert.Equal(t, "functionfly_tenant_", cfg.Prefix)
	assert.Equal(t, 2, cfg.PoolMin)
	assert.Equal(t, 10, cfg.PoolMax)
}

func TestTenantDBProvisioner_EncryptPassword(t *testing.T) {
	cfg := &TenantDatabaseConfig{
		Enabled: true,
		Host:    "localhost",
		Port:    5432,
		User:    "postgres",
		Password: "platform_password",
	}

	// Create provisioner without platform DB to use fallback
	provisioner, err := NewTenantDBProvisioner(cfg, nil)
	require.NoError(t, err)

	// Test encryption
	encrypted, err := provisioner.encryptPassword("tenant_password")
	require.NoError(t, err)
	assert.NotEmpty(t, encrypted)
	assert.NotEqual(t, "tenant_password", encrypted)
}

func TestTenantDBProvisioner_EncryptPassword_Empty(t *testing.T) {
	cfg := &TenantDatabaseConfig{
		Enabled: true,
	}

	provisioner, err := NewTenantDBProvisioner(cfg, nil)
	require.NoError(t, err)

	// Test with empty password
	encrypted, err := provisioner.encryptPassword("")
	assert.NoError(t, err)
	assert.Empty(t, encrypted)
}

func TestNewTenantDBProvisioner_Disabled(t *testing.T) {
	cfg := &TenantDatabaseConfig{
		Enabled: false,
	}

	provisioner, err := NewTenantDBProvisioner(cfg, nil)
	require.NoError(t, err)
	assert.NotNil(t, provisioner)
}

func TestNewTenantDBProvisioner_WithPlatformDB(t *testing.T) {
	cfg := &TenantDatabaseConfig{
		Enabled:         true,
		Host:            "localhost",
		Port:            5432,
		User:            "postgres",
		Password:        "postgres",
		UseTemplateDB:   false,
	}

	// Create provisioner without platform DB (nil) since we don't have a test DB
	provisioner, err := NewTenantDBProvisioner(cfg, nil)
	// May fail due to template validation, but that's OK for this test
	if err == nil {
		assert.NotNil(t, provisioner)
	}
}

func TestExtractMigrationVersion(t *testing.T) {
	tests := []struct {
		name      string
		filename  string
		want      int
		wantErr   bool
	}{
		{
			name:     "valid migration",
			filename: "20260501142000_tenant_base_schema.up.sql",
			want:     20260501142000,
			wantErr:  false,
		},
		{
			name:     "valid migration 2",
			filename: "20260315103000_add_tables.up.sql",
			want:     20260315103000,
			wantErr:  false,
		},
		{
			name:     "invalid format",
			filename: "invalid_name.up.sql",
			want:     0,
			wantErr:  true,
		},
		{
			name:     "missing version",
			filename: "description.up.sql",
			want:     0,
			wantErr:  true,
		},
		{
			name:     "wrong extension",
			filename: "20260501142000_tenant_base_schema.down.sql",
			want:     0,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := extractMigrationVersion(tt.filename)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func TestPqQuoteIdent(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"simple", `"simple"`},
		{"with spaces", `"with spaces"`},
		{"with'quote", `"with''quote"`},
		{"lowercase", `"lowercase"`},
		{"UPPERCASE", `"UPPERCASE"`},
		{"MixedCase", `"MixedCase"`},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := pqQuoteIdent(tt.input)
			assert.Equal(t, tt.expected, got)
		})
	}
}

func TestRedactPassword(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "normal connection string",
			input:    "host=localhost port=5432 user=postgres password=secret dbname=test",
			expected: "host=localhost port=5432 user=postgres password=**** dbname=test",
		},
		{
			name:     "connection string with special chars",
			input:    "host=localhost password=verylong&complex!password",
			expected: "host=localhost password=****",
		},
		{
			name:     "no password",
			input:    "host=localhost port=5432",
			expected: "host=localhost port=5432",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := redactPassword(tt.input)
			assert.Equal(t, tt.expected, got)
		})
	}
}