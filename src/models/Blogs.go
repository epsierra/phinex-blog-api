package models

import (
	"encoding/json"
	"time"
)

// Blog model
type Blog struct {
	BlogId            string          `gorm:"primaryKey;type:varchar(25);column:blog_id" json:"blogId"`
	UserId            string          `gorm:"type:varchar(25);not null;column:user_id" json:"userId"`
	Slug              string          `gorm:"type:varchar(25);column:slug" json:"slug"`
	Title             string          `gorm:"type:varchar(255);column:title" json:"title"`
	Url               string          `gorm:"type:text;column:url" json:"url"`
	ExternalLink      string          `gorm:"type:text;column:external_link" json:"externalLink"`
	ExternalLinkTitle string          `gorm:"type:varchar(255);column:external_link_title" json:"externalLinkTitle"`
	Text              string          `gorm:"type:text;column:text" json:"text"`
	Images            json.RawMessage `gorm:"type:json;column:images" json:"images"`
	Video             string          `gorm:"type:text;column:video" json:"video"`
	Audio             string          `gorm:"type:text;column:audio" json:"audio"`
	CreatedAt         time.Time       `gorm:"not null;column:created_at" json:"createdAt"`
	UpdatedAt         time.Time       `gorm:"not null;column:updated_at" json:"updatedAt"`
	CreatedBy         string          `gorm:"type:varchar(80);not null;column:created_by" json:"createdBy"`
	UpdatedBy         string          `gorm:"type:varchar(80);not null;column:updated_by" json:"updatedBy"`
	CommentsCount     int64           `gorm:"column:comments_count" json:"commentsCount"`
	LikesCount        int64           `gorm:"column:likes_count" json:"likesCount"`
	SharesCount       int64           `gorm:"column:shares_count" json:"sharesCount"`
	ViewsCount        int64           `gorm:"column:views_count" json:"viewsCount"`
	IsReel            bool            `gorm:"type:boolean;default:false;column:is_reel" json:"isReel"`
	User              User            `gorm:"foreignKey:user_id;references:user_id;constraint:OnDelete:CASCADE,OnUpdate:CASCADE" json:"user,omitempty"`
	Comments          []Comment       `gorm:"foreignKey:ref_id;references:blog_id;constraint:OnDelete:CASCADE,OnUpdate:CASCADE" json:"comments,omitempty"`
	Likes             []Like          `gorm:"foreignKey:ref_id;references:blog_id;constraint:OnDelete:CASCADE,OnUpdate:CASCADE" json:"likes,omitempty"`
	Shares            []Share         `gorm:"foreignKey:ref_id;references:blog_id;constraint:OnDelete:CASCADE,OnUpdate:CASCADE" json:"shares,omitempty"`
	Views             []View          `gorm:"foreignKey:ref_id;references:blog_id;constraint:OnDelete:CASCADE,OnUpdate:CASCADE" json:"views,omitempty"`
}

func (Blog) TableName() string {
	return "blogs"
}
