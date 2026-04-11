package database

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/ya-breeze/diary.be/pkg/config"
	"github.com/ya-breeze/diary.be/pkg/database/models"
	coremodels "github.com/ya-breeze/kin-core/models"
	"gorm.io/gorm"
)

//go:generate go tool github.com/golang/mock/mockgen -destination=mocks/mock_storage.go -package=mocks github.com/ya-breeze/diary.be/pkg/database Storage //nolint:lll // go:generate directive

const StorageError = "storage error: %w"

var ErrNotFound = errors.New("not found")

// SearchParams defines parameters for searching diary items
type SearchParams struct {
	// SearchText filters items by title and body content (case-insensitive)
	SearchText string
	// Tags filters items that contain any of the specified tags
	Tags []string
	// Date filters items by specific date (optional, for backward compatibility)
	Date string
}

//nolint:interfacebloat // keep a single storage interface for simplicity
type Storage interface {
	Open() error
	Close() error

	GetUserByUsername(username string) (*models.User, error)
	GetUser(userID uuid.UUID) (*models.User, error)
	GetAllUsers() ([]*models.User, error)
	CreateUser(username, password string, familyID uuid.UUID) (*models.User, error)
	PutUser(user *models.User) error

	GetFamilyByName(name string) (*models.Family, error)
	CreateFamily(name string) (*models.Family, error)

	GetItem(familyID uuid.UUID, date string) (*models.Item, error)
	GetItems(familyID uuid.UUID, searchParams SearchParams) ([]*models.Item, int, error)
	PutItem(familyID uuid.UUID, item *models.Item) error
	DeleteItem(familyID uuid.UUID, date string) error

	GetPreviousDate(familyID uuid.UUID, date string) (string, error)
	GetNextDate(familyID uuid.UUID, date string) (string, error)

	// Change tracking methods for synchronization
	CreateChangeRecord(familyID uuid.UUID, date string, operationType models.OperationType,
		itemSnapshot *models.Item, metadata []string) error
	GetChangesSince(familyID uuid.UUID, sinceID uint, limit int) ([]*models.ItemChange, error)
	GetLatestChangeID(familyID uuid.UUID) (uint, error)

	// Orphan ignore list
	GetIgnoredOrphans(familyID uuid.UUID) ([]string, error)
	AddIgnoredOrphan(familyID uuid.UUID, filename string) error
	RemoveIgnoredOrphan(familyID uuid.UUID, filename string) error

	// GetDB returns the underlying gorm.DB for use with authdb helpers.
	GetDB() *gorm.DB
}

type storage struct {
	log *slog.Logger
	cfg *config.Config
	db  *gorm.DB
}

func NewStorage(logger *slog.Logger, cfg *config.Config) Storage {
	return &storage{log: logger, db: nil, cfg: cfg}
}

func (s *storage) Open() error {
	dbPath := filepath.Join(s.cfg.DataPath, config.DBFilename)
	s.log.Info("Opening database", "path", dbPath)
	var err error
	s.db, err = openSqlite(s.log, dbPath, s.cfg.Verbose)
	if err != nil {
		if strings.Contains(err.Error(), "no such file or directory") {
			// Check if the directory exists
			dir := filepath.Dir(dbPath)
			if _, statErr := os.Stat(dir); os.IsNotExist(statErr) {
				s.log.Error("database directory does not exist", "dir", dir)
			} else {
				s.log.Error("database file not found in directory", "path", dbPath)
				// List files in the directory to help debugging
				if files, readErr := os.ReadDir(dir); readErr == nil {
					var fileNames []string
					for i, f := range files {
						if i >= 5 {
							break
						}
						fileNames = append(fileNames, f.Name())
					}
					s.log.Info("files in database directory", "files", fileNames)
				}
			}
		}
		s.log.Error("failed to connect database", "error", err)
		panic("failed to connect database")
	}

	if err := runMigrationIfNeeded(s.log, s.db, s.cfg); err != nil {
		s.log.Error("failed to run migration", "error", err)
		panic("failed to run migration")
	}

	if err := autoMigrateModels(s.db); err != nil {
		s.log.Error("failed to migrate database", "error", err)
		panic("failed to migrate database")
	}

	return nil
}

func (s *storage) Close() error {
	return nil
}

func (s *storage) GetDB() *gorm.DB {
	return s.db
}

// #region User

func (s *storage) CreateUser(username, hashedPassword string, familyID uuid.UUID) (*models.User, error) {
	_, err := s.GetUserByUsername(username)
	if err == nil {
		s.log.Error("user already exists", "username", username)
		return nil, fmt.Errorf("user %q already exists", username)
	}

	user := models.User{
		User: coremodels.User{
			ID:           uuid.New(),
			Username:     username,
			PasswordHash: hashedPassword,
			FamilyID:     familyID,
		},
		StartDate: time.Now(),
	}
	if err := s.db.Create(&user).Error; err != nil {
		return nil, fmt.Errorf(StorageError, err)
	}

	return &user, nil
}

