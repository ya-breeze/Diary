package common

import (
	"context"

	"github.com/google/uuid"
)

// GetFamilyID extracts the familyID from the request context.
func GetFamilyID(ctx context.Context) (uuid.UUID, bool) {
	v := ctx.Value(FamilyIDKey)
	id, ok := v.(uuid.UUID)
	return id, ok
}
