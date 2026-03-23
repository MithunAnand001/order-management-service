package models

import (
	"time"

	"github.com/google/uuid"
)

type Base struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	UUID       uuid.UUID `gorm:"type:uuid;unique;not null;default:gen_random_uuid()" json:"uuid"`
	CreatedOn  time.Time `gorm:"autoCreateTime;type:timestamptz;not null" json:"created_on"`
	CreatedBy  *uint     `json:"created_by"`
	ModifiedOn time.Time `gorm:"autoUpdateTime;type:timestamptz;not null" json:"modified_on"`
	ModifiedBy *uint     `json:"modified_by"`
	IsActive   bool      `gorm:"default:true" json:"is_active"`
}
