package models

import (
	"time"

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

	response := goserver.SyncChangeResponse{
		Id:            id,
		UserId:        ic.UserID,
		Date:          ic.Date,
		OperationType: string(ic.OperationType),
		Timestamp:     ic.Timestamp,
		Metadata:      []string(ic.Metadata),
	}

	// Include item data for all operations (including deleted items to show what was deleted)
	if ic.ItemSnapshot != nil {
		response.ItemSnapshot = &goserver.ItemsResponse{
			Date:  ic.ItemSnapshot.Date,
			Title: ic.ItemSnapshot.Title,
			Body:  ic.ItemSnapshot.Body,
			Tags:  []string(ic.ItemSnapshot.Tags),
		}
	}

	return response
}
