package helper

import (
	"account-service/src/middleware"
	"account-service/src/model"
	"account-service/src/router"
	"account-service/src/utils"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/compress"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/helmet"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// ClearAll clears all data from the account and cash_activity tables.  USE WITH CAUTION.
func ClearAll(db *gorm.DB) {
	ClearCashActivities(db)
	ClearAccounts(db)
}

// ClearAccounts deletes all accounts from the database.
func ClearAccounts(db *gorm.DB) {
	if err := db.Where("id is not null").Delete(&model.Account{}).Error; err != nil {
		logrus.Fatalf("Failed to clear account data: %+v", err)
	}
}

// ClearCashActivities deletes all cash activities from the database.
func ClearCashActivities(db *gorm.DB) {
	if err := db.Where("id is not null").Delete(&model.CashActivity{}).Error; err != nil {
		logrus.Fatalf("Failed to clear cash activity data: %+v", err)
	}
}

// CreateAccount creates a new account in the database.  It handles generating
// a unique account number.
func CreateAccount(db *gorm.DB, fullName, idNumber, phoneNumber string) (*model.Account, error) {
	accountNumber := utils.GenerateAccountNumber()
	for {
		var tempAccount model.Account
		err := db.Where("account_number = ?", accountNumber).First(&tempAccount).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			break // Account number is unique
		}
		if err != nil {
			logrus.Errorf("Error checking for unique account number: %+v", err)
			return nil, fmt.Errorf("database error: %w", err) // Wrap the error
		}
		accountNumber = utils.GenerateAccountNumber() // Regenerate if not unique
	}

	newAccount := &model.Account{
		AccountNumber: accountNumber,
		FullName:      fullName,
		IDNumber:      idNumber,
		PhoneNumber:   phoneNumber,
	}

	if err := db.Create(newAccount).Error; err != nil {
		logrus.Errorf("Failed to create account: %+v", err)
		return nil, fmt.Errorf("failed to create account: %w", err) // Wrap
	}
	return newAccount, nil
}

// InsertAccount creates multiple accounts.  Useful for seeding test data.
func InsertAccount(db *gorm.DB, accounts ...*model.Account) {
	for _, account := range accounts {

		// Generate unique account numbers
		account.AccountNumber = utils.GenerateAccountNumber() // Ensure unique account number
		for {
			var tempAccount model.Account
			err := db.Where("account_number = ?", account.AccountNumber).First(&tempAccount).Error
			if errors.Is(err, gorm.ErrRecordNotFound) {
				break
			}
			if err != nil {
				logrus.Errorf("Error checking unique account number: %+v", err)
			}
			account.AccountNumber = utils.GenerateAccountNumber()
		}
		if err := db.Create(account).Error; err != nil {
			logrus.Errorf("Failed to create account: %+v", err)
		}
	}
}

// GetAccountByNumber retrieves an account by its account number.
func GetAccountByNumber(db *gorm.DB, accountNumber string) (*model.Account, error) {
	var account model.Account
	if err := db.Where("account_number = ?", accountNumber).First(&account).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("account not found: %w", err) // Wrap
		}
		logrus.Errorf("Failed to get account by number: %+v", err)
		return nil, fmt.Errorf("database error: %w", err) // Wrap
	}
	return &account, nil
}

// CreateCashActivity adds a new cash activity record to the database.
func CreateCashActivity(db *gorm.DB, accountID uint, referenceID *uint, activityType string, nominal, balanceBefore, balanceAfter float64, description string) (*model.CashActivity, error) {
	newActivity := &model.CashActivity{
		AccountID:     accountID,
		ReferenceID:   referenceID,
		Type:          activityType,
		Nominal:       nominal,
		BalanceBefore: balanceBefore,
		BalanceAfter:  balanceAfter,
		Description:   description,
	}
	if err := db.Create(newActivity).Error; err != nil {
		logrus.Errorf("Failed to create cash activity: %+v", err)
		return nil, fmt.Errorf("failed to create cash activity: %w", err) // Wrap
	}

	return newActivity, nil
}

// UpdateAccountBalance updates the balance of a given account.
func UpdateAccountBalance(db *gorm.DB, accountID uint, newBalance float64) error {
	if err := db.Model(&model.Account{}).Where("id = ?", accountID).Update("balance", newBalance).Error; err != nil {
		logrus.Errorf("Failed to update account balance: %+v", err)
		return fmt.Errorf("failed to update account balance: %w", err)
	}
	return nil
}

// GetLatestCashActivity retrieves the latest cash activity for a given account ID.
func GetLatestCashActivity(db *gorm.DB, accountID uint) (*model.CashActivity, error) {
	var latestActivity model.CashActivity
	err := db.Where("account_id = ?", accountID).Order("created_at desc").First(&latestActivity).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		logrus.Errorf("Failed to get latest cash activity: %+v", err)
		return nil, fmt.Errorf("failed to get latest cash activity: %w", err) // Consistent wrapping
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil // No error, just no activity yet.  This is important!
	}
	return &latestActivity, nil
}

// MakeRequest is a helper function to make HTTP requests to the test server.
func MakeRequest(app *fiber.App, method, path, body string, headers map[string]string) (*http.Response, error) {
	req := httptest.NewRequest(method, path, strings.NewReader(body))

	// Set default content type
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	// Add any custom headers
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	return app.Test(req) // Use app.Test() for Fiber
}

func NewTestServer(db *gorm.DB) *fiber.App {
	app := fiber.New() // Use default Fiber config, or a test-specific one if needed.

	// Apply middleware (IMPORTANT: mirror your main.go setup)
	app.Use("/v1", middleware.LimiterConfig())
	app.Use(middleware.LoggerConfig())
	app.Use(helmet.New())
	app.Use(compress.New())
	app.Use(cors.New())
	app.Use(middleware.RecoverConfig())

	router.Routes(app, db)         // Use the same router setup as your main app
	app.Use(utils.NotFoundHandler) // and not found handler

	return app
}

// CreateTestAccount creates an account directly in the DB for testing.
func CreateTestAccount(db *gorm.DB, account *model.Account) error {
	if err := db.Create(account).Error; err != nil {
		return fmt.Errorf("failed to create test account: %w", err)
	}
	return nil
}
