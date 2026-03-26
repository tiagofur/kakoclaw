package storage

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"
)

// InstalledPackage represents a tracked package installation.
type InstalledPackage struct {
	ID               int64     `json:"id"`
	Name             string    `json:"name"`
	Version          string    `json:"version"`
	PackageManager   string    `json:"package_manager"`
	Status           string    `json:"status"`
	InstalledBy      int64     `json:"installed_by"`
	InstallCommand   string    `json:"install_command"`
	UninstallCommand string    `json:"uninstall_command"`
	OutputLog        string    `json:"output_log,omitempty"`
	ErrorLog         string    `json:"error_log,omitempty"`
	Metadata         string    `json:"metadata,omitempty"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

func (s *Storage) migratePackages() error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS installed_packages (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			version TEXT NOT NULL DEFAULT '',
			package_manager TEXT NOT NULL,
			status TEXT NOT NULL DEFAULT 'installed',
			installed_by INTEGER NOT NULL DEFAULT 1,
			install_command TEXT NOT NULL,
			uninstall_command TEXT NOT NULL DEFAULT '',
			output_log TEXT NOT NULL DEFAULT '',
			error_log TEXT NOT NULL DEFAULT '',
			metadata TEXT NOT NULL DEFAULT '{}',
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);`,
		`CREATE UNIQUE INDEX IF NOT EXISTS idx_installed_packages_name_pm ON installed_packages(name, package_manager);`,
	}
	for _, q := range queries {
		if _, err := s.db.Exec(q); err != nil {
			return fmt.Errorf("packages migration: %w", err)
		}
	}
	return nil
}

// RecordPackage upserts a package record keyed by (name, package_manager).
func (s *Storage) RecordPackage(pkg InstalledPackage) (int64, error) {
	result, err := s.db.Exec(`
		INSERT INTO installed_packages (name, version, package_manager, status, installed_by, install_command, uninstall_command, output_log, error_log, metadata)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(name, package_manager) DO UPDATE SET
			version = excluded.version,
			status = excluded.status,
			installed_by = excluded.installed_by,
			install_command = excluded.install_command,
			uninstall_command = excluded.uninstall_command,
			output_log = excluded.output_log,
			error_log = excluded.error_log,
			metadata = excluded.metadata,
			updated_at = CURRENT_TIMESTAMP
	`, pkg.Name, pkg.Version, pkg.PackageManager, pkg.Status, pkg.InstalledBy,
		pkg.InstallCommand, pkg.UninstallCommand, pkg.OutputLog, pkg.ErrorLog, pkg.Metadata)
	if err != nil {
		return 0, fmt.Errorf("recording package: %w", err)
	}
	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("getting package id: %w", err)
	}
	return id, nil
}

// UpdatePackageStatus updates a package record after install/uninstall completes.
func (s *Storage) UpdatePackageStatus(id int64, status, outputLog, errorLog, version string) error {
	_, err := s.db.Exec(`
		UPDATE installed_packages
		SET status = ?, output_log = ?, error_log = ?, version = ?, updated_at = CURRENT_TIMESTAMP
		WHERE id = ?
	`, status, outputLog, errorLog, version, id)
	if err != nil {
		return fmt.Errorf("updating package status: %w", err)
	}
	return nil
}

// GetPackageByNameAndPM retrieves a package by name and package manager.
func (s *Storage) GetPackageByNameAndPM(name, pm string) (*InstalledPackage, error) {
	row := s.db.QueryRow(`
		SELECT id, name, version, package_manager, status, installed_by,
		       install_command, uninstall_command, output_log, error_log, metadata,
		       created_at, updated_at
		FROM installed_packages
		WHERE name = ? AND package_manager = ?
	`, name, pm)
	return scanPackage(row)
}

// ListPackages returns all tracked packages, optionally filtered by status.
func (s *Storage) ListPackages(statusFilter string) ([]InstalledPackage, error) {
	var rows *sql.Rows
	var err error
	if statusFilter != "" && statusFilter != "all" {
		rows, err = s.db.Query(`
			SELECT id, name, version, package_manager, status, installed_by,
			       install_command, uninstall_command, output_log, error_log, metadata,
			       created_at, updated_at
			FROM installed_packages WHERE status = ?
			ORDER BY updated_at DESC
		`, statusFilter)
	} else {
		rows, err = s.db.Query(`
			SELECT id, name, version, package_manager, status, installed_by,
			       install_command, uninstall_command, output_log, error_log, metadata,
			       created_at, updated_at
			FROM installed_packages
			ORDER BY updated_at DESC
		`)
	}
	if err != nil {
		return nil, fmt.Errorf("listing packages: %w", err)
	}
	defer rows.Close()

	var packages []InstalledPackage
	for rows.Next() {
		var pkg InstalledPackage
		if err := rows.Scan(&pkg.ID, &pkg.Name, &pkg.Version, &pkg.PackageManager,
			&pkg.Status, &pkg.InstalledBy, &pkg.InstallCommand, &pkg.UninstallCommand,
			&pkg.OutputLog, &pkg.ErrorLog, &pkg.Metadata,
			&pkg.CreatedAt, &pkg.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scanning package: %w", err)
		}
		packages = append(packages, pkg)
	}
	return packages, rows.Err()
}

// DeletePackageRecord removes a package tracking record.
func (s *Storage) DeletePackageRecord(id int64) error {
	_, err := s.db.Exec(`DELETE FROM installed_packages WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("deleting package record: %w", err)
	}
	return nil
}

// PackagesToJSON serializes a package list to JSON.
func PackagesToJSON(packages []InstalledPackage) string {
	data, err := json.MarshalIndent(packages, "", "  ")
	if err != nil {
		return fmt.Sprintf("error serializing packages: %v", err)
	}
	return string(data)
}

func scanPackage(row *sql.Row) (*InstalledPackage, error) {
	var pkg InstalledPackage
	err := row.Scan(&pkg.ID, &pkg.Name, &pkg.Version, &pkg.PackageManager,
		&pkg.Status, &pkg.InstalledBy, &pkg.InstallCommand, &pkg.UninstallCommand,
		&pkg.OutputLog, &pkg.ErrorLog, &pkg.Metadata,
		&pkg.CreatedAt, &pkg.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("scanning package: %w", err)
	}
	return &pkg, nil
}
