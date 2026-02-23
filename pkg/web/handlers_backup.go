package web

import (
	"archive/zip"
	"compress/flate"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/sipeed/makoclaw/pkg/config"
	"github.com/sipeed/makoclaw/pkg/logger"
	"github.com/sipeed/makoclaw/pkg/storage"
)

// BackupManifest represents the metadata of a backup archive.
// Version "2.0" uses JSON data export instead of raw DB files.
type BackupManifest struct {
	Version            string    `json:"version"`
	CreatedAt          time.Time `json:"created_at"`
	MakoClawVersion    string    `json:"makoclaw_version"`
	BackupType         string    `json:"backup_type"` // "personal" — scoped to the exporting user
	DataSizeBytes      int64     `json:"data_size_bytes"`
	TotalFiles         int       `json:"total_files"`
	ConfigFileCount    int       `json:"config_file_count"`
	EnvFileCount       int       `json:"env_file_count"`
	DatabaseFileCount  int       `json:"database_file_count"`
	WorkspaceFileCount int       `json:"workspace_file_count"`
	SkillsFileCount    int       `json:"skills_file_count"`
	CronFileCount      int       `json:"cron_file_count"`
	BootstrapFileCount int       `json:"bootstrap_file_count"`
	ExportedFiles      []string  `json:"exported_files"`
	FailedFiles        []string  `json:"failed_files"`
}

// BackupOptions defines what to include in the backup
type BackupOptions struct {
	IncludeDatabase  bool `json:"include_database"`
	IncludeWorkspace bool `json:"include_workspace"`
	IncludeConfig    bool `json:"include_config"`
	IncludeEnv       bool `json:"include_env"`
}

// ImportOptions defines how to handle the import
type ImportOptions struct {
	ReplaceDatabase  bool `json:"replace_database"`
	ReplaceWorkspace bool `json:"replace_workspace"`
	ReplaceConfig    bool `json:"replace_config"`
	ReplaceEnv       bool `json:"replace_env"`
}

const (
	maxBackupSize = 500 * 1024 * 1024 // 500MB
	backupVersion = "2.0"
	userDataEntry = "database/user_data.json"
)

func expandTilde(path string) string {
	if path == "" {
		return path
	}
	if path[0] == '~' {
		home, err := os.UserHomeDir()
		if err != nil || home == "" {
			return path
		}
		if len(path) > 1 && path[1] == '/' {
			return home + path[1:]
		}
		return home
	}
	return path
}

// ==================== EXPORT ====================

