package config

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

// PersonaConfig represents a persona (real person) with multiple usernames
type PersonaConfig struct {
	DisplayName string              `mapstructure:"displayName"`
	Image       string              `mapstructure:"image"`     // custom image URL for the persona
	Usernames   map[string][]string `mapstructure:"usernames"` // username -> []address
}

// Config represents the application configuration
type Config struct {
	Server   ServerConfig             `mapstructure:"server"`
	Database DatabaseConfig           `mapstructure:"database"`
	Users    map[string][]string      `mapstructure:"users"`    // username -> []address (legacy)
	Personas map[string]PersonaConfig `mapstructure:"personas"` // slug -> PersonaConfig
	Sync     SyncConfig               `mapstructure:"sync"`
}

// ServerConfig contains HTTP server configuration
type ServerConfig struct {
	Host string `mapstructure:"host"`
	Port int    `mapstructure:"port"`
}

// DatabaseConfig contains database configuration
type DatabaseConfig struct {
	Path string `mapstructure:"path"`
}

// SyncConfig contains sync service configuration
type SyncConfig struct {
	IntervalMinutes int `mapstructure:"intervalMinutes"`
}

// Load loads configuration from a file
func Load(configPath string) (*Config, error) {
	v := viper.New()

	// Set defaults
	v.SetDefault("server.host", "0.0.0.0")
	v.SetDefault("server.port", 8080)
	v.SetDefault("database.path", "./data/pyre.db")
	v.SetDefault("sync.intervalMinutes", 5)

	// Set config file path
	if configPath != "" {
		v.SetConfigFile(configPath)
	} else {
		v.SetConfigName("config")
		v.SetConfigType("yaml")
		v.AddConfigPath(".")
		v.AddConfigPath("./config")
	}

	// Enable environment variables
	v.SetEnvPrefix("BLACKHOLE")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	// Read config file
	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			return nil, fmt.Errorf("config file not found: %w", err)
		}
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// Unmarshal config
	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// Validate config
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	return &cfg, nil
}

// Validate validates the configuration
func (c *Config) Validate() error {
	if c.Server.Port <= 0 || c.Server.Port > 65535 {
		return fmt.Errorf("invalid server port: %d", c.Server.Port)
	}

	if c.Database.Path == "" {
		return fmt.Errorf("database path is required")
	}

	if c.Sync.IntervalMinutes <= 0 {
		return fmt.Errorf("sync interval must be positive, got: %d", c.Sync.IntervalMinutes)
	}

	// Need either users or personas configured
	if len(c.Users) == 0 && len(c.Personas) == 0 {
		return fmt.Errorf("at least one user or persona must be configured")
	}

	// Validate legacy users
	for username, addresses := range c.Users {
		if username == "" {
			return fmt.Errorf("empty username is not allowed")
		}
		if len(addresses) == 0 {
			return fmt.Errorf("user %s has no addresses configured", username)
		}
		for i, addr := range addresses {
			if addr == "" {
				return fmt.Errorf("user %s has empty address at index %d", username, i)
			}
		}
	}

	// Validate personas
	for slug, persona := range c.Personas {
		if slug == "" {
			return fmt.Errorf("empty persona slug is not allowed")
		}
		if persona.DisplayName == "" {
			return fmt.Errorf("persona %s has no display name", slug)
		}
		if len(persona.Usernames) == 0 {
			return fmt.Errorf("persona %s has no usernames configured", slug)
		}
		for username, addresses := range persona.Usernames {
			if username == "" {
				return fmt.Errorf("persona %s has empty username", slug)
			}
			if len(addresses) == 0 {
				return fmt.Errorf("persona %s user %s has no addresses configured", slug, username)
			}
			for i, addr := range addresses {
				if addr == "" {
					return fmt.Errorf("persona %s user %s has empty address at index %d", slug, username, i)
				}
			}
		}
	}

	return nil
}

// GetAllUsers returns all users from both legacy users config and personas
// Returns a map of username -> addresses
func (c *Config) GetAllUsers() map[string][]string {
	allUsers := make(map[string][]string, len(c.Users))

	// Add legacy users
	for username, addresses := range c.Users {
		allUsers[username] = addresses
	}

	// Add users from personas
	for _, persona := range c.Personas {
		for username, addresses := range persona.Usernames {
			allUsers[username] = addresses
		}
	}

	return allUsers
}
