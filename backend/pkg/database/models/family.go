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
	// AITaggingBackfill enables the background "untagged" health check that finds
	// untagged/stale days and surfaces them via the health-issues flow.
	AITaggingBackfill bool `gorm:"default:false"`
	// AITaggingAuto lets unattended triggers (on-save, backfill) auto-apply
	// confident suggestions to untagged days instead of only staging pending tags.
	AITaggingAuto bool `gorm:"default:false"`
	// AITaggingUseImages sends the entry's referenced image assets to Gemini
	// alongside the text. Off by default; privacy-sensitive opt-in.
	AITaggingUseImages bool `gorm:"default:false"`
	// AITaggingUseVideo sends keyframes extracted from referenced video assets
	// to Gemini alongside the text. Off by default; requires ffmpeg at runtime.
	AITaggingUseVideo bool `gorm:"default:false"`
}

func (f Family) FromDB() goserver.FamilyResponse {
	members := make([]goserver.FamilyMember, len(f.Users))
	for i, u := range f.Users {
		members[i] = goserver.FamilyMember{Email: u.Username}
	}
	aiTaggingEnabled := f.AITaggingEnabled
	aiTaggingBackfill := f.AITaggingBackfill
	aiTaggingAuto := f.AITaggingAuto
	aiTaggingUseImages := f.AITaggingUseImages
	aiTaggingUseVideo := f.AITaggingUseVideo
	return goserver.FamilyResponse{
		Id:                 f.ID,
		Name:               f.Name,
		Members:            members,
		AiTaggingEnabled:   &aiTaggingEnabled,
		AiTaggingBackfill:  &aiTaggingBackfill,
		AiTaggingAuto:      &aiTaggingAuto,
		AiTaggingUseImages: &aiTaggingUseImages,
		AiTaggingUseVideo:  &aiTaggingUseVideo,
	}
}
