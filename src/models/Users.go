package models

import (
	"time"

	"gorm.io/gorm"
)

// Enums
type UserStatus string

const (
	UserStatusActive    UserStatus = "active"
	UserStatusBanned    UserStatus = "banned"
	UserStatusSuspended UserStatus = "suspended"
	UserStatusOnline    UserStatus = "online"
)

// User model
type User struct {
	UserId                string     `gorm:"primaryKey;type:varchar(25);column:user_id" json:"userId,omitempty"`
	FirstName             string     `gorm:"type:varchar(255);column:first_name" json:"firstName,omitempty"`
	MiddleName            string     `gorm:"type:varchar(255);column:middle_name" json:"middleName,omitempty"`
	FullName              string     `gorm:"type:varchar(255);column:full_name" json:"fullName,omitempty"`
	UserName              string     `gorm:"type:varchar(255);column:user_name" json:"userName,omitempty"`
	LastName              string     `gorm:"type:varchar(255);column:last_name" json:"lastName,omitempty"`
	ProfileImage          string     `gorm:"type:text;column:profile_image" json:"profileImage,omitempty"`
	Bio                   string     `gorm:"type:text;column:bio" json:"bio,omitempty"`
	PhoneNumber           string     `gorm:"type:varchar(255);column:phone_number" json:"phoneNumber,omitempty"`
	Status                UserStatus `gorm:"type:user_status;default:'active';column:status" json:"status,omitempty"`
	Password              string     `gorm:"type:varchar(255);not null;column:password" json:"password,omitempty"`
	Gender                string     `gorm:"type:varchar(255);column:gender" json:"gender,omitempty"`
	Dob                   string     `gorm:"type:varchar(255);column:dob" json:"dob,omitempty"`
	Email                 string     `gorm:"type:varchar(255);unique;column:email;not null" json:"email"`
	Verified              bool       `gorm:"type:boolean;default:false;column:verified" json:"verified"`
	EmailIsVerified       bool       `gorm:"type:boolean;default:false;column:email_is_verified" json:"emailIsVerified,omitempty"`
	PhoneNumberIsVerified bool       `gorm:"type:boolean;default:false;column:phone_number_is_verified" json:"phoneNumberIsVerified,omitempty"`
	CreatedAt             time.Time  `gorm:"not null;column:created_at" json:"createdAt,omitempty"`
	UpdatedAt             time.Time  `gorm:"column:updated_at" json:"updatedAt,omitempty"`
	CreatedBy             string     `gorm:"type:varchar(80);not null;column:created_by" json:"createdBy,omitempty"`
	UpdatedBy             string     `gorm:"type:varchar(80);not null;column:updated_by" json:"updatedBy,omitempty"`

	// Relationships
	Blogs      []Blog      `gorm:"foreignKey:user_id;references:user_id;constraint:OnDelete:CASCADE,OnUpdate:CASCADE" json:"blogs,omitempty"`
	Comments   []Comment   `gorm:"foreignKey:user_id;references:user_id;constraint:OnDelete:CASCADE,OnUpdate:CASCADE" json:"comments,omitempty"`
	Likes      []Like      `gorm:"foreignKey:user_id;references:user_id;constraint:OnDelete:CASCADE,OnUpdate:CASCADE" json:"likes,omitempty"`
	Shares     []Share     `gorm:"foreignKey:user_id;references:user_id;constraint:OnDelete:CASCADE,OnUpdate:CASCADE" json:"shares,omitempty"`
	UserRoles  []UserRole  `gorm:"foreignKey:user_id;references:user_id;constraint:OnDelete:CASCADE,OnUpdate:CASCADE" json:"userRoles,omitempty"`
	UsersStats *UsersStats `gorm:"foreignKey:user_id;references:user_id;constraint:OnDelete:CASCADE,OnUpdate:CASCADE" json:"usersStats,omitempty"`
	Following  bool        `gorm:"-" json:"following"`
}

func (User) TableName() string {
	return "users"
}

func (u *User) AfterFind(tx *gorm.DB) (err error) {
	u.Password = ""
	return
}
