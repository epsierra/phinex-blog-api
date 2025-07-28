package models

import (
	"time"
)

// UserRole model
type UserRole struct {
	UserRoleId string    `gorm:"primaryKey;type:varchar(25);column:user_role_id" json:"userRoleId,omitempty"`
	UserId     string    `gorm:"type:varchar(25);not null;column:user_id" json:"userId,omitempty"`
	RoleId     string    `gorm:"type:varchar(25);not null;column:role_id" json:"roleId,omitempty"`
	CreatedAt  time.Time `gorm:"not null;column:created_at" json:"createdAt,omitempty"`
	UpdatedAt  time.Time `gorm:"column:updated_at" json:"updatedAt,omitempty"`
	CreatedBy  string    `gorm:"type:varchar(80);not null;column:created_by" json:"createdBy,omitempty"`
	UpdatedBy  string    `gorm:"type:varchar(80);not null;column:updated_by" json:"updatedBy,omitempty"`

	// Relationships
	User User `gorm:"foreignKey:user_id;references:user_id;constraint:OnDelete:CASCADE,OnUpdate:CASCADE" json:"user,omitempty"`
	Role Role `gorm:"foreignKey:role_id;references:role_id;constraint:OnDelete:CASCADE,OnUpdate:CASCADE" json:"role,omitempty"`
}

func (UserRole) TableName() string {
	return "user_roles"
}
