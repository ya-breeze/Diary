package models

import (
	"github.com/ya-breeze/diary.be/pkg/generated/goserver"
	coremodels "github.com/ya-breeze/kin-core/models"
)

type Family struct {
	coremodels.Family
	Users []User
	// AITaggingEnabled is the per-family master switch for AI tag suggestion.
	// Defaults to false; the feature also requires GEMINI_API_KEY on the server.
	AITaggingEnabled bool `gorm:"default:false"`
}

func (f Family) FromDB() goserver.FamilyResponse {
	members := make([]goserver.FamilyMember, len(f.Users))
	for i, u := range f.Users {
		members[i] = goserver.FamilyMember{Email: u.Username}
	}
	aiTaggingEnabled := f.AITaggingEnabled
	return goserver.FamilyResponse{
		Id:               f.ID,
		Name:             f.Name,
		Members:          members,
		AiTaggingEnabled: &aiTaggingEnabled,
	}
}
