package config

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

// Load reads the configuration from the specified file path
// and environment variables.
func Load(path string) (*Config, error) {
	v := viper.New()

	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	// If a path is provided, read the configuration from the file.
	if path != "" {
		v.SetConfigFile(path)
		v.SetConfigType("yaml")
		//
		if err := v.ReadInConfig(); err != nil {
			return nil, fmt.Errorf("failed to read config %q: %w", path, err)
		}
	}

	// Unmarshal the configuration into the Config struct
	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	if err := validate(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

// validate checks the configuration for any required fields or
// invalid values.
func validate(cfg *Config) error {
	if cfg.Service.Name == "" {
		return fmt.Errorf("service.name is required")
	}
	return nil
}