func (s *Server) handleBackupExport(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	store, userUUID, ok := s.getUserStorage(r)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	// Resolve the full user record for username/ID (needed for logging and filename)
	userIDInt, _ := s.getUserIDFromClaims(r)
	user, err := s.resolveUserByID(userIDInt)
	if err != nil {
		http.Error(w, "user not found", http.StatusNotFound)
		return
	}

	userWorkspace, err := config.EnsureUserWorkspace(userUUID)
	if err != nil {
		http.Error(w, "failed to access workspace", http.StatusInternalServerError)
		return
	}

	var options BackupOptions
	if v := r.URL.Query().Get("include_database"); v != "" {
		options.IncludeDatabase = v == "true"
	} else {
		options.IncludeDatabase = true
	}
	if v := r.URL.Query().Get("include_workspace"); v != "" {
		options.IncludeWorkspace = v == "true"
	} else {
		options.IncludeWorkspace = true
	}
	if v := r.URL.Query().Get("include_config"); v != "" {
		options.IncludeConfig = v == "true"
	} else {
		options.IncludeConfig = true
	}
	if v := r.URL.Query().Get("include_env"); v != "" {
		options.IncludeEnv = v == "true"
	}

	if !options.IncludeDatabase && !options.IncludeWorkspace && !options.IncludeConfig && !options.IncludeEnv {
		http.Error(w, "at least one option must be selected", http.StatusBadRequest)
		return
	}

	userRoot := filepath.Dir(userWorkspace)

	logger.InfoCF("backup", "Starting personal backup export", map[string]interface{}{
		"user":              user.Username,
		"user_id":           user.ID,
		"workspace":         userWorkspace,
		"include_database":  options.IncludeDatabase,
		"include_workspace": options.IncludeWorkspace,
	})

	filename := fmt.Sprintf("makoclaw-%s-%s.makoclaw", user.Username, time.Now().Format("2006-01-02"))

	tempFile, err := os.CreateTemp("", "makoclaw-backup-*.zip")
	if err != nil {
		logger.ErrorCF("backup", "Failed to create temp file", map[string]interface{}{"error": err.Error()})
		http.Error(w, "failed to create backup", http.StatusInternalServerError)
		return
	}
	defer os.Remove(tempFile.Name())
	defer tempFile.Close()

	zipWriter := zip.NewWriter(tempFile)
	zipWriter.RegisterCompressor(zip.Deflate, func(out io.Writer) (io.WriteCloser, error) {
		return flate.NewWriter(out, flate.BestCompression)
	})
	defer zipWriter.Close()

	manifest := BackupManifest{
		Version:         "2.0",
		CreatedAt:       time.Now().UTC(),
		MakoClawVersion: "1.0.0",
		BackupType:      "personal",
		ExportedFiles:   make([]string, 0),
		FailedFiles:     make([]string, 0),
	}

	totalFiles := 0
	totalSize := int64(0)

	// ---- Database: export user data as JSON (not raw .db file) ----
	if options.IncludeDatabase && store != nil {
		userData, err := store.ExportUserData(0)
		if err != nil {
			logger.WarnCF("backup", "Failed to export user data from DB", map[string]interface{}{"error": err.Error()})
			manifest.FailedFiles = append(manifest.FailedFiles, userDataEntry)
		} else {
			dataJSON, err := json.MarshalIndent(userData, "", "  ")
			if err == nil {
				w2, err := zipWriter.Create(userDataEntry)
				if err == nil {
					n, err := w2.Write(dataJSON)
					if err == nil {
						totalFiles++
						totalSize += int64(n)
						manifest.DatabaseFileCount = 1
						manifest.ExportedFiles = append(manifest.ExportedFiles, userDataEntry)
						logger.InfoCF("backup", "Exported user DB data as JSON", map[string]interface{}{
							"sessions": len(userData.Sessions),
							"messages": len(userData.Messages),
							"tasks":    len(userData.Tasks),
							"bytes":    n,
						})
					}
				}
			}
		}
	}

	// ---- Workspace files ----
	if options.IncludeWorkspace {
		count, size, err := addDirToZipWithCounts(zipWriter, userWorkspace, "workspace")
		if err == nil || !os.IsNotExist(err) {
			totalFiles += count
			totalSize += size
			manifest.WorkspaceFileCount = count
			if count > 0 {
				manifest.ExportedFiles = append(manifest.ExportedFiles, "workspace/")
			}
			if err != nil && count == 0 {
				logger.WarnCF("backup", "Failed to add workspace", map[string]interface{}{"error": err.Error()})
				manifest.FailedFiles = append(manifest.FailedFiles, "workspace/")
			}
		}
	}

	// ---- Skills ----
	skillsPath := filepath.Join(userWorkspace, "skills")
	count, size, err := addDirToZipWithCounts(zipWriter, skillsPath, "skills")
	if count > 0 {
		totalFiles += count
		totalSize += size
		manifest.SkillsFileCount = count
		manifest.ExportedFiles = append(manifest.ExportedFiles, "skills/")
	}

	// ---- Cron ----
	cronPath := filepath.Join(userWorkspace, "cron")
	count, size, err = addDirToZipWithCounts(zipWriter, cronPath, "cron")
	if count > 0 {
		totalFiles += count
		totalSize += size
		manifest.CronFileCount = count
		manifest.ExportedFiles = append(manifest.ExportedFiles, "cron/")
	}

	// ---- Bootstrap files ----
	bootstrapFiles := []string{"AGENTS.md", "SOUL.md", "USER.md", "IDENTITY.md"}
	for _, bf := range bootstrapFiles {
		bfPath := filepath.Join(userWorkspace, bf)
		zipEntry := filepath.ToSlash(filepath.Join("workspace", bf))
		c, sz, err := addFileToZipWithCounts(zipWriter, bfPath, zipEntry)
		if err == nil {
			totalFiles += c
			totalSize += sz
			manifest.BootstrapFileCount += c
			manifest.ExportedFiles = append(manifest.ExportedFiles, zipEntry)
		}
	}

	// ---- Config ----
	if options.IncludeConfig {
		configPath := filepath.Join(userRoot, "config.json")
		c, sz, err := addFileToZipWithCounts(zipWriter, configPath, "config.json")
		if err == nil {
			totalFiles += c
			totalSize += sz
			manifest.ConfigFileCount = c
			manifest.ExportedFiles = append(manifest.ExportedFiles, "config.json")
		} else if !os.IsNotExist(err) {
			manifest.FailedFiles = append(manifest.FailedFiles, "config.json")
		}
	}

	// ---- .env ----
	if options.IncludeEnv {
		envPath := filepath.Join(userRoot, ".env")
		c, sz, err := addFileToZipWithCounts(zipWriter, envPath, ".env")
		if err == nil {
			totalFiles += c
			totalSize += sz
			manifest.EnvFileCount = c
			manifest.ExportedFiles = append(manifest.ExportedFiles, ".env")
		} else if !os.IsNotExist(err) {
			manifest.FailedFiles = append(manifest.FailedFiles, ".env")
		}
	}

	manifest.DataSizeBytes = totalSize
	manifest.TotalFiles = totalFiles

	if manifest.TotalFiles == 0 {
		logger.ErrorCF("backup", "Backup would be empty", map[string]interface{}{"user": user.Username})
		http.Error(w, "no data files found to backup", http.StatusBadRequest)
		return
	}

	// ---- Manifest ----
	manifestJSON, _ := json.MarshalIndent(manifest, "", "  ")
	if mf, err := zipWriter.Create("manifest.json"); err == nil {
		mf.Write(manifestJSON)
	}

	if err := zipWriter.Close(); err != nil {
		http.Error(w, "failed to create backup", http.StatusInternalServerError)
		return
	}
	if err := tempFile.Sync(); err != nil {
		http.Error(w, "failed to create backup", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/zip")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	w.Header().Set("Content-Length", fmt.Sprintf("%d", mustGetSize(tempFile.Name())))
	http.ServeFile(w, r, tempFile.Name())

	logger.InfoCF("backup", "Personal backup exported", map[string]interface{}{
		"user":        user.Username,
		"filename":    filename,
		"total_files": totalFiles,
		"size_bytes":  totalSize,
		"db_records":  manifest.DatabaseFileCount,
		"workspace":   manifest.WorkspaceFileCount,
	})
}

// ==================== IMPORT ====================

func (s *Server) handleBackupImport(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	store, userUUID, ok := s.getUserStorage(r)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	// Resolve the full user record for username/ID (needed for logging)
	userIDInt, _ := s.getUserIDFromClaims(r)
	user, err := s.resolveUserByID(userIDInt)
	if err != nil {
		http.Error(w, "user not found", http.StatusNotFound)
		return
	}

	userWorkspace, err := config.EnsureUserWorkspace(userUUID)
	if err != nil {
		http.Error(w, "failed to access workspace", http.StatusInternalServerError)
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, maxBackupSize)
	reader, err := r.MultipartReader()
	if err != nil {
		http.Error(w, "invalid multipart form", http.StatusBadRequest)
		return
	}

	file, err := reader.NextPart()
	if err != nil {
		http.Error(w, "no file uploaded", http.StatusBadRequest)
		return
	}
	if file.FileName() == "" || !strings.HasSuffix(file.FileName(), ".makoclaw") {
		http.Error(w, "invalid file: must be .makoclaw extension", http.StatusBadRequest)
		return
	}

	tempDir, err := os.MkdirTemp("", "makoclaw-import-*")
	if err != nil {
		http.Error(w, "failed to import backup", http.StatusInternalServerError)
		return
	}
	defer os.RemoveAll(tempDir)

	zipPath := filepath.Join(tempDir, "backup.zip")
	zipFile, err := os.Create(zipPath)
	if err != nil {
		http.Error(w, "failed to import backup", http.StatusInternalServerError)
		return
	}

	_, copyErr := io.Copy(zipFile, file)
	// Sync and explicitly close before opening as zip reader — deferred close is too late.
	_ = zipFile.Sync()
	_ = zipFile.Close()
	if copyErr != nil {
		http.Error(w, "failed to save uploaded file", http.StatusInternalServerError)
		return
	}

	zipReader, err := zip.OpenReader(zipPath)
	if err != nil {
		http.Error(w, "invalid backup file", http.StatusBadRequest)
		return
	}
	defer zipReader.Close()

	// Read manifest
	var manifest BackupManifest
	var manifestFound bool
	for _, f := range zipReader.File {
		if f.Name == "manifest.json" {
			if rc, err := f.Open(); err == nil {
				if json.NewDecoder(rc).Decode(&manifest) == nil {
					manifestFound = true
				}
				rc.Close()
			}
			break
		}
	}
	if !manifestFound {
		http.Error(w, "invalid backup: missing manifest.json", http.StatusBadRequest)
		return
	}

	// Parse import options
	var importOptions ImportOptions
	if body := r.FormValue("options"); body != "" {
		json.Unmarshal([]byte(body), &importOptions)
	} else {
		importOptions.ReplaceDatabase = true
		importOptions.ReplaceWorkspace = true
		// In multi-user mode config.json is user-scoped, so restoring it by default
		// is expected to recover provider/model/channel/email settings.
		importOptions.ReplaceConfig = true
		importOptions.ReplaceEnv = false
	}
	if !importOptions.ReplaceDatabase && !importOptions.ReplaceWorkspace && !importOptions.ReplaceConfig && !importOptions.ReplaceEnv {
		http.Error(w, "at least one replace option must be selected", http.StatusBadRequest)
		return
	}

	userRoot := filepath.Dir(userWorkspace)
	importedFiles := 0
	importedDBSessions := 0
	importedDBMessages := 0
	importedDBTasks := 0
	var importErrors []string

	// Detect if this is a v2 (JSON data) or legacy v1 (raw DB files) backup
	isV2 := false
	hasLegacyDB := false
	for _, f := range zipReader.File {
		if f.Name == userDataEntry {
			isV2 = true
		}
		if strings.HasPrefix(f.Name, "database/") && strings.HasSuffix(strings.ToLower(f.Name), ".db") {
			hasLegacyDB = true
		}
	}

	logger.InfoCF("backup", "Starting personal backup import", map[string]interface{}{
		"user":           user.Username,
		"user_id":        user.ID,
		"manifest_ver":   manifest.Version,
		"is_v2":          isV2,
		"has_legacy_db":  hasLegacyDB,
		"replace_db":     importOptions.ReplaceDatabase,
		"replace_ws":     importOptions.ReplaceWorkspace,
		"replace_config": importOptions.ReplaceConfig,
	})

	// Track if we already handled legacy DB (process main .db file only once)
	legacyDBHandled := false

	// ---- Process each file in the ZIP ----
	for _, f := range zipReader.File {
		if f.Name == "manifest.json" {
			continue
		}

		// ---- DATABASE: v2 JSON data import (no file replacement!) ----
		if f.Name == userDataEntry {
			if !importOptions.ReplaceDatabase {
				continue
			}
			rc, err := f.Open()
			if err != nil {
				importErrors = append(importErrors, "failed to read user_data.json: "+err.Error())
				continue
			}
			var userData storage.BackupUserData
			if err := json.NewDecoder(rc).Decode(&userData); err != nil {
				rc.Close()
				importErrors = append(importErrors, "failed to parse user_data.json: "+err.Error())
				continue
			}
			rc.Close()

			sessions, messages, tasks, err := store.ImportUserData(0, &userData)
			if err != nil {
				importErrors = append(importErrors, "failed to import DB data: "+err.Error())
			} else {
				importedDBSessions = sessions
				importedDBMessages = messages
				importedDBTasks = tasks
				logger.InfoCF("backup", "Imported user DB data", map[string]interface{}{
					"sessions": sessions, "messages": messages, "tasks": tasks,
				})
			}
			continue
		}

		// ---- LEGACY: raw .db files from v1 backups ----
		if strings.HasPrefix(f.Name, "database/") {
			if !importOptions.ReplaceDatabase {
				continue
			}
			// For legacy backups with raw DB, extract and import data via SQL
			if hasLegacyDB && !isV2 && !legacyDBHandled {
				// First extract all DB files to temp dir
				dbFileName := filepath.Base(f.Name)
				tempDBPath := filepath.Join(tempDir, dbFileName)
				if err := extractZipFile(f, tempDBPath); err != nil {
					importErrors = append(importErrors, "legacy DB extract: "+err.Error())
				}
				// Only process once we see the main .db file
				lower := strings.ToLower(dbFileName)
				if lower == "makoclaw.db" {
					legacyDBHandled = true
					if err := s.importLegacyDB(tempDBPath, store); err != nil {
						importErrors = append(importErrors, "legacy DB import: "+err.Error())
					} else {
						logger.InfoCF("backup", "Imported legacy DB data for user", map[string]interface{}{"user": user.Username})
					}
				}
			}
			continue
		}

		// ---- WORKSPACE files ----
		if strings.HasPrefix(f.Name, "workspace/") {
			if !importOptions.ReplaceWorkspace {
				continue
			}
			relPath := strings.TrimPrefix(f.Name, "workspace/")
			if relPath == "" {
				continue
			}
			targetPath := filepath.Join(userWorkspace, relPath)
			if err := extractZipFile(f, targetPath); err != nil {
				importErrors = append(importErrors, fmt.Sprintf("workspace/%s: %s", relPath, err.Error()))
			} else {
				importedFiles++
			}
			continue
		}

		// ---- SKILLS ----
		if strings.HasPrefix(f.Name, "skills/") {
			if !importOptions.ReplaceWorkspace {
				continue
			}
			relPath := strings.TrimPrefix(f.Name, "skills/")
			if relPath == "" {
				continue
			}
			targetPath := filepath.Join(userWorkspace, "skills", relPath)
			if err := extractZipFile(f, targetPath); err != nil {
				importErrors = append(importErrors, fmt.Sprintf("skills/%s: %s", relPath, err.Error()))
			} else {
				importedFiles++
			}
			continue
		}

		// ---- CRON ----
		if strings.HasPrefix(f.Name, "cron/") {
			if !importOptions.ReplaceWorkspace {
				continue
			}
			relPath := strings.TrimPrefix(f.Name, "cron/")
			if relPath == "" {
				continue
			}
			targetPath := filepath.Join(userWorkspace, "cron", relPath)
			if err := extractZipFile(f, targetPath); err != nil {
				importErrors = append(importErrors, fmt.Sprintf("cron/%s: %s", relPath, err.Error()))
			} else {
				importedFiles++
			}
			continue
		}

		// ---- CONFIG ----
		if f.Name == "config.json" {
			if !importOptions.ReplaceConfig {
				continue
			}
			targetPath := filepath.Join(userRoot, "config.json")
			if err := extractZipFile(f, targetPath); err != nil {
				importErrors = append(importErrors, "config.json: "+err.Error())
			} else {
				importedFiles++
			}
			continue
		}

		// ---- ENV ----
		if f.Name == ".env" {
			if !importOptions.ReplaceEnv {
				continue
			}
			targetPath := filepath.Join(userRoot, ".env")
			if err := extractZipFile(f, targetPath); err != nil {
				importErrors = append(importErrors, ".env: "+err.Error())
			} else {
				importedFiles++
			}
			continue
		}

		// Unknown entry — skip silently
		logger.WarnCF("backup", "Skipping unknown backup entry", map[string]interface{}{"name": f.Name})
	}

	logger.InfoCF("backup", "Personal backup imported", map[string]interface{}{
		"user":        user.Username,
		"files":       importedFiles,
		"db_sessions": importedDBSessions,
		"db_messages": importedDBMessages,
		"db_tasks":    importedDBTasks,
		"errors":      len(importErrors),
	})

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"ok":      len(importErrors) == 0,
		"message": fmt.Sprintf("Imported %d files, %d sessions, %d messages, %d tasks", importedFiles, importedDBSessions, importedDBMessages, importedDBTasks),
		"details": map[string]interface{}{
			"files_restored":    importedFiles,
			"sessions_imported": importedDBSessions,
			"messages_imported": importedDBMessages,
			"tasks_imported":    importedDBTasks,
		},
		"errors":   importErrors,
		"manifest": manifest,
		"imported_by": map[string]interface{}{
			"username":  user.Username,
			"user_id":   user.ID,
			"user_uuid": user.UUID,
			"workspace": userWorkspace,
		},
	})
}