func (s *storage) GetAllUsers() ([]*models.User, error) {
	var users []*models.User
	if err := s.db.Find(&users).Error; err != nil {
		return nil, fmt.Errorf(StorageError, err)
	}
	return users, nil
}

func (s *storage) GetUser(userID uuid.UUID) (*models.User, error) {
	var user models.User
	if err := s.db.Where("id = ?", userID).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf(StorageError, err)
	}
	return &user, nil
}

func (s *storage) GetUserByUsername(username string) (*models.User, error) {
	var user models.User
	if err := s.db.Where("username = ?", username).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf(StorageError, err)
	}
	return &user, nil
}

func (s *storage) PutUser(user *models.User) error {
	if err := s.db.Save(user).Error; err != nil {
		return fmt.Errorf(StorageError, err)
	}
	return nil
}

// #endregion User

// #region Family

func (s *storage) GetFamilyByName(name string) (*models.Family, error) {
	var family models.Family
	if err := s.db.Where("name = ?", name).First(&family).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf(StorageError, err)
	}
	return &family, nil
}

func (s *storage) CreateFamily(name string) (*models.Family, error) {
	family := models.Family{
		Family: coremodels.Family{
			ID:   uuid.New(),
			Name: name,
		},
	}
	if err := s.db.Create(&family).Error; err != nil {
		return nil, fmt.Errorf(StorageError, err)
	}
	return &family, nil
}

// #endregion Family

// #region Item

func (s *storage) GetItem(familyID uuid.UUID, date string) (*models.Item, error) {
	var item models.Item
	if err := s.db.Where("date = ? AND family_id = ?", date, familyID).First(&item).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf(StorageError, err)
	}
	return &item, nil
}

func (s *storage) GetItems(familyID uuid.UUID, searchParams SearchParams) ([]*models.Item, int, error) {
	var items []*models.Item
	query := s.db.Where("family_id = ?", familyID)

	// Apply date filter if specified (for backward compatibility)
	if searchParams.Date != "" {
		query = query.Where("date = ?", searchParams.Date)
	}

	// Apply text search filter if specified
	if searchParams.SearchText != "" {
		searchPattern := "%" + searchParams.SearchText + "%"
		query = query.Where("title LIKE ? OR body LIKE ?", searchPattern, searchPattern)
	}

	// Apply tag filters if specified
	if len(searchParams.Tags) > 0 {
		// For JSON tag filtering, we need to check if any of the specified tags exist in the JSON array
		tagConditions := make([]string, len(searchParams.Tags))
		tagArgs := make([]any, len(searchParams.Tags))
		for i, tag := range searchParams.Tags {
			tagConditions[i] = "JSON_EXTRACT(tags, '$') LIKE ?"
			tagArgs[i] = "%\"" + tag + "\"%"
		}
		tagQuery := strings.Join(tagConditions, " OR ")
		query = query.Where(tagQuery, tagArgs...)
	}

	// Get total count for pagination
	var totalCount int64
	if err := query.Model(&models.Item{}).Count(&totalCount).Error; err != nil {
		return nil, 0, fmt.Errorf(StorageError, err)
	}

	// Execute the query to get items, ordered by date descending
	if err := query.Order("date DESC").Find(&items).Error; err != nil {
		return nil, 0, fmt.Errorf(StorageError, err)
	}

	return items, int(totalCount), nil
}

func (s *storage) PutItem(familyID uuid.UUID, item *models.Item) error {
	item.FamilyID = familyID

	// Start a transaction to ensure atomicity
	tx := s.db.Begin()
	if tx.Error != nil {
		return fmt.Errorf(StorageError, tx.Error)
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Check if item exists to determine operation type
	var existingItem models.Item
	isUpdate := tx.Where("family_id = ? AND date = ?", familyID, item.Date).First(&existingItem).Error == nil

	// Preserve existing ID on update
	if isUpdate {
		item.ID = existingItem.ID
	} else {
		item.ID = uuid.New()
	}

	// Save the item
	if err := tx.Save(item).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf(StorageError, err)
	}

	// Create change record
	operationType := models.OperationTypeCreated
	if isUpdate {
		operationType = models.OperationTypeUpdated
	}

	if err := s.createChangeRecordInTx(tx, familyID, item.Date, operationType, item, nil); err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to create change record: %w", err)
	}

	// Commit the transaction
	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf(StorageError, err)
	}

	return nil
}

