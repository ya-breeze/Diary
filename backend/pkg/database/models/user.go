package models

import (
	"time"

	coremodels "github.com/ya-breeze/kin-core/models"
	"github.com/ya-breeze/diary.be/pkg/generated/goserver"
)

type User struct {
	coremodels.User
	StartDate time.Time
}

func (u User) FromDB() goserver.User {
	return goserver.User{
		Email:     u.Username,
		StartDate: u.StartDate,
	}
}
