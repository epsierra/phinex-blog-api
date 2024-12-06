package models

import (
	"time"
)

type Comment struct {
	CommentId string    `gorm:"primaryKey;type:uuid;default:uuid_generate_v4();column:commentId" json:"commentId"`
	RefId     string    `gorm:"type:uuid;column:refId" json:"refId"`
	UserId    string    `gorm:"type:uuid;not null;column:userId" json:"userId"`
	Content   string    `gorm:"type:text;not null;column:content" json:"content"`
	CreatedAt time.Time `gorm:"not null;column:createdAt" json:"createdAt"`
	UpdatedAt time.Time `gorm:"default:CURRENT_TIMESTAMP;column:updatedAt" json:"updatedAt"`

	User User
}

func (Comment) TableName() string {
	return "Comments"
}
