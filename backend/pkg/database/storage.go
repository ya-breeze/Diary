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

	GetUserID(username string) (string, error)
	GetUser(userID string) (*models.User, error)
	CreateUser(username, password string) (*models.User, error)
	PutUser(user *models.User) error

	GetItem(userID, itemID string) (*models.Item, error)
	GetItems(userID string, searchParams SearchParams) ([]*models.Item, int, error)
	PutItem(userID string, item *models.Item) error
	DeleteItem(userID, itemID string) error

	GetPreviousDate(userID, date string) (string, error)
	GetNextDate(userID, date string) (string, error)

	// Change tracking methods for synchronization
	CreateChangeRecord(userID, date string, operationType models.OperationType,
		itemSnapshot *models.Item, metadata []string) error
	GetChangesSince(userID string, sinceID uint, limit int) ([]*models.ItemChange, error)
	GetLatestChangeID(userID string) (uint, error)
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
	if err := autoMigrateModels(s.db); err != nil {
		s.log.Error("failed to migrate database", "error", err)
		panic("failed to migrate database")
	}

	return nil
}

func (s *storage) Close() error {
	// return s.db.Close()
	return nil
}

func (s *storage) CreateUser(username, hashedPassword string) (*models.User, error) {
	_, err := s.GetUserID(username)
	if err == nil {
		s.log.Error("user already exists", "username", username)
		return nil, fmt.Errorf("user %q already exists", username)
	}

	user := models.User{
		ID:             uuid.New(),
		Login:          username,
		HashedPassword: hashedPassword,
		StartDate:      time.Now(),
	}
	if err := s.db.Create(&user).Error; err != nil {
		return nil, fmt.Errorf(StorageError, err)
	}

	return &user, nil
}

func (s *storage) GetUser(userID string) (*models.User, error) {
	var user models.User
	if err := s.db.Where("id = ?", userID).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}

		return nil, fmt.Errorf(StorageError, err)
	}

	return &user, nil
}

func (s *storage) PutUser(user *models.User) error {
	existingUserID, err := s.GetUserID(user.Login)
	if err != nil {
		s.log.Error("failed to get user ID", "error", err, "user", user.Login)
		return fmt.Errorf("failed to get user ID: %w", err)
	}
	if existingUserID != user.ID.String() {
		s.log.Error("user ID mismatch", "expected", user.ID.String(), "actual", existingUserID)
		return fmt.Errorf("user ID mismatch: expected %s, actual %s", user.ID.String(), existingUserID)
	}

	// Update the user in the database
	if err := s.db.Save(user).Error; err != nil {
		return fmt.Errorf(StorageError, err)
	}

	return nil
}

func (s *storage) GetUserID(username string) (string, error) {
	var user models.User
	if err := s.db.Where("login = ?", username).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", ErrNotFound
		}

		return "", fmt.Errorf(StorageError, err)
	}

	return user.ID.String(), nil
}

// #region Item

func (s *storage) GetItem(userID, date string) (*models.Item, error) {
	var item models.Item
	if err := s.db.Where("date = ? and user_id = ?", date, userID).First(&item).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}

		return nil, fmt.Errorf(StorageError, err)
	}

	return &item, nil
}

func (s *storage) GetItems(userID string, searchParams SearchParams) ([]*models.Item, int, error) {
	var items []*models.Item
	query := s.db.Where("user_id = ?", userID)

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

func (s *storage) PutItem(userID string, item *models.Item) error {
	item.UserID = userID

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
	isUpdate := tx.Where("user_id = ? AND date = ?", userID, item.Date).First(&existingItem).Error == nil

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

	if err := s.createChangeRecordInTx(tx, userID, item.Date, operationType, item, nil); err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to create change record: %w", err)
	}

	// Commit the transaction
	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf(StorageError, err)
	}

	return nil
}

func (s *storage) DeleteItem(userID, itemID string) error {
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
	if err := tx.Where("user_id = ? AND date = ?", userID, itemID).First(&item).Error; err != nil {
		tx.Rollback()
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrNotFound
		}
		return fmt.Errorf(StorageError, err)
	}

	// Delete the item
	if err := tx.Where("user_id = ? AND date = ?", userID, itemID).Delete(&models.Item{}).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf(StorageError, err)
	}

	// Create change record for deletion
	if err := s.createChangeRecordInTx(tx, userID, itemID, models.OperationTypeDeleted, &item, nil); err != nil {
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

func (s *storage) GetPreviousDate(userID, date string) (string, error) {
	var item models.Item
	if err := s.db.Where("user_id = ? and date < ?", userID, date).Order("date desc").First(&item).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", ErrNotFound
		}

		return "", fmt.Errorf(StorageError, err)
	}

	return item.Date, nil
}

func (s *storage) GetNextDate(userID, date string) (string, error) {
	var item models.Item
	if err := s.db.Where("user_id = ? and date > ?", userID, date).Order("date asc").First(&item).Error; err != nil {
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
func (s *storage) createChangeRecordInTx(tx *gorm.DB, userID, date string,
	operationType models.OperationType, itemSnapshot *models.Item, metadata []string,
) error {
	change := &models.ItemChange{
		UserID:        userID,
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
func (s *storage) CreateChangeRecord(userID, date string, operationType models.OperationType,
	itemSnapshot *models.Item, metadata []string,
) error {
	change := &models.ItemChange{
		UserID:        userID,
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

// GetChangesSince retrieves changes for a user since a given change ID
func (s *storage) GetChangesSince(userID string, sinceID uint, limit int) ([]*models.ItemChange, error) {
	var changes []*models.ItemChange

	query := s.db.Where("user_id = ? AND id > ?", userID, sinceID).
		Order("id ASC").
		Limit(limit)

	if err := query.Find(&changes).Error; err != nil {
		return nil, fmt.Errorf(StorageError, err)
	}

	return changes, nil
}

// GetLatestChangeID returns the latest change ID for a user
func (s *storage) GetLatestChangeID(userID string) (uint, error) {
	var change models.ItemChange

	err := s.db.Where("user_id = ?", userID).
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
