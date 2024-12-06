package models

import (
	"time"
)

type Follow struct {
	FollowId    string    `gorm:"primaryKey;type:uuid;default:uuid_generate_v4();column:followId" json:"followId"`
	FollowerId  string    `gorm:"type:uuid;not null;column:followerId" json:"followerId"`
	FollowingId string    `gorm:"type:uuid;not null;column:followingId" json:"followingId"`
	CreatedAt   time.Time `gorm:"not null;column:createdAt" json:"createdAt"`
	UpdatedAt   time.Time `gorm:"default:CURRENT_TIMESTAMP;column:updatedAt" json:"updatedAt"`
}

func (Follow) TableName() string {
	return "Follows"
}
