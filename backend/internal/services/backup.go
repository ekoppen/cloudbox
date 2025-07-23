package services

import (
	"archive/tar"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/cloudbox/backend/internal/models"
	"gorm.io/gorm"
)

// BackupService handles backup and restore operations
type BackupService struct {
	db        *gorm.DB
	backupDir string
}

// NewBackupService creates a new backup service
func NewBackupService(db *gorm.DB, backupDir string) *BackupService {
	// Ensure backup directory exists
	os.MkdirAll(backupDir, 0755)
	
	return &BackupService{
		db:        db,
		backupDir: backupDir,
	}
}

// BackupData represents the structure of backup data
type BackupData struct {
	Metadata         BackupMetadata             `json:"metadata"`
	Collections      []models.Collection        `json:"collections"`
	Documents        []models.Document          `json:"documents"`
	Files            []models.File              `json:"files"`
	Functions        []models.Function          `json:"functions"`
	Deployments      []models.Deployment        `json:"deployments"`
	GitHubRepos      []models.GitHubRepository  `json:"github_repositories"`
	WebServers       []models.WebServer         `json:"web_servers"`
	SSHKeys          []models.SSHKey            `json:"ssh_keys"`
	APIKeys          []models.APIKey            `json:"api_keys"`
	CORSConfigs      []models.CORSConfig        `json:"cors_configs"`
	FunctionDomains  []models.FunctionDomain    `json:"function_domains"`
	AuditLogs        []models.AuditLog          `json:"audit_logs"`
	BackupVersion    string                     `json:"backup_version"`
}

// BackupMetadata contains metadata about the backup
type BackupMetadata struct {
	ProjectID    uint      `json:"project_id"`
	ProjectName  string    `json:"project_name"`
	CreatedAt    time.Time `json:"created_at"`
	BackupType   string    `json:"backup_type"`
	CloudBoxVersion string `json:"cloudbox_version"`
}

// CreateBackup creates a full backup of a project
func (s *BackupService) CreateBackup(projectID uint, backupType string) (*models.Backup, error) {
	// Get project information
	var project models.Project
	if err := s.db.First(&project, projectID).Error; err != nil {
		return nil, fmt.Errorf("project not found: %w", err)
	}

	// Create backup record
	backup := models.Backup{
		Name:        fmt.Sprintf("%s-backup-%s", project.Name, time.Now().Format("2006-01-02-15-04-05")),
		Description: fmt.Sprintf("Full backup of project %s", project.Name),
		Type:        backupType,
		Status:      "creating",
		ProjectID:   projectID,
	}

	if err := s.db.Create(&backup).Error; err != nil {
		return nil, fmt.Errorf("failed to create backup record: %w", err)
	}

	// Perform backup asynchronously
	go s.performBackup(&backup, project)

	return &backup, nil
}

// performBackup performs the actual backup operation
func (s *BackupService) performBackup(backup *models.Backup, project models.Project) {
	// Create backup data structure
	backupData := BackupData{
		Metadata: BackupMetadata{
			ProjectID:       project.ID,
			ProjectName:     project.Name,
			CreatedAt:       time.Now(),
			BackupType:      backup.Type,
			CloudBoxVersion: "1.0.0", // Should come from config
		},
		BackupVersion: "1.0",
	}

	// Collect all project data
	if err := s.collectProjectData(project.ID, &backupData); err != nil {
		s.updateBackupStatus(backup.ID, "failed", fmt.Sprintf("Failed to collect data: %v", err))
		return
	}

	// Create backup file
	backupFilePath := filepath.Join(s.backupDir, fmt.Sprintf("backup-%d-%d.tar.gz", project.ID, backup.ID))
	size, checksum, err := s.createBackupArchive(backupData, backupFilePath)
	if err != nil {
		s.updateBackupStatus(backup.ID, "failed", fmt.Sprintf("Failed to create archive: %v", err))
		return
	}

	// Update backup record with completion info
	now := time.Now()
	s.db.Model(backup).Updates(map[string]interface{}{
		"status":       "completed",
		"size":         size,
		"file_path":    backupFilePath,
		"checksum":     checksum,
		"completed_at": &now,
	})
}

