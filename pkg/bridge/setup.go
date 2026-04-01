package bridge

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
)

// EnsureBridge checks if the bridge script is correctly installed at targetDir.
// It writes the embedded bundleJS to that directory if missing or outdated.
// Returns the absolute path to the written script file.
func EnsureBridge(targetDir string, bundleJS []byte, backendName string) (string, error) {
	if err := os.MkdirAll(targetDir, 0700); err != nil {
		return "", fmt.Errorf("create bridge dir: %w", err)
	}

	bundlePath := filepath.Join(targetDir, backendName+".mjs")

	existing, readErr := os.ReadFile(bundlePath)
	bundleUpToDate := readErr == nil && string(existing) == string(bundleJS)

	if bundleUpToDate {
		return bundlePath, nil
	}

	// Clean up old .js file if we migrated to .mjs
	oldPath := filepath.Join(targetDir, backendName+".js")
	if oldPath != bundlePath {
		_ = os.Remove(oldPath)
	}

	log.Printf("Updating Dev Studio Bridge bundle for backend %q...\n", backendName)

	if err := os.WriteFile(bundlePath, bundleJS, 0600); err != nil {
		return "", fmt.Errorf("write bridge bundle: %w", err)
	}

	return bundlePath, nil
}
