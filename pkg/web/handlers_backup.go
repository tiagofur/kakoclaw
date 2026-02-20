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

	"github.com/sipeed/kakoclaw/pkg/logger"
)

// BackupManifest represents the metadata of a backup archive
type BackupManifest struct {
	Version            string    `json:"version"`
	CreatedAt          time.Time `json:"created_at"`
	KakoClawVersion    string    `json:"kakoclaw_version"`
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
	backupVersion = "1.0"
)

// ==================== BACKUP HANDLERS ====================

func (s *Server) handleBackupExport(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var options BackupOptions
	if includeDB := r.URL.Query().Get("include_database"); includeDB != "" {
		options.IncludeDatabase = includeDB == "true"
	} else {
		options.IncludeDatabase = true
	}
	if includeWS := r.URL.Query().Get("include_workspace"); includeWS != "" {
		options.IncludeWorkspace = includeWS == "true"
	} else {
		options.IncludeWorkspace = true
	}
	if includeConfig := r.URL.Query().Get("include_config"); includeConfig != "" {
		options.IncludeConfig = includeConfig == "true"
	}
	if includeEnv := r.URL.Query().Get("include_env"); includeEnv != "" {
		options.IncludeEnv = includeEnv == "true"
	}

	if !options.IncludeDatabase && !options.IncludeWorkspace && !options.IncludeConfig && !options.IncludeEnv {
		http.Error(w, "at least one option must be selected", http.StatusBadRequest)
		return
	}

	workspacePath := filepath.Join(s.workspace, "..")
	dataDir := workspacePath

	logger.InfoCF("backup", "Starting backup", map[string]interface{}{
		"workspace":         s.workspace,
		"workspacePath":     workspacePath,
		"dataDir":           dataDir,
		"include_database":  options.IncludeDatabase,
		"include_workspace": options.IncludeWorkspace,
	})

	filename := fmt.Sprintf("kakoclaw-%s.kakoclaw", time.Now().Format("2006-01-02"))

	// Create temporary file for the ZIP
	tempFile, err := os.CreateTemp("", "kakoclaw-backup-*.zip")
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
		Version:         backupVersion,
		CreatedAt:       time.Now().UTC(),
		KakoClawVersion: "1.0.0",
		ExportedFiles:   make([]string, 0),
		FailedFiles:     make([]string, 0),
	}

	totalFiles := 0
	totalSize := int64(0)

	// Add database files
	if options.IncludeDatabase {
		dbFiles := []string{"KakoClaw.db", "KakoClaw.db-shm", "KakoClaw.db-wal"}
		for _, dbFile := range dbFiles {
			dbPath := filepath.Join(dataDir, dbFile)
			count, size, err := addFileToZipWithCounts(zipWriter, dbPath, filepath.Join("database", filepath.Base(dbFile)))
			if err == nil {
				totalFiles += count
				totalSize += size
				manifest.DatabaseFileCount += count
				manifest.ExportedFiles = append(manifest.ExportedFiles, filepath.Join("database", filepath.Base(dbFile)))
				logger.InfoCF("backup", "Added database file", map[string]interface{}{"file": dbFile})
			} else if !os.IsNotExist(err) {
				logger.WarnCF("backup", "Failed to add database file", map[string]interface{}{"file": dbFile, "error": err.Error()})
				manifest.FailedFiles = append(manifest.FailedFiles, filepath.Join("database", filepath.Base(dbFile)))
			}
		}
	}

	// Add workspace directory
	if options.IncludeWorkspace {
		workspaceFull := filepath.Join(dataDir, "workspace")
		count, size, err := addDirToZipWithCounts(zipWriter, workspaceFull, "workspace")
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

	// Add skills directory
	skillsPath := filepath.Join(dataDir, "skills")
	count, size, err := addDirToZipWithCounts(zipWriter, skillsPath, "skills")
	if count > 0 {
		totalFiles += count
		totalSize += size
		manifest.SkillsFileCount = count
		manifest.ExportedFiles = append(manifest.ExportedFiles, "skills/")
		logger.InfoCF("backup", "Added skills directory", map[string]interface{}{"count": count})
	} else if err != nil && !os.IsNotExist(err) {
		logger.WarnCF("backup", "Failed to add skills", map[string]interface{}{"error": err.Error()})
	}

	// Add cron directory
	cronPath := filepath.Join(dataDir, "cron")
	count, size, err = addDirToZipWithCounts(zipWriter, cronPath, "cron")
	if count > 0 {
		totalFiles += count
		totalSize += size
		manifest.CronFileCount = count
		manifest.ExportedFiles = append(manifest.ExportedFiles, "cron/")
		logger.InfoCF("backup", "Added cron directory", map[string]interface{}{"count": count})
	} else if err != nil && !os.IsNotExist(err) {
		logger.WarnCF("backup", "Failed to add cron", map[string]interface{}{"error": err.Error()})
	}

	// Add bootstrap files
	bootstrapFiles := []string{"AGENTS.md", "SOUL.md", "USER.md", "IDENTITY.md"}
	for _, bootstrapFile := range bootstrapFiles {
		bootstrapPath := filepath.Join(dataDir, "workspace", bootstrapFile)
		count, size, err := addFileToZipWithCounts(zipWriter, bootstrapPath, filepath.Join("workspace", bootstrapFile))
		if err == nil {
			totalFiles += count
			totalSize += size
			manifest.BootstrapFileCount += count
			manifest.ExportedFiles = append(manifest.ExportedFiles, filepath.Join("workspace", bootstrapFile))
			logger.InfoCF("backup", "Added bootstrap file", map[string]interface{}{"file": bootstrapFile})
		} else if !os.IsNotExist(err) {
			logger.WarnCF("backup", "Failed to add bootstrap file", map[string]interface{}{"file": bootstrapFile, "error": err.Error()})
		}
	}

	// Add config.json
	if options.IncludeConfig {
		configPath := filepath.Join(dataDir, "config.json")
		count, size, err := addFileToZipWithCounts(zipWriter, configPath, "config.json")
		if err == nil {
			totalFiles += count
			totalSize += size
			manifest.ConfigFileCount = count
			manifest.ExportedFiles = append(manifest.ExportedFiles, "config.json")
			logger.InfoCF("backup", "Added config.json", map[string]interface{}{})
		} else if !os.IsNotExist(err) {
			logger.WarnCF("backup", "Failed to add config.json", map[string]interface{}{"error": err.Error()})
			manifest.FailedFiles = append(manifest.FailedFiles, "config.json")
		}
	}

	// Add .env file
	if options.IncludeEnv {
		envPath := filepath.Join(dataDir, ".env")
		count, size, err := addFileToZipWithCounts(zipWriter, envPath, ".env")
		if err == nil {
			totalFiles += count
			totalSize += size
			manifest.EnvFileCount = count
			manifest.ExportedFiles = append(manifest.ExportedFiles, ".env")
			logger.InfoCF("backup", "Added .env file", map[string]interface{}{})
		} else if !os.IsNotExist(err) {
			logger.WarnCF("backup", "Failed to add .env file", map[string]interface{}{"error": err.Error()})
			manifest.FailedFiles = append(manifest.FailedFiles, ".env")
		}
	}

	manifest.DataSizeBytes = totalSize
	manifest.TotalFiles = totalFiles

	// Validate that something was exported
	if manifest.TotalFiles == 0 {
		logger.ErrorCF("backup", "Backup would be empty - no files found", map[string]interface{}{
			"data_dir": dataDir,
			"options":  options,
		})
		http.Error(w, "no data files found to backup", http.StatusBadRequest)
		return
	}

	// Add manifest.json
	manifestJSON, _ := json.MarshalIndent(manifest, "", "  ")
	manifestPath := "manifest.json"
	if manifestFile, err := zipWriter.Create(manifestPath); err == nil {
		manifestFile.Write(manifestJSON)
		logger.InfoCF("backup", "Added manifest.json", map[string]interface{}{})
	}

	if err := zipWriter.Close(); err != nil {
		logger.ErrorCF("backup", "Failed to close zip writer", map[string]interface{}{"error": err.Error()})
		http.Error(w, "failed to create backup", http.StatusInternalServerError)
		return
	}

	if err := tempFile.Sync(); err != nil {
		logger.ErrorCF("backup", "Failed to sync temp file", map[string]interface{}{"error": err.Error()})
		http.Error(w, "failed to create backup", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/zip")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	w.Header().Set("Content-Length", fmt.Sprintf("%d", mustGetSize(tempFile.Name())))

	http.ServeFile(w, r, tempFile.Name())

	logger.InfoCF("backup", "Backup exported successfully", map[string]interface{}{
		"filename":        filename,
		"size_bytes":      totalSize,
		"total_files":     totalFiles,
		"exported_files":  len(manifest.ExportedFiles),
		"failed_files":    len(manifest.FailedFiles),
		"database_count":  manifest.DatabaseFileCount,
		"workspace_count": manifest.WorkspaceFileCount,
		"skills_count":    manifest.SkillsFileCount,
		"cron_count":      manifest.CronFileCount,
		"bootstrap_count": manifest.BootstrapFileCount,
		"config_count":    manifest.ConfigFileCount,
		"env_count":       manifest.EnvFileCount,
	})
}

func (s *Server) handleBackupImport(w http.ResponseWriter, r *http.Request) {
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

	if file.FileName() == "" || !strings.HasSuffix(file.FileName(), ".kakoclaw") {
		http.Error(w, "invalid file: must be .kakoclaw extension", http.StatusBadRequest)
		return
	}

	tempDir, err := os.MkdirTemp("", "kakoclaw-import-*")
	if err != nil {
		logger.ErrorCF("backup", "Failed to create temp dir", map[string]interface{}{"error": err.Error()})
		http.Error(w, "failed to import backup", http.StatusInternalServerError)
		return
	}
	defer os.RemoveAll(tempDir)

	zipPath := filepath.Join(tempDir, "backup.zip")
	zipFile, err := os.Create(zipPath)
	if err != nil {
		logger.ErrorCF("backup", "Failed to create temp zip", map[string]interface{}{"error": err.Error()})
		http.Error(w, "failed to import backup", http.StatusInternalServerError)
		return
	}
	defer zipFile.Close()

	if _, err := io.Copy(zipFile, file); err != nil {
		logger.ErrorCF("backup", "Failed to save uploaded file", map[string]interface{}{"error": err.Error()})
		http.Error(w, "failed to save uploaded file", http.StatusInternalServerError)
		return
	}

	zipReader, err := zip.OpenReader(zipPath)
	if err != nil {
		logger.ErrorCF("backup", "Failed to open zip file", map[string]interface{}{"error": err.Error()})
		http.Error(w, "invalid backup file", http.StatusBadRequest)
		return
	}
	defer zipReader.Close()

	var manifest BackupManifest
	var manifestFound bool

	for _, f := range zipReader.File {
		if f.Name == "manifest.json" {
			manifestFile, err := f.Open()
			if err != nil {
				continue
			}
			defer manifestFile.Close()

			decoder := json.NewDecoder(manifestFile)
			if err := decoder.Decode(&manifest); err == nil {
				manifestFound = true
			}
			break
		}
	}

	if !manifestFound {
		http.Error(w, "invalid backup: missing manifest.json", http.StatusBadRequest)
		return
	}

	workspacePath := filepath.Join(s.workspace, "..")
	dataDir := filepath.Join(workspacePath, ".KakoClaw")

	var importOptions ImportOptions

	if body := r.FormValue("options"); body != "" {
		json.Unmarshal([]byte(body), &importOptions)
	} else {
		importOptions.ReplaceDatabase = true
		importOptions.ReplaceWorkspace = true
		importOptions.ReplaceConfig = true
		importOptions.ReplaceEnv = true
	}

	if !importOptions.ReplaceDatabase && !importOptions.ReplaceWorkspace && !importOptions.ReplaceConfig && !importOptions.ReplaceEnv {
		http.Error(w, "at least one replace option must be selected", http.StatusBadRequest)
		return
	}

	if importOptions.ReplaceDatabase && s.store != nil {
		if err := s.store.Close(); err != nil {
			logger.WarnCF("backup", "Failed to close database", map[string]interface{}{"error": err.Error()})
		}
		time.Sleep(100 * time.Millisecond)
	}

	backupDir := filepath.Join(dataDir, "backup-before-import-"+time.Now().Format("20060102-150405"))

	for _, f := range zipReader.File {
		if f.Name == "manifest.json" {
			continue
		}

		targetPath := filepath.Join(dataDir, filepath.Clean(f.Name))

		if strings.HasPrefix(f.Name, "database/") {
			if !importOptions.ReplaceDatabase {
				continue
			}
			// Map old lowercase filenames to new uppercase filenames
			dbFileName := filepath.Base(f.Name)
			if dbFileName == "kakoclaw.db" {
				targetPath = filepath.Join(dataDir, "KakoClaw.db")
			} else if dbFileName == "kakoclaw.db-shm" {
				targetPath = filepath.Join(dataDir, "KakoClaw.db-shm")
			} else if dbFileName == "kakoclaw.db-wal" {
				targetPath = filepath.Join(dataDir, "KakoClaw.db-wal")
			}
		}

		if strings.HasPrefix(f.Name, "workspace/") {
			if !importOptions.ReplaceWorkspace {
				continue
			}
		}

		if f.Name == "config.json" {
			if !importOptions.ReplaceConfig {
				continue
			}
			targetPath = filepath.Join(dataDir, "config.json")
		}

		if f.Name == ".env" {
			if !importOptions.ReplaceEnv {
				continue
			}
			targetPath = filepath.Join(workspacePath, ".env")
		}

		if err := os.MkdirAll(filepath.Dir(targetPath), 0755); err != nil {
			logger.WarnCF("backup", "Failed to create directory", map[string]interface{}{"path": filepath.Dir(targetPath), "error": err.Error()})
			continue
		}

		if fileExists(targetPath) {
			backupFilePath := filepath.Join(backupDir, filepath.Base(f.Name))
			if err := os.MkdirAll(filepath.Dir(backupFilePath), 0755); err == nil {
				if err := os.Rename(targetPath, backupFilePath); err == nil {
					logger.InfoCF("backup", "Backed up existing file", map[string]interface{}{"file": targetPath, "backup": backupFilePath})
				}
			}
		}

		src, err := f.Open()
		if err != nil {
			logger.ErrorCF("backup", "Failed to open file in zip", map[string]interface{}{"file": f.Name, "error": err.Error()})
			continue
		}
		defer src.Close()

		dst, err := os.OpenFile(targetPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
		if err != nil {
			logger.ErrorCF("backup", "Failed to create file", map[string]interface{}{"file": targetPath, "error": err.Error()})
			continue
		}
		defer dst.Close()

		if _, err := io.Copy(dst, src); err != nil {
			logger.ErrorCF("backup", "Failed to copy file", map[string]interface{}{"file": targetPath, "error": err.Error()})
			continue
		}
	}

	logger.InfoCF("backup", "Backup imported successfully", map[string]interface{}{
		"version":           manifest.Version,
		"created_at":        manifest.CreatedAt,
		"replace_database":  importOptions.ReplaceDatabase,
		"replace_workspace": importOptions.ReplaceWorkspace,
		"replace_config":    importOptions.ReplaceConfig,
		"replace_env":       importOptions.ReplaceEnv,
	})

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"ok":         true,
		"message":    "Backup imported successfully",
		"backup_dir": backupDir,
		"manifest":   manifest,
	})
}

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

	if file.FileName() == "" || !strings.HasSuffix(file.FileName(), ".kakoclaw") {
		http.Error(w, "invalid file: must be .kakoclaw extension", http.StatusBadRequest)
		return
	}

	tempFile, err := os.CreateTemp("", "kakoclaw-validate-*.zip")
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
			manifestFile, err := f.Open()
			if err != nil {
				continue
			}

			decoder := json.NewDecoder(manifestFile)
			if err := decoder.Decode(&manifest); err == nil {
				manifestFound = true
				manifestJSON, _ = json.MarshalIndent(manifest, "", "  ")
			}
			manifestFile.Close()
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

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"valid":                true,
		"version":              manifest.Version,
		"created_at":           manifest.CreatedAt,
		"kakoclaw_version":     manifest.KakoClawVersion,
		"config_file_count":    manifest.ConfigFileCount,
		"env_file_count":       manifest.EnvFileCount,
		"database_file_count":  manifest.DatabaseFileCount,
		"workspace_file_count": manifest.WorkspaceFileCount,
		"data_size_bytes":      manifest.DataSizeBytes,
		"total_files":          manifest.TotalFiles,
		"exported_files":       manifest.ExportedFiles,
		"failed_files":         manifest.FailedFiles,
		"manifest":             string(manifestJSON),
	})
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
	header.Name = zipPath
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

		zipEntryPath := filepath.Join(zipPath, relPath)

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
