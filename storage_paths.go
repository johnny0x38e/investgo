package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

const appIdentifier = "com.johnny0x38e.investgo"

func storageDirectoryForConfigDirectory(configDirectory string) string {
	if configDirectory == "" {
		return "data"
	}
	return filepath.Join(configDirectory, appIdentifier)
}

func storagePathsForConfigDirectory(configDirectory string) (jsonPath, databasePath string) {
	directory := storageDirectoryForConfigDirectory(configDirectory)
	return filepath.Join(directory, "state.json"), filepath.Join(directory, "investgo.db")
}

func logPathForConfigDirectory(configDirectory string) string {
	return filepath.Join(storageDirectoryForConfigDirectory(configDirectory), "logs", "app.log")
}

// migrateStorageDirectory runs before logging or SQLite can create the new
// directory. Renaming the whole directory keeps SQLite sidecars, legacy JSON
// backups, and logs together. Existing destinations are never merged.
func migrateStorageDirectory(configDirectory string) (bool, error) {
	if configDirectory == "" {
		return false, nil
	}
	destination := storageDirectoryForConfigDirectory(configDirectory)
	if info, err := os.Lstat(destination); err == nil {
		if !info.IsDir() {
			return false, fmt.Errorf("application data path is not a directory: %s", destination)
		}
		return false, nil
	} else if !errors.Is(err, os.ErrNotExist) {
		return false, fmt.Errorf("inspect application data directory: %w", err)
	}

	source := filepath.Join(configDirectory, "investgo")
	info, err := os.Lstat(source)
	if errors.Is(err, os.ErrNotExist) {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("inspect legacy data directory: %w", err)
	}
	if !info.IsDir() {
		return false, fmt.Errorf("legacy data path is not a directory: %s", source)
	}
	if err := os.Rename(source, destination); err != nil {
		return false, fmt.Errorf("rename legacy data directory: %w", err)
	}
	return true, nil
}
