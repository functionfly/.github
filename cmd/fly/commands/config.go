package commands

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type GlobalConfig struct {
	Version   string    `yaml:"version" json:"version"`
	API       APIConfig `yaml:"api"     json:"api"`
	Dev       DevConfig `yaml:"dev"     json:"dev"`
	Publish   PubConfig `yaml:"publish" json:"publish"`
	Telemetry TelConfig `yaml:"telemetry" json:"telemetry"`
}

type APIConfig struct {
	URL     string `yaml:"url"     json:"url"`
	Timeout string `yaml:"timeout" json:"timeout"`
}

type DevConfig struct {
	Port      int  `yaml:"port"       json:"port"`
	Watch     bool `yaml:"watch"      json:"watch"`
	HotReload bool `yaml:"hot_reload" json:"hot_reload"`
}

type PubConfig struct {
	Confirm    bool `yaml:"confirm"     json:"confirm"`
	AutoUpdate bool `yaml:"auto_update" json:"auto_update"`
}

type TelConfig struct {
	Enabled   bool `yaml:"enabled"   json:"enabled"`
	Anonymize bool `yaml:"anonymize" json:"anonymize"`
}

func DefaultConfig() *GlobalConfig {
	return &GlobalConfig{
		Version: "1.0.0",
		API:     APIConfig{URL: "https://api.functionfly.com", Timeout: "30s"},
		Dev:     DevConfig{Port: 8787, Watch: true, HotReload: true},
		Publish: PubConfig{Confirm: false, AutoUpdate: true},
		Telemetry: TelConfig{Enabled: true, Anonymize: false},
	}
}

func configPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("could not determine home directory: %w", err)
	}
	return filepath.Join(home, ".functionfly", "config.yaml"), nil
}

func LoadConfig() (*GlobalConfig, error) {
	path, err := configPath()
	if err != nil {
		return DefaultConfig(), nil
	}
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return DefaultConfig(), nil
		}
		return nil, fmt.Errorf("could not read config: %w", err)
	}
	cfg := DefaultConfig()
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("could not parse config: %w", err)
	}
	return cfg, nil
}

func SaveConfig(cfg *GlobalConfig) error {
	path, err := configPath()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return fmt.Errorf("could not create config directory: %w", err)
	}
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("could not serialize config: %w", err)
	}
	return os.WriteFile(path, data, 0644)
}

type Manifest struct {
	Schema        string            `json:"$schema,omitempty"`
	Name          string            `json:"name"`
	Version       string            `json:"version"`
	Runtime       string            `json:"runtime"`
	Public        bool              `json:"public"`
	Deterministic bool              `json:"deterministic"`
	CacheTTL      int               `json:"cache_ttl,omitempty"`
	TimeoutMS     int               `json:"timeout_ms,omitempty"`
	MemoryMB      int               `json:"memory_mb,omitempty"`
	Description   string            `json:"description,omitempty"`
	Dependencies  map[string]string `json:"dependencies,omitempty"`
	Env           map[string]string `json:"env,omitempty"`
}

func LoadManifest(dir string) (*Manifest, error) {
	if dir == "" {
		var err error
		dir, err = os.Getwd()
		if err != nil {
			return nil, err
		}
	}
	candidates := []string{
		filepath.Join(dir, "functionfly.jsonc"),
		filepath.Join(dir, "functionfly.json"),
	}
	for _, path := range candidates {
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		data = stripJSONCComments(data)
		var m Manifest
		if err := json.Unmarshal(data, &m); err != nil {
			return nil, fmt.Errorf("invalid manifest at %s: %w", path, err)
		}
		return &m, nil
	}
	return nil, fmt.Errorf("no functionfly.jsonc found in %s\n   → Run: fly init <name>", dir)
}

func SaveManifest(dir string, m *Manifest) error {
	if dir == "" {
		var err error
		dir, err = os.Getwd()
		if err != nil {
			return err
		}
	}
	data, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, "functionfly.jsonc"), data, 0644)
}

func stripJSONCComments(data []byte) []byte {
	var result []byte
	inString := false
	i := 0
	for i < len(data) {
		c := data[i]
		if c == '"' && (i == 0 || data[i-1] != '\\') {
			inString = !inString
		}
		if !inString && c == '/' && i+1 < len(data) && data[i+1] == '/' {
			for i < len(data) && data[i] != '\n' {
				i++
			}
			continue
		}
		result = append(result, c)
		i++
	}
	return result
}

func APIURL() string {
	cfg, err := LoadConfig()
	if err != nil || cfg.API.URL == "" {
		return "https://api.functionfly.com"
	}
	return cfg.API.URL
}
