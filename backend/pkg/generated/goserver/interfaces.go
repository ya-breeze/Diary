// Hand-written service interfaces.  These define the contracts that custom
// service implementations must satisfy.  The StrictServerImpl adapter (in
// adapter.go) bridges from these per-service interfaces to the oapi-codegen
// generated StrictServerInterface.
package goserver

import (
	"context"
	"os"
)

// AssetsAPIService defines the business logic for the Assets API.
type AssetsAPIService interface {
	GetAsset(ctx context.Context, path string) (ImplResponse, error)
	UploadAssetsBatch(ctx context.Context, files []*os.File) (ImplResponse, error)
}

// AuthAPIService defines the business logic for the Auth API.
type AuthAPIService interface {
	Authorize(ctx context.Context, authData AuthData) (ImplResponse, error)
}

// AuthAPIServicer is an alias for backward compatibility with custom controllers.
type AuthAPIServicer = AuthAPIService

// HealthAPIService defines the business logic for the Health API.
type HealthAPIService interface {
	GetHealthIssues(ctx context.Context) (ImplResponse, error)
	FixHealthIssues(ctx context.Context, req HealthFixRequest) (ImplResponse, error)
	DeleteOrphan(ctx context.Context, filename string) (ImplResponse, error)
	AttachOrphan(ctx context.Context, filename string, req AttachOrphanRequest) (ImplResponse, error)
	IgnoreOrphan(ctx context.Context, filename string) (ImplResponse, error)
	UnignoreOrphan(ctx context.Context, filename string) (ImplResponse, error)
}

// ItemsAPIService defines the business logic for the Items API.
type ItemsAPIService interface {
	GetItems(ctx context.Context, date string, search string, tags string) (ImplResponse, error)
	PutItems(ctx context.Context, itemsRequest ItemsRequest) (ImplResponse, error)
	SuggestItemTags(ctx context.Context, req SuggestTagsRequest) (ImplResponse, error)
	DismissItemTag(ctx context.Context, req DismissTagRequest) (ImplResponse, error)
	AcceptItemTag(ctx context.Context, req DismissTagRequest) (ImplResponse, error)
	GetTags(ctx context.Context) (ImplResponse, error)
	GetTagStats(ctx context.Context) (ImplResponse, error)
	RenameTag(ctx context.Context, name string, req RenameTagRequest) (ImplResponse, error)
	DeleteTag(ctx context.Context, name string) (ImplResponse, error)
}

// SyncAPIService defines the business logic for the Sync API.
type SyncAPIService interface {
	GetChanges(ctx context.Context, since int32, limit int32) (ImplResponse, error)
}

// UserAPIService defines the business logic for the User API.
type UserAPIService interface {
	GetUser(ctx context.Context) (ImplResponse, error)
}

// FamilyAPIService defines the business logic for the Family API.
type FamilyAPIService interface {
	GetFamily(ctx context.Context) (ImplResponse, error)
	UpdateFamilySettings(ctx context.Context, req FamilySettingsRequest) (ImplResponse, error)
}
