package platform

import "path/filepath"

const appDataDirectory = "investgo"

func storageDirectoryForConfigDirectory(configDirectory string) string {
	if configDirectory == "" {
		return "data"
	}
	return filepath.Join(configDirectory, appDataDirectory)
}

// DatabasePathForConfigDirectory returns the SQLite database path.
// An empty config directory uses the local data directory.
func DatabasePathForConfigDirectory(configDirectory string) string {
	return filepath.Join(storageDirectoryForConfigDirectory(configDirectory), "investgo.db")
}

// LogPathForConfigDirectory returns the log path alongside application data.
func LogPathForConfigDirectory(configDirectory string) string {
	return filepath.Join(storageDirectoryForConfigDirectory(configDirectory), "logs", "app.log")
}
