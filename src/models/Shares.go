package models

import (
	"time"
)

type Share struct {
	ShareId   string    `gorm:"primaryKey;type:uuid;default:uuid_generate_v4();column:shareId" json:"shareId"`
	RefId     string    `gorm:"type:uuid;not null;column:refId" json:"refId"`
	UserId    string    `gorm:"type:uuid;not null;column:userId" json:"userId"`
	CreatedAt time.Time `gorm:"not null;column:createdAt" json:"createdAt"`
	UpdatedAt time.Time `gorm:"default:CURRENT_TIMESTAMP;column:updatedAt" json:"updatedAt"`
}

func (Share) TableName() string {
	return "Shares"
}
