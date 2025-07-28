package models

import (
	"time"
)

// Follow model
type Follow struct {
	FollowId    string    `gorm:"primaryKey;type:varchar(25);column:follow_id" json:"followId"`
	FollowerId  string    `gorm:"type:varchar(25);not null;column:follower_id" json:"followerId"`
	FollowingId string    `gorm:"type:varchar(25);not null;column:following_id" json:"followingId"`
	CreatedAt   time.Time `gorm:"not null;column:created_at" json:"createdAt"`
	UpdatedAt   time.Time `gorm:"column:updated_at" json:"updatedAt"`
	CreatedBy   string    `gorm:"type:varchar(80);not null;column:created_by" json:"createdBy"`
	UpdatedBy   string    `gorm:"type:varchar(80);not null;column:updated_by" json:"updatedBy"`

	Follower  User `gorm:"foreignKey:follower_id;references:user_id;constraint:OnDelete:CASCADE,OnUpdate:CASCADE" json:"follower"`
	Following User `gorm:"foreignKey:following_id;references:user_id;constraint:OnDelete:CASCADE,OnUpdate:CASCADE" json:"following"`
}

func (Follow) TableName() string {
	return "follows"
}
