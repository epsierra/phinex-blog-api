package models

import (
	"time"
)

// Share model
type Share struct {
	ShareId   string    `gorm:"primaryKey;type:varchar(25);column:share_id" json:"shareId"`
	RefId     string    `gorm:"type:varchar(25);not null;column:ref_id" json:"refId"`
	UserId    string    `gorm:"type:varchar(25);not null;column:user_id" json:"userId"`
	CreatedAt time.Time `gorm:"not null;column:created_at" json:"createdAt"`
	UpdatedAt time.Time `gorm:"column:updated_at" json:"updatedAt"`
	CreatedBy string    `gorm:"type:varchar(80);not null;column:created_by" json:"createdBy"`
	UpdatedBy string    `gorm:"type:varchar(80);not null;column:updated_by" json:"updatedBy"`

	User User `gorm:"foreignKey:user_id;references:user_id;constraint:OnDelete:CASCADE,OnUpdate:CASCADE" json:"user"`
}

func (Share) TableName() string {
	return "shares"
}
