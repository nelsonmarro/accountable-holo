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

type Storage struct {
	AttachmentPath string `mapstructure:"attachment_path"`
}

type Email struct {
	Provider string `mapstructure:"provider"` // "resend" or "smtp" (legacy)
	APIKey   string `mapstructure:"api_key"`
	From     string `mapstructure:"from"`
	// Legacy SMTP fields (optional)
	Host string `mapstructure:"host"`
	Port int    `mapstructure:"port"`
	User string `mapstructure:"user"`
}

type Config struct {
	Database Database `mapstructure:"database"`
	App      App      `mapstructure:"app"`
	Storage  Storage  `mapstructure:"storage"`
	Email    Email    `mapstructure:"email"`
}

var (
	one            sync.Once
	configInstance *Config
	configErr      error
)

func LoadConfig(path string) (*Config, error) {
	one.Do(func() {
		viper.AddConfigPath(path)
		viper.SetConfigName("config")
		viper.SetConfigType("yaml")
		viper.AutomaticEnv()
		viper.SetEnvPrefix("VERITH")

		if err := viper.ReadInConfig(); err != nil {
			configErr = fmt.Errorf("error reading config file: %w", err)
			return
		}

		if err := viper.Unmarshal(&configInstance); err != nil {
			configErr = fmt.Errorf("unable to decode into struct: %w", err)
			return
		}
	})

	return configInstance, configErr
}