// collectProjectData collects all data for a project
func (s *BackupService) collectProjectData(projectID uint, backupData *BackupData) error {
	// Collect collections
	if err := s.db.Where("project_id = ?", projectID).Find(&backupData.Collections).Error; err != nil {
		return fmt.Errorf("failed to collect collections: %w", err)
	}

	// Collect documents
	if err := s.db.Where("project_id = ?", projectID).Find(&backupData.Documents).Error; err != nil {
		return fmt.Errorf("failed to collect documents: %w", err)
	}

	// Collect files
	if err := s.db.Where("project_id = ?", projectID).Find(&backupData.Files).Error; err != nil {
		return fmt.Errorf("failed to collect files: %w", err)
	}

	// Collect functions
	if err := s.db.Where("project_id = ?", projectID).Find(&backupData.Functions).Error; err != nil {
		return fmt.Errorf("failed to collect functions: %w", err)
	}

	// Collect deployments
	if err := s.db.Where("project_id = ?", projectID).Find(&backupData.Deployments).Error; err != nil {
		return fmt.Errorf("failed to collect deployments: %w", err)
	}

	// Collect GitHub repositories
	if err := s.db.Where("project_id = ?", projectID).Find(&backupData.GitHubRepos).Error; err != nil {
		return fmt.Errorf("failed to collect github repos: %w", err)
	}

	// Collect web servers
	if err := s.db.Where("project_id = ?", projectID).Find(&backupData.WebServers).Error; err != nil {
		return fmt.Errorf("failed to collect web servers: %w", err)
	}

	// Collect SSH keys
	if err := s.db.Where("project_id = ?", projectID).Find(&backupData.SSHKeys).Error; err != nil {
		return fmt.Errorf("failed to collect ssh keys: %w", err)
	}

	// Collect API keys
	if err := s.db.Where("project_id = ?", projectID).Find(&backupData.APIKeys).Error; err != nil {
		return fmt.Errorf("failed to collect api keys: %w", err)
	}

	// Collect CORS configs
	if err := s.db.Where("project_id = ?", projectID).Find(&backupData.CORSConfigs).Error; err != nil {
		return fmt.Errorf("failed to collect cors configs: %w", err)
	}

	// Collect function domains
	if err := s.db.Where("project_id = ?", projectID).Find(&backupData.FunctionDomains).Error; err != nil {
		return fmt.Errorf("failed to collect function domains: %w", err)
	}

	// Collect audit logs (recent ones only to avoid huge backups)
	since := time.Now().AddDate(0, -3, 0) // Last 3 months
	if err := s.db.Where("project_id = ? AND created_at > ?", projectID, since).Find(&backupData.AuditLogs).Error; err != nil {
		return fmt.Errorf("failed to collect audit logs: %w", err)
	}

	return nil
}

// createBackupArchive creates a compressed tar archive of the backup data
func (s *BackupService) createBackupArchive(backupData BackupData, filePath string) (int64, string, error) {
	// Create file
	file, err := os.Create(filePath)
	if err != nil {
		return 0, "", fmt.Errorf("failed to create backup file: %w", err)
	}
	defer file.Close()

	// Create gzip writer
	gzipWriter := gzip.NewWriter(file)
	defer gzipWriter.Close()

	// Create tar writer
	tarWriter := tar.NewWriter(gzipWriter)
	defer tarWriter.Close()

	// Convert backup data to JSON
	jsonData, err := json.MarshalIndent(backupData, "", "  ")
	if err != nil {
		return 0, "", fmt.Errorf("failed to marshal backup data: %w", err)
	}

	// Add backup data to archive
	header := &tar.Header{
		Name: "backup.json",
		Mode: 0644,
		Size: int64(len(jsonData)),
	}

	if err := tarWriter.WriteHeader(header); err != nil {
		return 0, "", fmt.Errorf("failed to write tar header: %w", err)
	}

	if _, err := tarWriter.Write(jsonData); err != nil {
		return 0, "", fmt.Errorf("failed to write backup data: %w", err)
	}

	// Close writers to ensure all data is written
	tarWriter.Close()
	gzipWriter.Close()

	// Get file info
	fileInfo, err := file.Stat()
	if err != nil {
		return 0, "", fmt.Errorf("failed to get file info: %w", err)
	}

	// Calculate checksum
	file.Seek(0, 0)
	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return 0, "", fmt.Errorf("failed to calculate checksum: %w", err)
	}

	checksum := hex.EncodeToString(hash.Sum(nil))
	return fileInfo.Size(), checksum, nil
}