// importLegacyDB handles v1 backups that contain raw .db files.
// It opens the DB in a temp location, reads user data from it (user_id=1),
// and imports it into the target per-user storage. No global DB replacement.
func (s *Server) importLegacyDB(tempDBPath string, targetStore *storage.Storage) error {
	legacyStore, err := storage.New(config.StorageConfig{Path: tempDBPath})
	if err != nil {
		return fmt.Errorf("opening legacy DB: %w", err)
	}
	defer legacyStore.Close()

	// Export data from legacy DB (user_id=1 is the default single-user ID)
	userData, err := legacyStore.ExportUserData(1)
	if err != nil {
		return fmt.Errorf("reading legacy data: %w", err)
	}

	sessions, messages, tasks, err := targetStore.ImportUserData(0, userData)
	if err != nil {
		return fmt.Errorf("importing legacy data: %w", err)
	}

	logger.InfoCF("backup", "Legacy DB data imported", map[string]interface{}{
		"sessions": sessions, "messages": messages, "tasks": tasks,
	})
	return nil
}

// ==================== VALIDATE ====================

func (s *Server) handleBackupValidate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, maxBackupSize)
	reader, err := r.MultipartReader()
	if err != nil {
		http.Error(w, "invalid multipart form", http.StatusBadRequest)
		return
	}

	file, err := reader.NextPart()
	if err != nil {
		http.Error(w, "no file uploaded", http.StatusBadRequest)
		return
	}
	if file.FileName() == "" || !strings.HasSuffix(file.FileName(), ".makoclaw") {
		http.Error(w, "invalid file: must be .makoclaw extension", http.StatusBadRequest)
		return
	}

	tempFile, err := os.CreateTemp("", "makoclaw-validate-*.zip")
	if err != nil {
		http.Error(w, "failed to validate backup", http.StatusInternalServerError)
		return
	}
	defer os.Remove(tempFile.Name())
	defer tempFile.Close()

	if _, err := io.Copy(tempFile, file); err != nil {
		http.Error(w, "failed to save uploaded file", http.StatusInternalServerError)
		return
	}

	zipReader, err := zip.OpenReader(tempFile.Name())
	if err != nil {
		http.Error(w, "invalid backup file", http.StatusBadRequest)
		return
	}
	defer zipReader.Close()

	var manifest BackupManifest
	var manifestFound bool
	manifestJSON, _ := json.Marshal(BackupManifest{})

	for _, f := range zipReader.File {
		if f.Name == "manifest.json" {
			if rc, err := f.Open(); err == nil {
				if json.NewDecoder(rc).Decode(&manifest) == nil {
					manifestFound = true
					manifestJSON, _ = json.MarshalIndent(manifest, "", "  ")
				}
				rc.Close()
			}
			break
		}
	}

	if !manifestFound {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"valid":   false,
			"error":   "missing manifest.json",
			"files":   len(zipReader.File),
			"version": "unknown",
		})
		return
	}

	// Detect backup format
	hasUserDataJSON := false
	hasRawDB := false
	for _, f := range zipReader.File {
		if f.Name == userDataEntry {
			hasUserDataJSON = true
		}
		if strings.HasPrefix(f.Name, "database/") && strings.HasSuffix(strings.ToLower(f.Name), ".db") {
			hasRawDB = true
		}
	}

	backupFormat := "v2_personal"
	if !hasUserDataJSON && hasRawDB {
		backupFormat = "v1_legacy"
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"valid":                true,
		"version":              manifest.Version,
		"backup_type":          manifest.BackupType,
		"backup_format":        backupFormat,
		"created_at":           manifest.CreatedAt,
		"makoclaw_version":     manifest.MakoClawVersion,
		"config_file_count":    manifest.ConfigFileCount,
		"env_file_count":       manifest.EnvFileCount,
		"database_file_count":  manifest.DatabaseFileCount,
		"workspace_file_count": manifest.WorkspaceFileCount,
		"data_size_bytes":      manifest.DataSizeBytes,
		"total_files":          manifest.TotalFiles,
		"exported_files":       manifest.ExportedFiles,
		"failed_files":         manifest.FailedFiles,
		"manifest":             string(manifestJSON),
		"includes_database":    manifest.DatabaseFileCount > 0 || hasUserDataJSON,
		"includes_workspace":   manifest.WorkspaceFileCount > 0,
		"includes_config":      manifest.ConfigFileCount > 0,
		"includes_env":         manifest.EnvFileCount > 0,
		"has_any_content":      manifest.TotalFiles > 0,
	})
}

