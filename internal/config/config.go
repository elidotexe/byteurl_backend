package config

import (
	"github.com/spf13/viper"
)

type AppConfig struct {
	DSN           string
	PORT          string `mapstructure:"PORT"`
	COOKIE_DOMAIN string `mapstructure:"COOKIE_DOMAIN"`
	DOMAIN        string `mapstructure:"DOMAIN"`
	DB_HOST       string `mapstructure:"DB_HOST"`
	DB_PORT       string `mapstructure:"DB_PORT"`
	DB_NAME       string `mapstructure:"DB_NAME"`
	DB_USER       string `mapstructure:"DB_USER"`
	DB_PASSWORD   string `mapstructure:"DB_PASSWORD"`
	JWT_SECRET    string `mapstructure:"JWT_SECRET"`
	JWT_ISSUER    string `mapstructure:"JWT_ISSUER"`
	JWT_AUDIENCE  string `mapstructure:"JWT_AUDIENCE"`
}

func LoadConfig() (config *AppConfig, err error) {
	viper.AddConfigPath(".")
	viper.SetConfigName(".env")
	viper.SetConfigType("env")

	viper.AutomaticEnv()

	err = viper.ReadInConfig()
	if err != nil {
		return nil, err
	}

	config = new(AppConfig)
	err = viper.Unmarshal(&config)
	if err != nil {
		return nil, err
	}

	return config, nil
}
