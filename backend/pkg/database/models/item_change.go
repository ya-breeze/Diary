package models

import (
	"time"

	openapi_types "github.com/oapi-codegen/runtime/types"
	"github.com/ya-breeze/diary.be/pkg/generated/goserver"
)

// OperationType represents the type of operation performed on an item
type OperationType string

const (
	OperationTypeCreated OperationType = "created"
	OperationTypeUpdated OperationType = "updated"
	OperationTypeDeleted OperationType = "deleted"
)

// ItemChange tracks changes to diary items for synchronization purposes
type ItemChange struct {
	// ID is the auto-incrementing primary key for change tracking
	ID uint `gorm:"primaryKey;autoIncrement" json:"id"`

	// UserID identifies which user's data was changed
	UserID string `gorm:"index;not null" json:"userId"`

	// Date is the date identifier of the item that was modified
	Date string `gorm:"index;not null" json:"date"`

	// OperationType indicates what kind of operation was performed
	OperationType OperationType `gorm:"type:varchar(10);not null" json:"operationType"`

	// Timestamp records when the change occurred
	Timestamp time.Time `gorm:"index;not null;default:CURRENT_TIMESTAMP" json:"timestamp"`

	// ItemSnapshot contains the current state of the item after the change
	// For deleted items, this contains the last known state before deletion
	ItemSnapshot *Item `gorm:"embedded;embeddedPrefix:item_" json:"itemSnapshot,omitempty"`

	// Metadata stores additional information about the change
	Metadata StringList `gorm:"type:json" json:"metadata,omitempty"`
}

// ToSyncResponse converts ItemChange to the API response format
func (ic ItemChange) ToSyncResponse() goserver.SyncChangeResponse {
	var id int32
	if ic.ID <= uint(^uint32(0)>>1) { // Check if it fits in int32 (max positive value)
		id = int32(ic.ID) // #nosec G115 - checked above
	}

	date := openapi_types.Date{Time: mustParseDate(ic.Date)}
	metadata := []string(ic.Metadata)
	response := goserver.SyncChangeResponse{
		Id:            id,
		UserId:        ic.UserID,
		Date:          date,
		OperationType: goserver.SyncChangeResponseOperationType(ic.OperationType),
		Timestamp:     ic.Timestamp,
		Metadata:      &metadata,
	}

	// Include item data for all operations (including deleted items to show what was deleted)
	if ic.ItemSnapshot != nil {
		snapshotDate := openapi_types.Date{Time: mustParseDate(ic.ItemSnapshot.Date)}
		body := ic.ItemSnapshot.Body
		tags := []string(ic.ItemSnapshot.Tags)
		response.ItemSnapshot = &goserver.ItemsResponse{
			Date:  snapshotDate,
			Title: ic.ItemSnapshot.Title,
			Body:  &body,
			Tags:  &tags,
		}
	}

	return response
}

// mustParseDate parses a "2006-01-02" date string; returns zero time on error.
func mustParseDate(s string) time.Time {
	t, _ := time.Parse("2006-01-02", s)
	return t
}
