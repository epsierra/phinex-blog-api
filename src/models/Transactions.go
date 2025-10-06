package models

import (
	"time"
)

// TransactionType defines the type of a transaction
type TransactionType string

const (
	Deposit    TransactionType = "deposit"
	Withdrawal TransactionType = "withdrawal"
	Transfer   TransactionType = "transfer"
)

// TransactionStatus defines the status of a transaction
type TransactionStatus string

const (
	Pending   TransactionStatus = "pending"
	Completed TransactionStatus = "completed"
	Failed    TransactionStatus = "failed"
)

// Transaction model
type Transaction struct {
	TransactionId string            `gorm:"primaryKey;type:varchar(25);column:transaction_id" json:"transactionId,omitempty"`
	WalletId      string            `gorm:"type:varchar(25);not null;column:wallet_id" json:"walletId,omitempty"`
	Amount        float64           `gorm:"type:numeric(15,2);not null;column:amount" json:"amount,omitempty"`
	Type          TransactionType   `gorm:"type:varchar(20);not null;column:type" json:"type,omitempty"`
	Status        TransactionStatus `gorm:"type:varchar(20);not null;column:status" json:"status,omitempty"`
	Description   string            `gorm:"type:varchar(255);column:description" json:"description,omitempty"`
	CreatedAt     time.Time         `gorm:"not null;default:CURRENT_TIMESTAMP;column:created_at" json:"createdAt,omitempty"`
	UpdatedAt     time.Time         `gorm:"not null;default:CURRENT_TIMESTAMP;column:updated_at" json:"updatedAt,omitempty"`
	CreatedBy     string            `gorm:"type:varchar(80);not null;column:created_by" json:"createdBy,omitempty"`
	UpdatedBy     string            `gorm:"type:varchar(80);not null;column:updated_by" json:"updatedBy,omitempty"`

	// Relationship
	Wallet Wallet `gorm:"foreignKey:WalletId;references:WalletId" json:"wallet,omitempty"`
}

func (Transaction) TableName() string {
	return "transactions"
}
