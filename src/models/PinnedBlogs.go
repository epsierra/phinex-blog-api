package models

import (
	"time"
)

type PinnedBlog struct {
	PinnedBlogId string    `gorm:"primaryKey;type:varchar(25);column:pinned_blog_id" json:"pinnedBlogId,omitempty"`
	BlogId       string    `gorm:"type:varchar(25);not null;unique;column:blog_id" json:"blogId,omitempty"`
	UserId       string    `gorm:"type:varchar(25);not null;column:user_id" json:"userId,omitempty"`
	StartDate    time.Time `gorm:"not null;column:start_date" json:"startDate,omitempty"`
	EndDate      time.Time `gorm:"not null;column:end_date" json:"endDate,omitempty"`
	CreatedAt    time.Time `gorm:"not null;column:created_at" json:"createdAt,omitempty"`
	UpdatedAt    time.Time `gorm:"column:updated_at" json:"updatedAt,omitempty"`
	CreatedBy    string    `gorm:"type:varchar(80);not null;column:created_by" json:"createdBy,omitempty"`
	UpdatedBy    string    `gorm:"type:varchar(80);not null;column:updated_by" json:"updatedBy,omitempty"`

	Blog *Blog `gorm:"foreignKey:blog_id;references:blog_id;constraint:OnDelete:CASCADE,OnUpdate:CASCADE" json:"blog,omitempty"`
	User *User `gorm:"foreignKey:user_id;references:user_id;constraint:OnDelete:CASCADE,OnUpdate:CASCADE" json:"user,omitempty"`
}

func (PinnedBlog) TableName() string {
	return "pinned_blogs"
}
