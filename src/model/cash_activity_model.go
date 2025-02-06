package model

import (
	"time"
)

// CashActivity Model
type CashActivity struct {
	ID            uint          `gorm:"primaryKey" json:"id"`
	AccountID     uint          `gorm:"not null" json:"account_id"`
	ReferenceID   *uint         `gorm:"null" json:"reference_id"` // Allow null for the first transaction
	Type          string        `gorm:"not null" json:"type"`     // 'debit' or 'credit'
	Nominal       float64       `gorm:"not null" json:"nominal"`
	BalanceBefore float64       `gorm:"not null" json:"balance_before"`
	BalanceAfter  float64       `gorm:"not null" json:"balance_after"`
	Description   string        `gorm:"type:text" json:"description"`
	CreatedAt     time.Time     `gorm:"autoCreateTime" json:"created_at"`
	Account       Account       `gorm:"foreignKey:AccountID;references:ID" json:"-"`   // Belongs to Account
	Reference     *CashActivity `gorm:"foreignKey:ReferenceID;references:ID" json:"-"` // Belongs to another CashActivity (previous transaction)
}

// DepositRequest struct for deposit operation (tabung)
type DepositRequest struct {
	AccountNumber string  `json:"no_rekening" validate:"required,numeric" example:"9876543210"`
	Nominal       float64 `json:"nominal" validate:"required,gt=0" example:"100000"`
}

// DepositResponse struct for deposit operation response
type DepositResponse struct {
	Balance float64 `json:"saldo" example:"500000"`
}

// Withdrawal struct for withdrawal operation (tarik)
type Withdrawal struct {
	AccountNumber string  `json:"no_rekening" validate:"required,numeric" example:"9876543210"`
	Nominal       float64 `json:"nominal" validate:"required,gt=0" example:"50000"`
}

// WithdrawalResponse struct for withdrawal operation response
type WithdrawalResponse struct {
	Balance float64 `json:"saldo" example:"450000"`
}

// BalanceResponse struct for checking balance (saldo) response
type BalanceResponse struct {
	Balance float64 `json:"saldo" example:"450000"`
}

// Mutation struct for withdrawal operation (tarik)
type Mutation struct {
	AccountNumber string `json:"no_rekening" validate:"required,numeric" example:"9876543210"`
	Month         int    `json:"bulan" validate:"required,numeric,lt=13,gt=0" example:"1"`
}
