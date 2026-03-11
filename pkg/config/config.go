package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
)

type Config struct {
	Host       string            `json:"host"`
	Apps       map[string]string `json:"apps"`
	DefaultApp string            `json:"default_app,omitempty"`
}

func configDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("cannot determine home directory: %w", err)
	}
	return filepath.Join(home, ".config", "dify-cli"), nil
}

func configPath() (string, error) {
	dir, err := configDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "config.json"), nil
}

func Load() (*Config, error) {
	p, err := configPath()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(p)
	if err != nil {
		if os.IsNotExist(err) {
			return &Config{Apps: make(map[string]string)}, nil
		}
		return nil, fmt.Errorf("failed to read config: %w", err)
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}
	if cfg.Apps == nil {
		cfg.Apps = make(map[string]string)
	}
	return &cfg, nil
}

func (c *Config) Save() error {
	dir, err := configDir()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(dir, 0o750); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	p := filepath.Join(dir, "config.json")
	if err := os.WriteFile(p, data, 0o600); err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}
	return nil
}

func (c *Config) ValidateHost() error {
	if c.Host == "" {
		return fmt.Errorf("host not configured. Run: dify config set-host <host>")
	}
	return nil
}

// ResolveAPIKey resolves the API key from direct key, app name, or default app.
// Priority: directKey > appName > defaultApp
func (c *Config) ResolveAPIKey(directKey, appName string) (string, error) {
	if directKey != "" {
		return directKey, nil
	}

	name := appName
	if name == "" {
		name = c.DefaultApp
	}
	if name == "" {
		return "", fmt.Errorf(
			"no app specified. Use -a <app_name>, -k <api_key>, or set a default: dify app default <name>",
		)
	}

	key, ok := c.Apps[name]
	if !ok {
		return "", fmt.Errorf("app %q not found. Run: dify app add <name> <api_key>", name)
	}
	return key, nil
}

func (c *Config) ListApps() []string {
	names := make([]string, 0, len(c.Apps))
	for name := range c.Apps {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

func MaskKey(key string) string {
	if len(key) <= 8 {
		return "****"
	}
	return key[:4] + "****" + key[len(key)-4:]
}