// ==================== HELPERS ====================

// extractZipFile extracts a single zip entry to the given target path,
// creating parent directories as needed. Directory entries are skipped.
func extractZipFile(f *zip.File, targetPath string) error {
	// Skip directory entries
	if f.FileInfo().IsDir() {
		return nil
	}

	// Security: prevent path traversal
	cleanName := filepath.Clean(f.Name)
	if strings.Contains(cleanName, "..") {
		return fmt.Errorf("invalid path: %s", f.Name)
	}

	if err := os.MkdirAll(filepath.Dir(targetPath), 0755); err != nil {
		return fmt.Errorf("creating directory: %w", err)
	}

	src, err := f.Open()
	if err != nil {
		return fmt.Errorf("opening zip entry: %w", err)
	}
	defer src.Close()

	dst, err := os.OpenFile(targetPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return fmt.Errorf("creating file: %w", err)
	}
	defer dst.Close()

	if _, err := io.Copy(dst, src); err != nil {
		return fmt.Errorf("writing file: %w", err)
	}
	return nil
}

// addFileToZipWithCounts adds a single file to the zip and returns count, size, error
func addFileToZipWithCounts(zipWriter *zip.Writer, filePath, zipPath string) (int, int64, error) {
	info, err := os.Stat(filePath)
	if err != nil {
		return 0, 0, err
	}
	if info.IsDir() {
		return addDirToZipWithCounts(zipWriter, filePath, zipPath)
	}

	file, err := os.Open(filePath)
	if err != nil {
		return 0, 0, err
	}
	defer file.Close()

	header, err := zip.FileInfoHeader(info)
	if err != nil {
		return 0, 0, err
	}
	header.Name = filepath.ToSlash(zipPath)
	header.Method = zip.Deflate

	writer, err := zipWriter.CreateHeader(header)
	if err != nil {
		return 0, 0, err
	}

	_, err = io.Copy(writer, file)
	if err != nil {
		return 0, 0, err
	}
	return 1, info.Size(), nil
}