// RestoreBackup restores a project from a backup
func (s *BackupService) RestoreBackup(backupID uint, targetProjectID uint) error {
	// Get backup record
	var backup models.Backup
	if err := s.db.First(&backup, backupID).Error; err != nil {
		return fmt.Errorf("backup not found: %w", err)
	}

	if backup.Status != "completed" {
		return fmt.Errorf("backup is not completed, cannot restore")
	}

	// Extract and parse backup data
	backupData, err := s.extractBackupData(backup.FilePath)
	if err != nil {
		return fmt.Errorf("failed to extract backup data: %w", err)
	}

	// Begin transaction for atomic restore
	return s.db.Transaction(func(tx *gorm.DB) error {
		// Clear existing project data (if replacing)
		if err := s.clearProjectData(tx, targetProjectID); err != nil {
			return fmt.Errorf("failed to clear existing data: %w", err)
		}

		// Restore data
		if err := s.restoreProjectData(tx, targetProjectID, backupData); err != nil {
			return fmt.Errorf("failed to restore data: %w", err)
		}

		return nil
	})
}

// extractBackupData extracts and parses backup data from archive
func (s *BackupService) extractBackupData(filePath string) (*BackupData, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open backup file: %w", err)
	}
	defer file.Close()

	gzipReader, err := gzip.NewReader(file)
	if err != nil {
		return nil, fmt.Errorf("failed to create gzip reader: %w", err)
	}
	defer gzipReader.Close()

	tarReader := tar.NewReader(gzipReader)

	// Find backup.json file
	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("error reading tar: %w", err)
		}

		if header.Name == "backup.json" {
			var backupData BackupData
			if err := json.NewDecoder(tarReader).Decode(&backupData); err != nil {
				return nil, fmt.Errorf("failed to decode backup data: %w", err)
			}
			return &backupData, nil
		}
	}

	return nil, fmt.Errorf("backup.json not found in archive")
}

// clearProjectData removes existing project data (for full restore)
func (s *BackupService) clearProjectData(tx *gorm.DB, projectID uint) error {
	// Clear in reverse dependency order
	models := []interface{}{
		&models.AuditLog{},
		&models.FunctionDomain{},
		&models.CORSConfig{},
		&models.APIKey{},
		&models.SSHKey{},
		&models.WebServer{},
		&models.GitHubRepository{},
		&models.Deployment{},
		&models.Function{},
		&models.File{},
		&models.Document{},
		&models.Collection{},
	}

	for _, model := range models {
		if err := tx.Where("project_id = ?", projectID).Delete(model).Error; err != nil {
			return fmt.Errorf("failed to clear %T: %w", model, err)
		}
	}

	return nil
}

// restoreProjectData restores data to target project
func (s *BackupService) restoreProjectData(tx *gorm.DB, projectID uint, backupData *BackupData) error {
	// Update project IDs in backup data
	s.updateProjectIDs(backupData, projectID)

	// Restore in dependency order
	if len(backupData.Collections) > 0 {
		if err := tx.Create(&backupData.Collections).Error; err != nil {
			return fmt.Errorf("failed to restore collections: %w", err)
		}
	}

	if len(backupData.Documents) > 0 {
		if err := tx.Create(&backupData.Documents).Error; err != nil {
			return fmt.Errorf("failed to restore documents: %w", err)
		}
	}

	if len(backupData.Files) > 0 {
		if err := tx.Create(&backupData.Files).Error; err != nil {
			return fmt.Errorf("failed to restore files: %w", err)
		}
	}

	if len(backupData.Functions) > 0 {
		if err := tx.Create(&backupData.Functions).Error; err != nil {
			return fmt.Errorf("failed to restore functions: %w", err)
		}
	}

	if len(backupData.Deployments) > 0 {
		if err := tx.Create(&backupData.Deployments).Error; err != nil {
			return fmt.Errorf("failed to restore deployments: %w", err)
		}
	}

	if len(backupData.GitHubRepos) > 0 {
		if err := tx.Create(&backupData.GitHubRepos).Error; err != nil {
			return fmt.Errorf("failed to restore github repos: %w", err)
		}
	}

	if len(backupData.WebServers) > 0 {
		if err := tx.Create(&backupData.WebServers).Error; err != nil {
			return fmt.Errorf("failed to restore web servers: %w", err)
		}
	}

	if len(backupData.SSHKeys) > 0 {
		if err := tx.Create(&backupData.SSHKeys).Error; err != nil {
			return fmt.Errorf("failed to restore ssh keys: %w", err)
		}
	}

	if len(backupData.APIKeys) > 0 {
		if err := tx.Create(&backupData.APIKeys).Error; err != nil {
			return fmt.Errorf("failed to restore api keys: %w", err)
		}
	}

	if len(backupData.CORSConfigs) > 0 {
		if err := tx.Create(&backupData.CORSConfigs).Error; err != nil {
			return fmt.Errorf("failed to restore cors configs: %w", err)
		}
	}

	if len(backupData.FunctionDomains) > 0 {
		if err := tx.Create(&backupData.FunctionDomains).Error; err != nil {
			return fmt.Errorf("failed to restore function domains: %w", err)
		}
	}

	if len(backupData.AuditLogs) > 0 {
		if err := tx.Create(&backupData.AuditLogs).Error; err != nil {
			return fmt.Errorf("failed to restore audit logs: %w", err)
		}
	}

	return nil
}

