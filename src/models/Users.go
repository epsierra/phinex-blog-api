package models

import (
	"time"

	"gorm.io/gorm"
)

type User struct {
	UserId       string    `gorm:"primaryKey;type:uuid;default:uuid_generate_v4();column:userId" json:"userId,omitempty"`
	FirstName    string    `gorm:"type:varchar(255);column:firstName" json:"firstName,omitempty"`
	MiddleName   string    `gorm:"type:varchar(255);column:middleName" json:"middleName,omitempty"`
	LastName     string    `gorm:"type:varchar(255);column:lastName" json:"lastName,omitempty"`
	FullName     string    `gorm:"type:varchar(255);column:fullName" json:"fullName,omitempty"`
	ProfileImage string    `gorm:"type:text;column:profileImage" json:"profileImage,omitempty"`
	Bio          string    `gorm:"type:text;column:bio" json:"bio,omitempty"`
	PhoneNumber  string    `gorm:"type:varchar(255);column:phoneNumber" json:"phoneNumber,omitempty"`
	Status       string    `gorm:"default:'active';column:status" json:"status,omitempty"`
	Password     string    `gorm:"type:varchar(255);column:password" json:"password,omitempty"`
	Gender       string    `gorm:"type:varchar(255);column:gender" json:"gender,omitempty"`
	Dob          string    `gorm:"type:varchar(255);column:dob" json:"dob,omitempty"`
	Email        string    `gorm:"type:varchar(255);unique;column:email" json:"email,omitempty"`
	Verified     bool      `gorm:"type:boolean;default:false;column:verified" json:"verified,omitempty"`
	CreatedAt    time.Time `gorm:"not null;column:createdAt" json:"createdAt,omitempty"`
	UpdatedAt    time.Time `gorm:"default:CURRENT_TIMESTAMP;column:updatedAt" json:"updatedAt,omitempty"`

	// One-to-many relationship with Blog
	Blogs []Blog `gorm:"foreignKey:UserId;references:UserId;constraint:OnDelete:CASCADE,OnUpdate:CASCADE" json:"blogs,omitempty"`
}

func (User) TableName() string {
	return "Users"
}

// Remove the Password field from query results
func (u *User) AfterFind(tx *gorm.DB) (err error) {
	u.Password = "" // Clear the password after a query
	return
}
