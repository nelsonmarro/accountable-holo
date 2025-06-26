// Package config provides configuration management for the application.
package config

import (
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
	Database Database
	App      App
}

var (
	one            sync.Once
	configInstance *Config
)

func LoadConfig(path string) *Config {
	one.Do(func() {
		viper.AddConfigPath(path)
		viper.SetConfigName("config")
		viper.SetConfigType("yaml")
		viper.AutomaticEnv()
		viper.SetEnvPrefix("HOLO")

		if err := viper.ReadInConfig(); err != nil {
			panic("Error reading config file: " + err.Error())
		}

		if err := viper.Unmarshal(&configInstance); err != nil {
			panic("Error unmarshalling config: " + err.Error())
		}
	})

	return configInstance
}
