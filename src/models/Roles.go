package models

import (
	"time"
)

// RoleName enum
type RoleName string

const (
	RoleNameAuthenticated RoleName = "Authenticated"
	RoleNameAnonymous     RoleName = "Anonymous"
	RoleNameBusinessOwner RoleName = "BusinessOwner"
	RoleNameSuperAdmin    RoleName = "SuperAdmin"
	RoleNamePaymentAgent  RoleName = "PaymentAgent"
	RoleNameAdmin         RoleName = "Admin"
)

// Role model
type Role struct {
	RoleId    string    `gorm:"primaryKey;type:varchar(25);column:role_id" json:"roleId,omitempty"`
	RoleName  RoleName  `gorm:"type:role_name;not null;column:role_name" json:"roleName,omitempty"`
	CreatedAt time.Time `gorm:"not null;column:created_at" json:"createdAt,omitempty"`
	UpdatedAt time.Time `gorm:"column:updated_at" json:"updatedAt,omitempty"`
	CreatedBy string    `gorm:"type:varchar(80);not null;column:created_by" json:"createdBy,omitempty"`
	UpdatedBy string    `gorm:"type:varchar(80);not null;column:updated_by" json:"updatedBy,omitempty"`

	// Relationships
	UserRoles []UserRole `gorm:"foreignKey:role_id;references:role_id;constraint:OnDelete:CASCADE,OnUpdate:CASCADE" json:"userRoles,omitempty"`
}

func (Role) TableName() string {
	return "roles"
}
