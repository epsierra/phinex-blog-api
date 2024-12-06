package models

import (
	"time"

	"github.com/lib/pq"
)

type Blog struct {
	BlogId            string         `gorm:"primaryKey;type:uuid;default:uuid_generate_v4();column:blogId" json:"blogId"`
	UserId            string         `gorm:"type:uuid;column:userId;not null" json:"userId"`
	Slug              string         `gorm:"type:uuid;column:slug" json:"slug"`
	Title             string         `gorm:"column:title" json:"title"`
	URL               string         `gorm:"column:url" json:"url"`
	ExternalLink      string         `gorm:"type:text;column:externalLink" json:"externalLink"`
	ExternalLinkTitle string         `gorm:"column:externalLinkTitle" json:"externalLinkTitle"`
	Text              string         `gorm:"type:text;column:text" json:"text"`
	Images            pq.StringArray `gorm:"type:text[];column:images" json:"images"`
	Video             string         `gorm:"column:video" json:"video"`
	CreatedAt         time.Time      `gorm:"not null;column:createdAt" json:"createdAt"`
	UpdatedAt         time.Time      `gorm:"not null;column:updatedAt" json:"updatedAt"`

	// Belongs to User
	User User `json:"User"`
}

func (Blog) TableName() string {
	return "Blogs"
}
