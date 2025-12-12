// Package config provides configuration management for the application.
package config

import (
	"fmt"
	"sync"

	"github.com/spf13/viper"
)

type Database struct {
	Host     string
	Port     int
	User     string
	Password string
	Name     string
	SSLMode  string
	TimeZone string
}

type App struct {
	Name       string
	Enviroment string
	Version    string
}

type Config struct {
	Database Database `mapstructure:"database"`
	App      App      `mapstructure:"app"`
	Storage  Storage  `mapstructure:"storage"`
}

type Storage struct {
	AttachmentPath string `mapstructure:"attachment_path"`
}

var (
	one            sync.Once
	configInstance *Config
)

func LoadConfig(path string) (*Config, error) {
	errChan := make(chan error, 1)
	one.Do(func() {
		viper.AddConfigPath(path)
		viper.SetConfigName("config")
		viper.SetConfigType("yaml")
		viper.AutomaticEnv()
		viper.SetEnvPrefix("HOLO")

		if err := viper.ReadInConfig(); err != nil {
			errChan <- fmt.Errorf("error reading config file: %w", err)
		}

		if err := viper.Unmarshal(&configInstance); err != nil {
			errChan <- fmt.Errorf("unable to decode into struct: %w", err)
		}
	})

	err := <-errChan

	return configInstance, err
}
