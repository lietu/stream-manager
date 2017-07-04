package config

import (
	"os"
	"path/filepath"
)

func GetLocalStoragePath() string {
	appData := os.Getenv("APPDATA")
	location := filepath.Join(appData, "stream-manager")
	return location
}
