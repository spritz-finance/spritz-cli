package config

import (
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

func Dir() string {
	if d := os.Getenv("SPRITZ_CONFIG_DIR"); d != "" {
		return d
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "spritz")
}

func Init() {
	dir := Dir()
	os.MkdirAll(dir, 0700)

	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(dir)

	viper.SetDefault("output", "csv")
	viper.SetDefault("api_url", "https://platform.spritz.finance")

	viper.SetEnvPrefix("SPRITZ")
	viper.AutomaticEnv()

	viper.ReadInConfig() // ignore error — file may not exist yet
}

func Set(key, value string) error {
	viper.Set(key, value)
	dir := Dir()
	os.MkdirAll(dir, 0700)
	return viper.WriteConfigAs(filepath.Join(dir, "config.yaml"))
}

func Get(key string) string {
	return viper.GetString(key)
}

func APIURL() string {
	if u := os.Getenv("SPRITZ_API_URL"); u != "" {
		return u
	}
	return viper.GetString("api_url")
}
