package models

import (
	"time"
)

type Like struct {
	LikeId    string    `gorm:"primaryKey;type:uuid;default:uuid_generate_v4();column:likeId" json:"likeId"`
	RefId     string    `gorm:"type:uuid;column:refId" json:"refId"`
	UserId    string    `gorm:"type:uuid;not null;column:userId" json:"userId"`
	CreatedAt time.Time `gorm:"not null;column:createdAt" json:"createdAt"`
	UpdatedAt time.Time `gorm:"default:CURRENT_TIMESTAMP;column:updatedAt" json:"updatedAt"`

	User User
}

func (Like) TableName() string {
	return "Likes"
}
