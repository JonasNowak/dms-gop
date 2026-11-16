package config

import (
	"log"
	"os"
	"path/filepath"

	"github.com/caarlos0/env/v11"
	"github.com/duxlabs-dev/dux-go/paths"
)

var appPaths = paths.New("dms-gop")

// DefaultDMSThemePath returns the path to the DMS theme JSON file.
// If DMS_THEME_JSON environment variable is set, it returns that path.
// Otherwise, it returns the default path in Library/Application Support/DMS/theme.json.
func DefaultDMSThemePath() string {
	if envPath := os.Getenv("DMS_THEME_JSON"); envPath != "" {
		return envPath
	}
	// Get home directory
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, "Library", "Application Support", "DMS", "theme.json")
}

type Config struct {
	ApiPort string `env:"API_PORT" envDefault:":63484"` // Default port for the API server
}

// Parse environment variables into a Config struct
func NewConfig() *Config {
	cfg := Config{}
	if err := env.Parse(&cfg); err != nil {
		log.Fatal("Error parsing environment", "err", err)
	}

	return &cfg
}
