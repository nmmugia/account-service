package service

import (
	"account-service/src/model"
	"account-service/src/utils"
	"context"
	"errors"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type AccountServices interface {
	CreateAccount(c context.Context, req *model.CreateAccount) (*model.Account, error)
	Deposit(c context.Context, req *model.DepositRequest) error
	Withdraw(c context.Context, req *model.Withdrawal) error
	GetBalance(c context.Context, id string) (*model.Account, error)
	Mutation(c context.Context, req *model.Mutation) ([]*model.CashActivity, error) // You might not need this in the initial implementation
}

type AccountService struct {
	Log      *logrus.Logger
	DB       *gorm.DB
	Validate *validator.Validate
}

func NewAccountService(db *gorm.DB, validate *validator.Validate) AccountServices { // Corrected: return AccountService
	return &AccountService{
		Log:      utils.Log,
		DB:       db,
		Validate: validate,
	}
}

var (
	ErrDuplicateIDNumber    = errors.New("ID number already registered")
	ErrDuplicatePhoneNumber = errors.New("phone number already registered")
	ErrAccountNotFound      = errors.New("account not found")
	ErrInsufficientBalance  = errors.New("insufficient balance")
)

func (s *AccountService) CreateAccount(c context.Context, req *model.CreateAccount) (*model.Account, error) {
	// Validate the request struct
	if err := s.Validate.Struct(req); err != nil {
		return nil, fiber.NewError(fiber.StatusBadRequest, err.Error())
	}

	// Check for duplicate ID number
	var existingAccountByID model.Account
	if err := s.DB.WithContext(c).Where("id_number = ?", req.IDNumber).First(&existingAccountByID).Error; err == nil {
		return nil, fiber.NewError(fiber.StatusConflict, ErrDuplicateIDNumber.Error())
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		s.Log.Errorf("Error checking for duplicate ID number: %+v", err)
		return nil, fiber.NewError(fiber.StatusInternalServerError, "Database error")
	}

	// Check for duplicate phone number
	var existingAccountByPhone model.Account
	if err := s.DB.WithContext(c).Where("phone_number = ?", req.PhoneNumber).First(&existingAccountByPhone).Error; err == nil {
		return nil, fiber.NewError(fiber.StatusConflict, ErrDuplicatePhoneNumber.Error())
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		s.Log.Errorf("Error checking for duplicate phone number: %+v", err)
		return nil, fiber.NewError(fiber.StatusInternalServerError, "Database error")
	}

	// Generate a unique account number
	accountNumber := utils.GenerateAccountNumber()
	for {
		var tempAccount model.Account
		err := s.DB.WithContext(c).Where("account_number = ?", accountNumber).First(&tempAccount).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			break // Account number is unique
		}
		if err != nil {
			s.Log.Errorf("Error checking for unique account number: %+v", err)
			return nil, fiber.NewError(fiber.StatusInternalServerError, "Database error")
		}
		accountNumber = utils.GenerateAccountNumber() // Regenerate if not unique
	}

	// Create the new account
	newAccount := model.Account{
		AccountNumber: accountNumber,
		FullName:      req.FullName,
		IDNumber:      req.IDNumber,
		PhoneNumber:   req.PhoneNumber,
	}

	if err := s.DB.WithContext(c).Create(&newAccount).Error; err != nil {
		s.Log.Errorf("failed to create account: %+v", err)
		return nil, fiber.NewError(fiber.StatusInternalServerError, "failed to create account")
	}

	return &newAccount, nil
}

func (s *AccountService) Deposit(c context.Context, req *model.DepositRequest) error {
	// Validate input
	if err := s.Validate.Struct(req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}

	// Get account
	var account model.Account
	if err := s.DB.Where("account_number = ?", req.AccountNumber).First(&account).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fiber.NewError(fiber.StatusNotFound, ErrAccountNotFound.Error())
		}
		s.Log.Errorf("Failed to get account: %+v", err)
		return fiber.NewError(fiber.StatusInternalServerError, "Database error")
	}

	// Get the latest cash activity for the account

	// Create cash activity record
	newActivity := model.CashActivity{
		AccountID:     account.ID,
		ReferenceID:   nil,
		Type:          "credit",
		Nominal:       req.Nominal,
		BalanceBefore: account.Balance,
		BalanceAfter:  account.Balance + req.Nominal,
	}

	tx := s.DB.Begin() //begin transaction

	if err := tx.Create(&newActivity).Error; err != nil {
		s.Log.Errorf("Failed to create cash activity: %+v", err)
		tx.Rollback() //rollback
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to record transaction")
	}

	// Update account balance
	account.Balance += req.Nominal
	if err := tx.Save(&account).Error; err != nil {
		s.Log.Errorf("Failed to update account balance: %+v", err)
		tx.Rollback() // Rollback if the update fails
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to record transaction")
	}

	tx.Commit() // commit transaction

	return nil
}

