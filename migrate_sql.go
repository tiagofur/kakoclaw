package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	_ "modernc.org/sqlite"
)

func main() {
	home, _ := os.UserHomeDir()
	
	// Check central db
	centralDB := filepath.Join(home, ".MakoClaw", "makoclaw.db")
	migrateDB(centralDB)
	// Check user dbs
	usersDirs := []string{
		filepath.Join(home, ".MakoClaw", "users"),
		filepath.Join(home, ".MakoClaw", "workspace", "users"),
	}
	for _, usersDir := range usersDirs {
		entries, _ := os.ReadDir(usersDir)
		for _, entry := range entries {
			if entry.IsDir() {
				userDB := filepath.Join(usersDir, entry.Name(), "makoclaw.db")
				migrateDB(userDB)
			}
		}
	}
}

func migrateDB(path string) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return
	}
	
	db, err := sql.Open("sqlite", path)
	if err != nil {
		log.Printf("Failed to open %s: %v", path, err)
		return
	}
	defer db.Close()
	
	query := "ALTER TABLE tasks ADD COLUMN agent TEXT NOT NULL DEFAULT '';"
	_, err = db.Exec(query)
	if err != nil {
		if !strings.Contains(err.Error(), "duplicate column name") {
			log.Printf("Error migrating %s: %v", path, err)
		} else {
			log.Printf("Migration already applied for %s", path)
		}
	} else {
		fmt.Printf("Successfully migrated %s\n", path)
	}
}
