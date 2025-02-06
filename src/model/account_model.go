package model

import (
	"time"
)

// Account Model
type Account struct {
	ID            uint           `gorm:"primaryKey" json:"id"`
	AccountNumber string         `gorm:"uniqueIndex;not null" json:"account_number"`
	FullName      string         `gorm:"not null" json:"full_name"`
	IDNumber      string         `gorm:"uniqueIndex;not null" json:"id_number"`
	PhoneNumber   string         `gorm:"uniqueIndex;not null" json:"phone_number"`
	Balance       float64        `gorm:"not null;default:0.00" json:"balance"`
	CreatedAt     time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt     time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	CashActivity  []CashActivity `gorm:"foreignKey:AccountID;references:ID" json:"-"`
}

// CreateAccount struct for account registration (daftar)
type CreateAccount struct {
	FullName    string `json:"nama" validate:"required,max=50" example:"John Doe"`
	IDNumber    string `json:"nik" validate:"required,len=16,numeric" example:"1234567890123456"`
	PhoneNumber string `json:"no_hp" validate:"required,max=15,numeric" example:"081234567890"`
}

// CreateAccountResponse struct for account of user registration
type CreateAccountResponse struct {
	AccountNumber string `json:"no_rekening" example:"9876543210"`
}

// ErrorResponse struct for error responses
type ErrorResponse struct {
	Remark string `json:"remark" example:"Invalid input data"`
}
