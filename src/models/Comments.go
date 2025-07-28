package models

import (
	"time"
)

// Comment model
type Comment struct {
	CommentId    string    `gorm:"primaryKey;type:varchar(25);column:comment_id" json:"commentId"`
	RefId        string    `gorm:"type:varchar(25);column:ref_id" json:"refId"`
	UserId       string    `gorm:"type:varchar(25);column:user_id" json:"userId"`
	Text         string    `gorm:"type:text;column:text" json:"text"`
	Image        string    `gorm:"type:text;column:image" json:"image"`
	Sticker      string    `gorm:"type:text;column:sticker" json:"sticker"`
	Video        string    `gorm:"type:text;column:video" json:"video"`
	CreatedAt    time.Time `gorm:"not null;column:created_at" json:"createdAt"`
	UpdatedAt    time.Time `gorm:"not null;column:updated_at" json:"updatedAt"`
	CreatedBy    string    `gorm:"type:varchar(80);not null;column:created_by" json:"createdBy"`
	UpdatedBy    string    `gorm:"type:varchar(80);not null;column:updated_by" json:"updatedBy"`
	RepliesCount int64     `gorm:"column:replies_count" json:"repliesCount"`
	LikesCount   int64     `gorm:"column:likes_count" json:"likesCount"`

	User User `gorm:"foreignKey:user_id;references:user_id;constraint:OnDelete:CASCADE,OnUpdate:CASCADE" json:"user,omitempty"`
}

func (Comment) TableName() string {
	return "comments"
}
