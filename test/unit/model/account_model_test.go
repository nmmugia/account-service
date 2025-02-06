package model_test

import (
	"account-service/src/model"
	"account-service/src/utils"
	"testing"

	"github.com/stretchr/testify/assert"
)

var validate = utils.Validator()

func TestAccountModel(t *testing.T) {
	t.Run("CreateAccount validation", func(t *testing.T) {
		validAccount := model.CreateAccount{
			FullName:    "John Doe",
			IDNumber:    "1234567890123456",
			PhoneNumber: "081234567890",
		}

		t.Run("should validate a valid account", func(t *testing.T) {
			err := validate.Struct(validAccount)
			assert.NoError(t, err)
		})

		t.Run("should fail with missing FullName", func(t *testing.T) {
			invalidAccount := validAccount
			invalidAccount.FullName = ""
			err := validate.Struct(invalidAccount)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "FullName") // Check field name
		})

		t.Run("should fail with too long FullName", func(t *testing.T) {
			invalidAccount := validAccount
			invalidAccount.FullName = "This is a very long name that exceeds the maximum allowed length"
			err := validate.Struct(invalidAccount)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "FullName")
		})

		t.Run("should fail with invalid IDNumber (length)", func(t *testing.T) {
			invalidAccount := validAccount
			invalidAccount.IDNumber = "12345"
			err := validate.Struct(invalidAccount)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "IDNumber")
		})

		t.Run("should fail with invalid IDNumber (not numeric)", func(t *testing.T) {
			invalidAccount := validAccount
			invalidAccount.IDNumber = "abcdefghijklmnop"
			err := validate.Struct(invalidAccount)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "IDNumber")
		})

		t.Run("should fail with invalid PhoneNumber (length)", func(t *testing.T) {
			invalidAccount := validAccount
			invalidAccount.PhoneNumber = "00000000000000000" // Too long
			err := validate.Struct(invalidAccount)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "PhoneNumber")
		})
		t.Run("should fail with invalid PhoneNumber (not numeric)", func(t *testing.T) {
			invalidAccount := validAccount
			invalidAccount.PhoneNumber = "abcdefghijk"
			err := validate.Struct(invalidAccount)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "PhoneNumber")
		})
		t.Run("should fail with empty PhoneNumber", func(t *testing.T) {
			invalidAccount := validAccount
			invalidAccount.PhoneNumber = ""
			err := validate.Struct(invalidAccount)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "PhoneNumber")
		})
	})

	// No need to test Account struct directly, as it's validated via GORM tags and during DB operations.
	// If you had custom methods on Account, you would test those here.
}

func TestDepositRequestModel(t *testing.T) {
	t.Run("DepositRequest validation", func(t *testing.T) {
		validDeposit := model.DepositRequest{
			AccountNumber: "1234567890",
			Nominal:       100.0,
		}

		t.Run("should validate a valid deposit request", func(t *testing.T) {
			err := validate.Struct(validDeposit)
			assert.NoError(t, err)
		})

		t.Run("should fail with empty AccountNumber", func(t *testing.T) {
			invalidDeposit := validDeposit
			invalidDeposit.AccountNumber = ""
			err := validate.Struct(invalidDeposit)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "AccountNumber")
		})

		t.Run("should fail with zero Nominal", func(t *testing.T) {
			invalidDeposit := validDeposit
			invalidDeposit.Nominal = 0
			err := validate.Struct(invalidDeposit)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "Nominal")
		})

		t.Run("should fail with negative Nominal", func(t *testing.T) {
			invalidDeposit := validDeposit
			invalidDeposit.Nominal = -100.0
			err := validate.Struct(invalidDeposit)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "Nominal")
		})
	})
}

func TestWithdrawalModel(t *testing.T) {
	t.Run("Withdrawal validation", func(t *testing.T) {
		validWithdrawal := model.Withdrawal{
			AccountNumber: "1234567890",
			Nominal:       100.0,
		}

		t.Run("should validate a valid withdrawal request", func(t *testing.T) {
			err := validate.Struct(validWithdrawal)
			assert.NoError(t, err)
		})

		t.Run("should fail with empty AccountNumber", func(t *testing.T) {
			invalidWithdrawal := validWithdrawal
			invalidWithdrawal.AccountNumber = ""
			err := validate.Struct(invalidWithdrawal)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "AccountNumber")
		})
		t.Run("should fail with zero Nominal", func(t *testing.T) {
			invalidWithdrawal := validWithdrawal
			invalidWithdrawal.Nominal = 0
			err := validate.Struct(invalidWithdrawal)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "Nominal")
		})

		t.Run("should fail with negative Nominal", func(t *testing.T) {
			invalidWithdrawal := validWithdrawal
			invalidWithdrawal.Nominal = -100.0
			err := validate.Struct(invalidWithdrawal)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "Nominal")
		})
	})
}
