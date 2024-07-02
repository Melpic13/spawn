package config

import (
	"fmt"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
)

// Watch invokes callback when the target config file changes.
func Watch(path string, onChange func(*DaemonConfig)) error {
	v := viper.New()
	v.SetConfigFile(path)
	if err := v.ReadInConfig(); err != nil {
		return fmt.Errorf("watch config: %w", err)
	}

	v.OnConfigChange(func(_ fsnotify.Event) {
		var cfg DaemonConfig
		if err := v.Unmarshal(&cfg); err == nil {
			onChange(&cfg)
		}
	})
	v.WatchConfig()
	return nil
}