// updateProjectIDs updates all project IDs in backup data
func (s *BackupService) updateProjectIDs(backupData *BackupData, newProjectID uint) {
	for i := range backupData.Collections {
		backupData.Collections[i].ProjectID = newProjectID
		backupData.Collections[i].ID = 0 // Reset ID for new creation
	}
	
	for i := range backupData.Documents {
		backupData.Documents[i].ProjectID = newProjectID
		backupData.Documents[i].ID = 0
	}
	
	for i := range backupData.Files {
		backupData.Files[i].ProjectID = newProjectID
		backupData.Files[i].ID = 0
	}
	
	for i := range backupData.Functions {
		backupData.Functions[i].ProjectID = newProjectID
		backupData.Functions[i].ID = 0
	}
	
	for i := range backupData.Deployments {
		backupData.Deployments[i].ProjectID = newProjectID
		backupData.Deployments[i].ID = 0
	}
	
	for i := range backupData.GitHubRepos {
		backupData.GitHubRepos[i].ProjectID = newProjectID
		backupData.GitHubRepos[i].ID = 0
	}
	
	for i := range backupData.WebServers {
		backupData.WebServers[i].ProjectID = newProjectID
		backupData.WebServers[i].ID = 0
	}
	
	for i := range backupData.SSHKeys {
		backupData.SSHKeys[i].ProjectID = newProjectID
		backupData.SSHKeys[i].ID = 0
	}
	
	for i := range backupData.APIKeys {
		backupData.APIKeys[i].ProjectID = newProjectID
		backupData.APIKeys[i].ID = 0
	}
	
	for i := range backupData.CORSConfigs {
		backupData.CORSConfigs[i].ProjectID = newProjectID
		backupData.CORSConfigs[i].ID = 0
	}
	
	for i := range backupData.FunctionDomains {
		backupData.FunctionDomains[i].ProjectID = newProjectID
		backupData.FunctionDomains[i].ID = 0
	}
	
	for i := range backupData.AuditLogs {
		backupData.AuditLogs[i].ProjectID = newProjectID
		backupData.AuditLogs[i].ID = 0
	}
}

// updateBackupStatus updates backup status with error message
func (s *BackupService) updateBackupStatus(backupID uint, status string, message string) {
	s.db.Model(&models.Backup{}).Where("id = ?", backupID).Updates(map[string]interface{}{
		"status": status,
	})
}

// ListBackups returns all backups for a project
func (s *BackupService) ListBackups(projectID uint) ([]models.Backup, error) {
	var backups []models.Backup
	err := s.db.Where("project_id = ?", projectID).Order("created_at DESC").Find(&backups).Error
	return backups, err
}

// GetBackup returns a specific backup
func (s *BackupService) GetBackup(backupID uint) (*models.Backup, error) {
	var backup models.Backup
	err := s.db.Preload("Project").First(&backup, backupID).Error
	if err != nil {
		return nil, err
	}
	return &backup, nil
}

// DeleteBackup deletes a backup and its file
func (s *BackupService) DeleteBackup(backupID uint) error {
	var backup models.Backup
	if err := s.db.First(&backup, backupID).Error; err != nil {
		return fmt.Errorf("backup not found: %w", err)
	}

	// Delete backup file if it exists
	if backup.FilePath != "" {
		if err := os.Remove(backup.FilePath); err != nil && !os.IsNotExist(err) {
			// Log error but don't fail the operation
			fmt.Printf("Warning: failed to delete backup file %s: %v\n", backup.FilePath, err)
		}
	}

	// Delete backup record
	return s.db.Delete(&backup).Error
}