package migrate

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/sipeed/kakoclaw/pkg/config"
	"github.com/sipeed/kakoclaw/pkg/storage"
)

type MultiuserResult struct {
	WorkspaceMoved bool
	ConfigCopied   bool
	DataBackfilled bool
	Warnings       []string
}

// MigrateToMultiuser moves the legacy single-user workspace/config into a user-scoped structure
// and backfills user_id values in the database.
func MigrateToMultiuser(cfg *config.Config, store *storage.Storage, user *storage.User) (*MultiuserResult, error) {
	if cfg == nil || store == nil || user == nil {
		return nil, fmt.Errorf("missing inputs for migration")
	}
	if user.UUID == "" {
		return nil, fmt.Errorf("user UUID is required")
	}

	result := &MultiuserResult{}

	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	userRoot := filepath.Join(home, ".kakoclaw", "users", user.UUID)
	userWorkspace := filepath.Join(userRoot, "workspace")

	// Move legacy workspace if present and user workspace does not exist yet
	legacyWorkspace := cfg.WorkspacePath()
	if legacyWorkspace != "" {
		if _, err := os.Stat(legacyWorkspace); err == nil {
			if _, err := os.Stat(userWorkspace); os.IsNotExist(err) {
				if err := os.Rename(legacyWorkspace, userWorkspace); err != nil {
					result.Warnings = append(result.Warnings, fmt.Sprintf("failed to move workspace: %v", err))
				} else {
					result.WorkspaceMoved = true
				}
			}
		}
	}

	// Ensure user workspace exists and bootstrap files are present
	if _, err := config.EnsureUserWorkspace(user.UUID); err != nil {
		return nil, err
	}

	// Copy global config into user config if it does not exist
	userConfigPath := filepath.Join(userRoot, "config.json")
	globalConfigPath := filepath.Join(home, ".kakoclaw", "config.json")
	if _, err := os.Stat(userConfigPath); os.IsNotExist(err) {
		if data, err := os.ReadFile(globalConfigPath); err == nil {
			if err := os.WriteFile(userConfigPath, data, 0644); err == nil {
				result.ConfigCopied = true
			} else {
				result.Warnings = append(result.Warnings, fmt.Sprintf("failed to copy config: %v", err))
			}
		}
	}

	// Backfill user_id columns in DB
	if err := store.BackfillUserID(user.ID); err != nil {
		return nil, err
	}
	result.DataBackfilled = true

	// Record migration setting
	_ = store.SetSetting("multiuser_migrated", user.UUID)

	return result, nil
}

func isDirEmpty(path string) bool {
	entries, err := os.ReadDir(path)
	if err != nil {
		return false
	}
	return len(entries) == 0
}
