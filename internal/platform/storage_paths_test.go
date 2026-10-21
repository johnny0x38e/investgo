package platform

import (
	"path/filepath"
	"testing"
)

func TestDatabaseAndLogPathsForConfigDirectory(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		name      string
		config    string
		directory string
	}{
		{name: "local fallback", directory: "data"},
		{
			name:      "user config",
			config:    filepath.Join("root", "config"),
			directory: filepath.Join("root", "config", "investgo"),
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			databasePath := DatabasePathForConfigDirectory(tt.config)
			if want := filepath.Join(tt.directory, "investgo.db"); databasePath != want {
				t.Fatalf("database path = %q; want %q", databasePath, want)
			}
			if got, want := LogPathForConfigDirectory(tt.config), filepath.Join(tt.directory, "logs", "app.log"); got != want {
				t.Fatalf("log path = %q; want %q", got, want)
			}
		})
	}
}