// addDirToZipWithCounts adds a directory to the zip and returns file count, total size, error
func addDirToZipWithCounts(zipWriter *zip.Writer, dirPath, zipPath string) (int, int64, error) {
	fileCount := 0
	totalSize := int64(0)

	err := filepath.Walk(dirPath, func(filePath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(dirPath, filePath)
		if err != nil {
			return err
		}

		zipEntryPath := filepath.ToSlash(filepath.Join(zipPath, relPath))

		if info.IsDir() {
			return nil
		}

		file, err := os.Open(filePath)
		if err != nil {
			return err
		}
		defer file.Close()

		header, err := zip.FileInfoHeader(info)
		if err != nil {
			return err
		}
		header.Name = zipEntryPath
		header.Method = zip.Deflate

		writer, err := zipWriter.CreateHeader(header)
		if err != nil {
			return err
		}

		if _, err := io.Copy(writer, file); err != nil {
			return err
		}

		fileCount++
		totalSize += info.Size()
		return nil
	})

	return fileCount, totalSize, err
}

// addFileToZip (DEPRECATED: use addFileToZipWithCounts) adds a single file to the zip
func addFileToZip(zipWriter *zip.Writer, filePath, zipPath string) error {
	_, _, err := addFileToZipWithCounts(zipWriter, filePath, zipPath)
	return err
}

// addDirToZip (DEPRECATED: use addDirToZipWithCounts) adds a directory to the zip
func addDirToZip(zipWriter *zip.Writer, dirPath, zipPath string, fileCount *int, totalSize *int64) error {
	count, size, err := addDirToZipWithCounts(zipWriter, dirPath, zipPath)
	if fileCount != nil {
		*fileCount += count
	}
	if totalSize != nil {
		*totalSize += size
	}
	return err
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func mustGetSize(path string) int64 {
	info, err := os.Stat(path)
	if err != nil {
		return 0
	}
	return info.Size()
}
