package models

import (
	coremodels "github.com/ya-breeze/kin-core/models"
	"github.com/ya-breeze/diary.be/pkg/generated/goserver"
)

type Family struct {
	coremodels.Family
	Users []User
}

func (f Family) FromDB() goserver.FamilyResponse {
	members := make([]goserver.FamilyMember, len(f.Users))
	for i, u := range f.Users {
		members[i] = goserver.FamilyMember{Email: u.Username}
	}
	return goserver.FamilyResponse{
		Id:      f.ID,
		Name:    f.Name,
		Members: members,
	}
}
