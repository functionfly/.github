package registry

import (
	"testing"
)

func TestMCPSettingsInput_Validate(t *testing.T) {
	cases := []struct {
		name    string
		in      MCPSettingsInput
		wantErr bool
	}{
		{
			name:    "empty",
			in:      MCPSettingsInput{},
			wantErr: false,
		},
		{
			name: "valid override",
			in: MCPSettingsInput{
				ToolNameOverride: "my-alias_v2",
			},
			wantErr: false,
		},
		{
			name: "override too long",
			in: MCPSettingsInput{
				ToolNameOverride: string(make([]byte, 65)),
			},
			wantErr: true,
		},
		{
			name: "override with bad chars",
			in: MCPSettingsInput{
				ToolNameOverride: "bad name!",
			},
			wantErr: true,
		},
		{
			name: "rate limit ok",
			in: MCPSettingsInput{
				RateLimitPerMin: 100,
			},
			wantErr: false,
		},
		{
			name: "rate limit too high",
			in: MCPSettingsInput{
				RateLimitPerMin: 100000,
			},
			wantErr: true,
		},
		{
			name: "rate limit negative",
			in: MCPSettingsInput{
				RateLimitPerMin: -1,
			},
			wantErr: true,
		},
		{
			name: "valid transports",
			in: MCPSettingsInput{
				Transports: []string{"streamable-http"},
			},
			wantErr: false,
		},
		{
			name: "invalid transport",
			in: MCPSettingsInput{
				Transports: []string{"websocket"},
			},
			wantErr: true,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.in.Validate()
			if tc.wantErr && err == nil {
				t.Errorf("expected error, got nil")
			}
			if !tc.wantErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestMCPSettingsInput_ApplyDefaults(t *testing.T) {
	in := MCPSettingsInput{}
	in.ApplyDefaults()
	if len(in.Transports) == 0 {
		t.Error("transports not defaulted")
	}
	if in.Transports[0] != "streamable-http" {
		t.Errorf("default transport = %q, want streamable-http", in.Transports[0])
	}
	if in.RateLimitPerMin != 60 {
		t.Errorf("default rate limit = %d, want 60", in.RateLimitPerMin)
	}
	if in.AllowlistOrigins == nil {
		t.Error("allowlist not defaulted to empty slice")
	}
}
