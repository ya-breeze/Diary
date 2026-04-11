package models

import (
	coremodels "github.com/ya-breeze/kin-core/models"
)

type Family struct {
	coremodels.Family
	Users []User
}
