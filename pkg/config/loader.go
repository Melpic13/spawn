package config

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

// Load reads daemon config from file.
func Load(path string) (*DaemonConfig, error) {
	v := viper.New()
	v.SetConfigFile(path)
	v.SetConfigType(configType(path))
	v.SetEnvPrefix("SPAWN")
	v.AutomaticEnv()

	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}

	var cfg DaemonConfig
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("unmarshal config: %w", err)
	}

	if err := Validate(&cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

func configType(path string) string {
	if strings.HasSuffix(path, ".json") {
		return "json"
	}
	return "yaml"
}