func (s *AccountService) Withdraw(c context.Context, req *model.Withdrawal) error {
	// Validate input
	if err := s.Validate.Struct(req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}

	// Begin transaction
	tx := s.DB.WithContext(c).Begin()
	if tx.Error != nil {
		s.Log.Errorf("Failed to start transaction: %+v", tx.Error)
		return fiber.NewError(fiber.StatusInternalServerError, "Transaction failed")
	}

	defer func() {
		if r := recover(); r != nil {
			tx.Rollback() // Rollback if panic
		}
	}()

	// Get account
	var account model.Account
	if err := tx.Where("account_number = ?", req.AccountNumber).First(&account).Error; err != nil {
		tx.Rollback()
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fiber.NewError(fiber.StatusNotFound, ErrAccountNotFound.Error())
		}
		s.Log.Errorf("Failed to get account: %+v", err)
		return fiber.NewError(fiber.StatusInternalServerError, "Database error")
	}

	// Check for sufficient balance
	if account.Balance < req.Nominal {
		tx.Rollback()
		return fiber.NewError(fiber.StatusBadRequest, ErrInsufficientBalance.Error())
	}

	// Get the latest cash activity for the account
	var latestActivity model.CashActivity
	err := tx.Where("account_id = ?", account.ID).Order("created_at desc").First(&latestActivity).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		tx.Rollback()
		s.Log.Errorf("Failed to get latest cash activity: %+v", err)
		return fiber.NewError(fiber.StatusInternalServerError, "Database error during getting latest cash activity")
	}
	var refID *uint
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		refID = &latestActivity.ID
	}

	// Create cash activity record
	newActivity := model.CashActivity{
		AccountID:     account.ID,
		ReferenceID:   refID,
		Type:          "debit",
		Nominal:       req.Nominal,
		BalanceBefore: account.Balance,
		BalanceAfter:  account.Balance - req.Nominal,
	}

	if err := tx.Create(&newActivity).Error; err != nil {
		tx.Rollback()
		s.Log.Errorf("Failed to create cash activity: %+v", err)
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to record transaction")
	}

	account.Balance -= req.Nominal
	if err := tx.Save(&account).Error; err != nil {
		s.Log.Errorf("Failed to update account balance: %+v", err)
		tx.Rollback() // Rollback if the update fails
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to record transaction")
	}

	// Commit the transaction
	if err := tx.Commit().Error; err != nil {
		s.Log.Errorf("Failed to commit transaction: %+v", err)
		return fiber.NewError(fiber.StatusInternalServerError, "Transaction failed")
	}
	return nil
}

func (s *AccountService) GetBalance(c context.Context, accountNumber string) (*model.Account, error) {
	var account model.Account
	if err := s.DB.WithContext(c).Where("account_number = ?", accountNumber).First(&account).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fiber.NewError(fiber.StatusNotFound, ErrAccountNotFound.Error())
		}
		s.Log.Errorf("Failed to get account: %+v", err)
		return nil, fiber.NewError(fiber.StatusInternalServerError, "Database error")
	}
	return &account, nil
}

// Mutation -  Placeholder, as it's complex and needs more context
func (s *AccountService) Mutation(c context.Context, req *model.Mutation) ([]*model.CashActivity, error) {
	// Implement logic to fetch account mutations (transaction history).
	// This typically involves querying the cash_activity table based on account ID and date range.
	// You'll need to define the structure of the `req` parameter (e.g., date range, account ID).
	return nil, fiber.NewError(fiber.StatusNotImplemented, "Mutation functionality not implemented yet")
}
