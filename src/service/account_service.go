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
	CreateAccount(c context.Context, req *model.CreateAccount) (*model.Account, *fiber.Error)
	Deposit(c context.Context, req *model.DepositRequest) *fiber.Error
	Withdraw(c context.Context, req *model.Withdrawal) *fiber.Error
	GetBalance(c context.Context, id string) (*model.Account, *fiber.Error)
}

type AccountService struct {
	Log      *logrus.Logger
	DB       *gorm.DB
	Validate *validator.Validate
}

func NewAccountService(db *gorm.DB, validate *validator.Validate) AccountServices {
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

func (accountService *AccountService) CreateAccount(c context.Context, req *model.CreateAccount) (*model.Account, *fiber.Error) {

	if err := accountService.Validate.Struct(req); err != nil {
		return nil, fiber.NewError(fiber.StatusBadRequest, err.Error())
	}

	var existingAccountByID model.Account
	if err := accountService.DB.WithContext(c).Where("id_number = ?", req.IDNumber).First(&existingAccountByID).Error; err == nil {
		return nil, fiber.NewError(fiber.StatusConflict, ErrDuplicateIDNumber.Error())
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		accountService.Log.Errorf("Error checking for duplicate ID number: %+v", err)
		return nil, fiber.NewError(fiber.StatusInternalServerError, "Database error")
	}

	var existingAccountByPhone model.Account
	if err := accountService.DB.WithContext(c).Where("phone_number = ?", req.PhoneNumber).First(&existingAccountByPhone).Error; err == nil {
		return nil, fiber.NewError(fiber.StatusConflict, ErrDuplicatePhoneNumber.Error())
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		accountService.Log.Errorf("Error checking for duplicate phone number: %+v", err)
		return nil, fiber.NewError(fiber.StatusInternalServerError, "Database error")
	}

	accountNumber := utils.GenerateAccountNumber()
	for {
		var tempAccount model.Account
		err := accountService.DB.WithContext(c).Where("account_number = ?", accountNumber).First(&tempAccount).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			break
		}
		if err != nil {
			accountService.Log.Errorf("Error checking for unique account number: %+v", err)
			return nil, fiber.NewError(fiber.StatusInternalServerError, "Database error")
		}
		accountNumber = utils.GenerateAccountNumber()
	}

	newAccount := model.Account{
		AccountNumber: accountNumber,
		FullName:      req.FullName,
		IDNumber:      req.IDNumber,
		PhoneNumber:   req.PhoneNumber,
	}

	if err := accountService.DB.WithContext(c).Create(&newAccount).Error; err != nil {
		accountService.Log.Errorf("failed to create account: %+v", err)
		return nil, fiber.NewError(fiber.StatusInternalServerError, "failed to create account")
	}

	return &newAccount, nil
}

func (accountService *AccountService) Deposit(c context.Context, req *model.DepositRequest) *fiber.Error {

	if err := accountService.Validate.Struct(req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}

	var account model.Account
	if err := accountService.DB.Where("account_number = ?", req.AccountNumber).First(&account).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fiber.NewError(fiber.StatusNotFound, ErrAccountNotFound.Error())
		}
		accountService.Log.Errorf("Failed to get account: %+v", err)
		return fiber.NewError(fiber.StatusInternalServerError, "Database error")
	}

	var latestActivity model.CashActivity
	err := accountService.DB.Where("account_id = ?", account.ID).Order("created_at desc").First(&latestActivity).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		accountService.DB.Rollback()
		accountService.Log.Errorf("Failed to get latest cash activity: %+v", err)
		return fiber.NewError(fiber.StatusInternalServerError, "Database error during getting latest cash activity")
	}
	var refID *uint
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		refID = &latestActivity.ID
	}

	newActivity := model.CashActivity{
		AccountID:     account.ID,
		ReferenceID:   refID,
		Type:          "credit",
		Nominal:       req.Nominal,
		BalanceBefore: account.Balance,
		BalanceAfter:  account.Balance + req.Nominal,
	}

	tx := accountService.DB.Begin()
	if err := tx.Create(&newActivity).Error; err != nil {
		accountService.Log.Errorf("Failed to create cash activity: %+v", err)
		tx.Rollback()
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to record transaction")
	}

	account.Balance += req.Nominal
	if err := tx.Save(&account).Error; err != nil {
		accountService.Log.Errorf("Failed to update account balance: %+v", err)
		tx.Rollback()
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to record transaction")
	}

	tx.Commit()

	return nil
}

func (accountService *AccountService) Withdraw(c context.Context, req *model.Withdrawal) *fiber.Error {

	if err := accountService.Validate.Struct(req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}

	tx := accountService.DB.WithContext(c).Begin()
	if tx.Error != nil {
		accountService.Log.Errorf("Failed to start transaction: %+v", tx.Error)
		return fiber.NewError(fiber.StatusInternalServerError, "Transaction failed")
	}

	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	var account model.Account
	if err := tx.Where("account_number = ?", req.AccountNumber).First(&account).Error; err != nil {
		tx.Rollback()
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fiber.NewError(fiber.StatusNotFound, ErrAccountNotFound.Error())
		}
		accountService.Log.Errorf("Failed to get account: %+v", err)
		return fiber.NewError(fiber.StatusInternalServerError, "Database error")
	}

	if account.Balance < req.Nominal {
		tx.Rollback()
		return fiber.NewError(fiber.StatusBadRequest, ErrInsufficientBalance.Error())
	}

	var latestActivity model.CashActivity
	err := tx.Where("account_id = ?", account.ID).Order("created_at desc").First(&latestActivity).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		tx.Rollback()
		accountService.Log.Errorf("Failed to get latest cash activity: %+v", err)
		return fiber.NewError(fiber.StatusInternalServerError, "Database error during getting latest cash activity")
	}
	var refID *uint
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		refID = &latestActivity.ID
	}

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
		accountService.Log.Errorf("Failed to create cash activity: %+v", err)
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to record transaction")
	}

	account.Balance -= req.Nominal
	if err := tx.Save(&account).Error; err != nil {
		accountService.Log.Errorf("Failed to update account balance: %+v", err)
		tx.Rollback()
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to record transaction")
	}

	if err := tx.Commit().Error; err != nil {
		accountService.Log.Errorf("Failed to commit transaction: %+v", err)
		return fiber.NewError(fiber.StatusInternalServerError, "Transaction failed")
	}
	return nil
}

func (accountService *AccountService) GetBalance(c context.Context, accountNumber string) (*model.Account, *fiber.Error) {
	var account model.Account
	if err := accountService.DB.WithContext(c).Where("account_number = ?", accountNumber).First(&account).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fiber.NewError(fiber.StatusNotFound, ErrAccountNotFound.Error())
		}
		accountService.Log.Errorf("Failed to get account: %+v", err)
		return nil, fiber.NewError(fiber.StatusInternalServerError, "Database error")
	}
	return &account, nil
}
