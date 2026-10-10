package platform

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestLogPathForConfigDirectory(t *testing.T) {
	t.Parallel()
	for _, config := range []string{"", filepath.Join("root", "config")} {
		jsonPath, _ := StoragePathsForConfigDirectory(config)
		want := filepath.Join(filepath.Dir(jsonPath), "logs", "app.log")
		if got := LogPathForConfigDirectory(config); got != want {
			t.Fatalf("log path = %q; want %q", got, want)
		}
	}
}

func TestMigrateStorageDirectoryPreservesFiles(t *testing.T) {
	t.Parallel()
	config := t.TempDir()
	source := filepath.Join(config, "com.johnny0x38e.investgo")
	if err := os.MkdirAll(filepath.Join(source, "logs"), 0700); err != nil {
		t.Fatal(err)
	}
	files := []string{
		"investgo.db",
		"investgo.db-wal",
		"investgo.db-shm",
		"state.json",
		"state.json.bak",
		filepath.Join("logs", "app.log"),
	}
	for _, name := range files {
		if err := os.WriteFile(filepath.Join(source, name), []byte(name), 0600); err != nil {
			t.Fatal(err)
		}
	}
	migrated, err := MigrateStorageDirectory(config)
	if err != nil || !migrated {
		t.Fatalf("migration = %v, %v; want true, nil", migrated, err)
	}
	if _, err := os.Stat(source); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("legacy directory still exists: %v", err)
	}
	for _, name := range files {
		data, err := os.ReadFile(filepath.Join(config, "investgo", name))
		if err != nil || string(data) != name {
			t.Fatalf("migrated %s = %q, %v", name, data, err)
		}
	}
	if migrated, err := MigrateStorageDirectory(config); err != nil || migrated {
		t.Fatalf("repeated migration = %v, %v; want false, nil", migrated, err)
	}
}

func TestMigrateStorageDirectoryKeepsExistingDestination(t *testing.T) {
	t.Parallel()
	config := t.TempDir()
	for _, directory := range []string{"investgo", "com.johnny0x38e.investgo"} {
		path := filepath.Join(config, directory)
		if err := os.Mkdir(path, 0700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(path, "investgo.db"), []byte(directory), 0600); err != nil {
			t.Fatal(err)
		}
	}
	if migrated, err := MigrateStorageDirectory(config); err != nil || migrated {
		t.Fatalf("migration = %v, %v; want false, nil", migrated, err)
	}
	for _, directory := range []string{"investgo", "com.johnny0x38e.investgo"} {
		data, err := os.ReadFile(filepath.Join(config, directory, "investgo.db"))
		if err != nil || string(data) != directory {
			t.Fatalf("%s database changed: %q, %v", directory, data, err)
		}
	}
}

func TestMigrateStorageDirectoryWithoutLegacyData(t *testing.T) {
	t.Parallel()
	for _, config := range []string{"", t.TempDir()} {
		if migrated, err := MigrateStorageDirectory(config); err != nil || migrated {
			t.Fatalf("migration = %v, %v; want false, nil", migrated, err)
		}
		if config != "" {
			if _, err := os.Stat(filepath.Join(config, "investgo")); !errors.Is(err, os.ErrNotExist) {
				t.Fatalf("migration created a destination: %v", err)
			}
		}
	}
}

func TestMigrateStorageDirectoryRejectsInvalidPaths(t *testing.T) {
	t.Parallel()
	for _, name := range []string{"investgo", "com.johnny0x38e.investgo"} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			config := t.TempDir()
			path := filepath.Join(config, name)
			if err := os.WriteFile(path, []byte("keep"), 0600); err != nil {
				t.Fatal(err)
			}
			if migrated, err := MigrateStorageDirectory(config); err == nil || migrated {
				t.Fatalf("migration = %v, %v; want false, error", migrated, err)
			}
			data, err := os.ReadFile(path)
			if err != nil || string(data) != "keep" {
				t.Fatalf("invalid path changed: %q, %v", data, err)
			}
		})
	}
}

func TestMigrateStorageDirectoryRejectsSymlinks(t *testing.T) {
	t.Parallel()
	for _, name := range []string{"investgo", "com.johnny0x38e.investgo"} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			config := t.TempDir()
			target := t.TempDir()
			if err := os.Symlink(target, filepath.Join(config, name)); err != nil {
				t.Skipf("symlinks unavailable: %v", err)
			}
			if migrated, err := MigrateStorageDirectory(config); err == nil || migrated {
				t.Fatalf("migration = %v, %v; want false, error", migrated, err)
			}
		})
	}
}
