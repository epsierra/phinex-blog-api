package models

import (
	"time"
)

// Wallet model
type Wallet struct {
	WalletId      string    `gorm:"primaryKey;type:varchar(25);column:wallet_id" json:"walletId,omitempty"`
	AccountNumber string    `gorm:"type:varchar(20);column:account_number" json:"accountNumber,omitempty"`
	UserId        string    `gorm:"type:varchar(25);unique;not null;column:user_id" json:"userId,omitempty"`
	Balance       float64   `gorm:"type:numeric(15,2);not null;default:0.00;column:balance" json:"balance,omitempty"`
	Currency      string    `gorm:"type:varchar(3);not null;default:'SLE';column:currency" json:"currency,omitempty"`
	IsActive      bool      `gorm:"type:boolean;not null;default:true;column:is_active" json:"isActive,omitempty"`
	CreatedBy     string    `gorm:"type:varchar(80);not null;column:created_by" json:"createdBy,omitempty"`
	CreatedAt     time.Time `gorm:"not null;default:CURRENT_TIMESTAMP;column:created_at" json:"createdAt,omitempty"`
	UpdatedBy     string    `gorm:"type:varchar(80);not null;column:updated_by" json:"updatedBy,omitempty"`
	UpdatedAt     time.Time `gorm:"not null;default:CURRENT_TIMESTAMP;column:updated_at" json:"updatedAt,omitempty"`

	// Relationship
	User User `gorm:"foreignKey:UserId;references:UserId" json:"user,omitempty"`
}

func (Wallet) TableName() string {
	return "wallets"
}
