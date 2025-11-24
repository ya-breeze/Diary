package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/ya-breeze/diary.be/pkg/generated/goserver"
)

type User struct {
	ID             uuid.UUID `gorm:"type:uuid;primaryKey"`
	StartDate      time.Time
	Login          string `gorm:"unique"`
	HashedPassword string
}

func (u User) FromDB() goserver.User {
	return goserver.User{
		Email:     u.Login,
		StartDate: u.StartDate,
	}
}