func (s *storage) DeleteItem(familyID uuid.UUID, date string) error {
	// Start a transaction to ensure atomicity
	tx := s.db.Begin()
	if tx.Error != nil {
		return fmt.Errorf(StorageError, tx.Error)
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Get the item before deletion for the change record
	var item models.Item
	if err := tx.Where("family_id = ? AND date = ?", familyID, date).First(&item).Error; err != nil {
		tx.Rollback()
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrNotFound
		}
		return fmt.Errorf(StorageError, err)
	}

	// Delete the item
	if err := tx.Where("family_id = ? AND date = ?", familyID, date).Delete(&models.Item{}).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf(StorageError, err)
	}

	// Create change record for deletion
	if err := s.createChangeRecordInTx(tx, familyID, date, models.OperationTypeDeleted, &item, nil); err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to create change record: %w", err)
	}

	// Commit the transaction
	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf(StorageError, err)
	}

	return nil
}

// #endregion Item

// #region Dates

func (s *storage) GetPreviousDate(familyID uuid.UUID, date string) (string, error) {
	var item models.Item
	if err := s.db.Where("family_id = ? AND date < ?", familyID, date).Order("date desc").First(&item).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", ErrNotFound
		}
		return "", fmt.Errorf(StorageError, err)
	}
	return item.Date, nil
}

func (s *storage) GetNextDate(familyID uuid.UUID, date string) (string, error) {
	var item models.Item
	if err := s.db.Where("family_id = ? AND date > ?", familyID, date).Order("date asc").First(&item).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", ErrNotFound
		}
		return "", fmt.Errorf(StorageError, err)
	}
	return item.Date, nil
}

// #endregion Dates

// #region Change Tracking

// createChangeRecordInTx creates a change record within an existing transaction
func (s *storage) createChangeRecordInTx(tx *gorm.DB, familyID uuid.UUID, date string,
	operationType models.OperationType, itemSnapshot *models.Item, metadata []string,
) error {
	change := &models.ItemChange{
		FamilyID:      familyID,
		Date:          date,
		OperationType: operationType,
		Timestamp:     time.Now(),
		ItemSnapshot:  itemSnapshot,
		Metadata:      models.StringList(metadata),
	}

	if err := tx.Create(change).Error; err != nil {
		return fmt.Errorf(StorageError, err)
	}

	return nil
}

// CreateChangeRecord creates a change record for synchronization
func (s *storage) CreateChangeRecord(familyID uuid.UUID, date string, operationType models.OperationType,
	itemSnapshot *models.Item, metadata []string,
) error {
	change := &models.ItemChange{
		FamilyID:      familyID,
		Date:          date,
		OperationType: operationType,
		Timestamp:     time.Now(),
		ItemSnapshot:  itemSnapshot,
		Metadata:      models.StringList(metadata),
	}

	if err := s.db.Create(change).Error; err != nil {
		return fmt.Errorf(StorageError, err)
	}

	return nil
}

// GetChangesSince retrieves changes for a family since a given change ID
func (s *storage) GetChangesSince(familyID uuid.UUID, sinceID uint, limit int) ([]*models.ItemChange, error) {
	var changes []*models.ItemChange

	query := s.db.Where("family_id = ? AND id > ?", familyID, sinceID).
		Order("id ASC").
		Limit(limit)

	if err := query.Find(&changes).Error; err != nil {
		return nil, fmt.Errorf(StorageError, err)
	}

	return changes, nil
}

// GetLatestChangeID returns the latest change ID for a family
func (s *storage) GetLatestChangeID(familyID uuid.UUID) (uint, error) {
	var change models.ItemChange

	err := s.db.Where("family_id = ?", familyID).
		Order("id DESC").
		First(&change).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return 0, nil // No changes yet
		}
		return 0, fmt.Errorf(StorageError, err)
	}

	return change.ID, nil
}

// #endregion Change Tracking

// #region Orphan Ignore

func (s *storage) GetIgnoredOrphans(familyID uuid.UUID) ([]string, error) {
	var records []models.OrphanIgnore
	if err := s.db.Where("family_id = ?", familyID).Find(&records).Error; err != nil {
		return nil, fmt.Errorf(StorageError, err)
	}
	filenames := make([]string, len(records))
	for i, r := range records {
		filenames[i] = r.Filename
	}
	return filenames, nil
}

func (s *storage) AddIgnoredOrphan(familyID uuid.UUID, filename string) error {
	record := models.OrphanIgnore{FamilyID: familyID, Filename: filename}
	if err := s.db.Where(models.OrphanIgnore{FamilyID: familyID, Filename: filename}).
		FirstOrCreate(&record).Error; err != nil {
		return fmt.Errorf(StorageError, err)
	}
	return nil
}

func (s *storage) RemoveIgnoredOrphan(familyID uuid.UUID, filename string) error {
	if err := s.db.Where("family_id = ? AND filename = ?", familyID, filename).
		Delete(&models.OrphanIgnore{}).Error; err != nil {
		return fmt.Errorf(StorageError, err)
	}
	return nil
}

// #endregion Orphan Ignore
