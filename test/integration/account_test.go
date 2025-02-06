package integration

import (
	"account-service/src/config"
	"account-service/src/controller" // Import your controller
	"errors"
	"log"

	// Import your database package
	"account-service/src/model"
	"account-service/src/response"
	"account-service/src/service"
	"account-service/src/utils"
	"account-service/test/fixture"
	"account-service/test/helper"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var (
	app           *fiber.App
	db            *gorm.DB // Database connection for tests
	accountNumber string
)

// setup sets up the test environment before *all* tests in the package.
func setup() {
	// Replace with your actual test database connection details
	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%d sslmode=disable TimeZone=Asia/Shanghai",
		config.DBHost, config.DBUser, config.DBPassword, config.DBName, config.DBPort,
	)
	err := errors.New("test")
	db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger:                 logger.Default.LogMode(logger.Info),
		SkipDefaultTransaction: true,
		PrepareStmt:            true,
		TranslateError:         true,
	})
	if err != nil {
		utils.Log.Errorf("Failed to connect to database: %+v", err)
	}
	app = helper.NewTestServer(db) // Create a Fiber app instance

	validate := validator.New()
	accountService := service.NewAccountService(db, validate) //Use DB
	accountController := controller.NewAccountController(accountService, validate)

	//Define routes
	app.Post("/daftar", accountController.Register)
	app.Post("/tabung", accountController.Deposit)
	app.Post("/tarik", accountController.Withdrawal)
	app.Get("/saldo/:accountNumber", accountController.GetBalance)

	helper.ClearAll(db) // Clean the database before running tests.
}

// TestMain is the entry point for running tests in this package.
func TestMain(m *testing.M) {
	setup()
	code := m.Run() // Run the tests
	// teardown() // Optionally clean up after *all* tests (e.g., drop the test database)
	os.Exit(code)
}

func TestRegisterAccount_Success(t *testing.T) {
	helper.ClearAll(db)

	requestBody := fixture.ValidCreateAccount //Valid Request
	req, _ := json.Marshal(&requestBody)
	resp, err := helper.MakeRequest(app, http.MethodPost, "/v1/daftar", string(req), nil)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	// Read and parse the response body
	body, err := io.ReadAll(resp.Body)
	assert.NoError(t, err)
	defer resp.Body.Close()

	var apiResponse response.SuccessWithData
	err = json.Unmarshal(body, &apiResponse)
	assert.NoError(t, err)

	assert.Equal(t, fiber.StatusCreated, apiResponse.Code)
	assert.Equal(t, "success", apiResponse.Status)
	assert.NotEmpty(t, apiResponse.Data)

	// Type assertion!  This is the key part.
	accountData, ok := apiResponse.Data.(map[string]interface{})
	assert.True(t, ok, "Data should be a map[string]interface{}")

	// Now you can access fields from accountData
	accountNumber, ok := accountData["account_number"].(string)
	assert.True(t, ok, "account_number should be a string")
	assert.NotEmpty(t, accountNumber)

	//Check in DB
	var createdAccount model.Account
	err = db.Where("account_number = ?", accountNumber).First(&createdAccount).Error
	assert.NoError(t, err)
	assert.Equal(t, fixture.ValidCreateAccount.FullName, createdAccount.FullName)
	assert.Equal(t, fixture.ValidCreateAccount.IDNumber, createdAccount.IDNumber)
	assert.Equal(t, fixture.ValidCreateAccount.PhoneNumber, createdAccount.PhoneNumber)

	helper.ClearAll(db) //Clear all data
}

func TestRegisterAccount_DuplicateID(t *testing.T) {
	helper.ClearAll(db)

	// Create an account with the same ID number first.
	existingAccount := model.Account{
		FullName:      "Existing User",
		IDNumber:      "1234567890123456", // Same ID
		PhoneNumber:   "081111111111",
		AccountNumber: accountNumber,
	}
	err := helper.CreateTestAccount(db, &existingAccount) // Use helper function
	assert.NoError(t, err)

	requestBody, _ := json.Marshal(model.CreateAccount{
		FullName:    "New User",
		IDNumber:    "1234567890123456", // Duplicate ID
		PhoneNumber: "081234567890",
	})

	resp, err := helper.MakeRequest(app, http.MethodPost, "/v1/daftar", string(requestBody), nil)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusConflict, resp.StatusCode) // Expect a conflict

	body, err := io.ReadAll(resp.Body)
	assert.NoError(t, err)
	defer resp.Body.Close()
	log.Println(string(body) + "TTTTTT")

	var errorResponse response.ErrorDetails
	err = json.Unmarshal(body, &errorResponse)
	assert.NoError(t, err)
	assert.Equal(t, service.ErrDuplicateIDNumber.Error(), errorResponse.Message)

	helper.ClearAll(db) //Clear all data

}

func TestRegisterAccount_DuplicatePhoneNumber(t *testing.T) {
	helper.ClearAll(db)

	existingAccount := model.Account{
		FullName:      "Existing User",
		IDNumber:      "1111111111111111",
		PhoneNumber:   "081234567890", // Same Phone Number
		AccountNumber: accountNumber,
	}
	err := helper.CreateTestAccount(db, &existingAccount) // Use helper function
	assert.NoError(t, err)

	requestBody, _ := json.Marshal(model.CreateAccount{
		FullName:    "New User",
		IDNumber:    "1234567890123456",
		PhoneNumber: "081234567890", // Duplicate Phone Number
	})

	resp, err := helper.MakeRequest(app, http.MethodPost, "/v1/daftar", string(requestBody), nil)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusConflict, resp.StatusCode) // Expect a conflict

	body, err := io.ReadAll(resp.Body)
	assert.NoError(t, err)
	defer resp.Body.Close()

	var errorResponse response.ErrorDetails
	err = json.Unmarshal(body, &errorResponse)
	assert.NoError(t, err)
	assert.Equal(t, service.ErrDuplicatePhoneNumber.Error(), errorResponse.Message)
	helper.ClearAll(db)
}
func TestRegisterAccount_ValidationError(t *testing.T) {
	helper.ClearAll(db)
	invalidRequest := model.CreateAccount{ // No FullName
		IDNumber:    "1234567890123456",
		PhoneNumber: "081234567890",
	}
	requestBody, _ := json.Marshal(invalidRequest)

	resp, err := helper.MakeRequest(app, http.MethodPost, "/v1/daftar", string(requestBody), nil)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	assert.NoError(t, err)
	defer resp.Body.Close()

	var errorResponse response.ErrorDetails
	err = json.Unmarshal(body, &errorResponse)
	assert.NoError(t, err)
	assert.Contains(t, errorResponse.Message, "FullName") // Check for field name in error

	helper.ClearAll(db)
}

func TestDeposit_Success(t *testing.T) {
	helper.ClearAll(db)

	// 1. Create an account to deposit into.
	existingAccount := model.Account{
		FullName:      "Deposit Test User",
		IDNumber:      "1122334455667788",
		PhoneNumber:   "081122334455",
		AccountNumber: utils.GenerateAccountNumber(), // Use helper, ensure uniqueness
		Balance:       0,                             // Initial balance
	}
	err := helper.CreateTestAccount(db, &existingAccount)
	assert.NoError(t, err)

	// 2. Prepare the deposit request.
	depositRequest := model.DepositRequest{
		AccountNumber: existingAccount.AccountNumber, // Use the created account's number
		Nominal:       500000,
	}
	requestBody, _ := json.Marshal(depositRequest)

	// 3. Make the deposit request.
	resp, err := helper.MakeRequest(app, http.MethodPost, "/v1/tabung", string(requestBody), nil)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// 4. Read and parse the response body.
	body, err := io.ReadAll(resp.Body)
	assert.NoError(t, err)
	defer resp.Body.Close()

	var apiResponse response.SuccessWithData
	err = json.Unmarshal(body, &apiResponse)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, apiResponse.Code)
	assert.Equal(t, "success", apiResponse.Status)

	// 5. Type assert the balance.
	accountData, ok := apiResponse.Data.(map[string]interface{})
	assert.True(t, ok, "Data should be a map[string]interface{}")
	// Now you can access fields from accountData
	balance, ok := accountData["saldo"].(float64)
	assert.True(t, ok, "Data should be a float64 (balance)")
	assert.Equal(t, 500000.0, balance)

	// 6.  Verify the account balance in the database.
	var updatedAccount model.Account
	err = db.Where("account_number = ?", existingAccount.AccountNumber).First(&updatedAccount).Error
	assert.NoError(t, err)
	assert.Equal(t, depositRequest.Nominal, updatedAccount.Balance)

	//7. Verify that cash activity also created
	var cashActivity model.CashActivity
	err = db.Where("account_id = ?", updatedAccount.ID).Order("created_at desc").First(&cashActivity).Error //get the latest
	assert.NoError(t, err)
	assert.Equal(t, "credit", cashActivity.Type)
	assert.Equal(t, depositRequest.Nominal, cashActivity.Nominal)
	assert.Equal(t, 0.0, cashActivity.BalanceBefore) // Initial balance was 0
	assert.Equal(t, depositRequest.Nominal, cashActivity.BalanceAfter)
	assert.Nil(t, cashActivity.ReferenceID) //Should be nil (first transaction)

	helper.ClearAll(db)
}

func TestDeposit_AccountNotFound(t *testing.T) {
	helper.ClearAll(db)

	depositRequest := model.DepositRequest{
		AccountNumber: "9999999999", // Non-existent account
		Nominal:       500000,
	}
	requestBody, _ := json.Marshal(depositRequest)

	resp, err := helper.MakeRequest(app, http.MethodPost, "/v1/tabung", string(requestBody), nil)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, resp.StatusCode) // Expect 404 Not Found

	body, err := io.ReadAll(resp.Body)
	assert.NoError(t, err)
	defer resp.Body.Close()

	var errorResponse response.ErrorDetails
	err = json.Unmarshal(body, &errorResponse)
	assert.NoError(t, err)
	assert.Equal(t, service.ErrAccountNotFound.Error(), errorResponse.Message)

	helper.ClearAll(db)
}

func TestDeposit_ValidationError(t *testing.T) {
	helper.ClearAll(db)

	depositRequest := model.DepositRequest{
		AccountNumber: "1234567890",
		Nominal:       -100, //Invalid Deposit
	}
	requestBody, _ := json.Marshal(depositRequest)

	resp, err := helper.MakeRequest(app, http.MethodPost, "/v1/tabung", string(requestBody), nil)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	assert.NoError(t, err)
	defer resp.Body.Close()

	var errorResponse response.ErrorDetails
	err = json.Unmarshal(body, &errorResponse)
	assert.NoError(t, err)
	assert.Contains(t, errorResponse.Message, "Nominal")

	helper.ClearAll(db)
}

func TestWithdrawal_Success(t *testing.T) {
	helper.ClearAll(db)
	// 1. Create an account
	initialBalance := 1000000.0 // Start with some balance
	existingAccount := model.Account{
		FullName:      "Withdrawal Test User",
		IDNumber:      "2233445566778899",
		PhoneNumber:   "082233445566",
		AccountNumber: utils.GenerateAccountNumber(),
		Balance:       initialBalance,
	}
	err := helper.CreateTestAccount(db, &existingAccount)
	assert.NoError(t, err)

	// 2. Prepare the withdrawal request
	withdrawalRequest := model.Withdrawal{
		AccountNumber: existingAccount.AccountNumber, // Use the created account's number
		Nominal:       250000,
	}
	requestBody, _ := json.Marshal(withdrawalRequest)

	// 3. Make the withdrawal request
	resp, err := helper.MakeRequest(app, http.MethodPost, "/v1/tarik", string(requestBody), nil)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// 4. Read/parse the response
	body, err := io.ReadAll(resp.Body)
	assert.NoError(t, err)
	defer resp.Body.Close()

	var apiResponse response.SuccessWithData
	err = json.Unmarshal(body, &apiResponse)
	assert.NoError(t, err)

	// 5. Assert on the response
	accountData, ok := apiResponse.Data.(map[string]interface{})
	assert.True(t, ok, "Data should be a map[string]interface{}")
	// Now you can access fields from accountData
	balance, ok := accountData["saldo"].(float64)
	assert.True(t, ok, "Data should be a float64 (balance)")
	assert.Equal(t, initialBalance-withdrawalRequest.Nominal, balance)

	// 6. Verify account balance in DB
	var updatedAccount model.Account
	err = db.Where("account_number = ?", existingAccount.AccountNumber).First(&updatedAccount).Error
	assert.NoError(t, err)
	assert.Equal(t, initialBalance-withdrawalRequest.Nominal, updatedAccount.Balance)

	//7. Verify that cash activity also created
	var cashActivity model.CashActivity
	err = db.Where("account_id = ?", updatedAccount.ID).Last(&cashActivity).Error //get the latest
	assert.NoError(t, err)
	assert.Equal(t, "debit", cashActivity.Type)
	assert.Equal(t, withdrawalRequest.Nominal, cashActivity.Nominal)
	assert.Equal(t, initialBalance, cashActivity.BalanceBefore) // Initial balance was 0
	assert.Equal(t, initialBalance-withdrawalRequest.Nominal, cashActivity.BalanceAfter)
	assert.Nil(t, cashActivity.ReferenceID) //Should be nil (first transaction)

	helper.ClearAll(db) //Clear all data
}
func TestWithdrawal_AccountNotFound(t *testing.T) {
	helper.ClearAll(db)
	withdrawalRequest := model.Withdrawal{
		AccountNumber: "9999999999", // Non-existent account
		Nominal:       500000,
	}
	requestBody, _ := json.Marshal(withdrawalRequest)

	resp, err := helper.MakeRequest(app, http.MethodPost, "/v1/tarik", string(requestBody), nil)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, resp.StatusCode) // Expect 404

	body, err := io.ReadAll(resp.Body)
	assert.NoError(t, err)
	defer resp.Body.Close()

	var errorResponse response.ErrorDetails
	err = json.Unmarshal(body, &errorResponse)
	assert.NoError(t, err)
	assert.Equal(t, service.ErrAccountNotFound.Error(), errorResponse.Message)
	helper.ClearAll(db)
}

func TestWithdrawal_InsufficientBalance(t *testing.T) {
	helper.ClearAll(db)
	// 1. Create an account to deposit into.
	existingAccount := model.Account{
		FullName:      "Withdrawal Test User",
		IDNumber:      "3344556677889900",
		PhoneNumber:   "083344556677",
		AccountNumber: utils.GenerateAccountNumber(), // Use helper, ensure uniqueness
		Balance:       100000,                        // Initial balance
	}
	err := helper.CreateTestAccount(db, &existingAccount)
	assert.NoError(t, err)

	withdrawalRequest := model.Withdrawal{
		AccountNumber: existingAccount.AccountNumber, // Use existing account
		Nominal:       500000,                        // More than balance
	}
	requestBody, _ := json.Marshal(withdrawalRequest)

	resp, err := helper.MakeRequest(app, http.MethodPost, "/v1/tarik", string(requestBody), nil)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode) // Expect 400 Bad Request

	body, err := io.ReadAll(resp.Body)
	assert.NoError(t, err)
	defer resp.Body.Close()

	var errorResponse response.ErrorDetails
	err = json.Unmarshal(body, &errorResponse)
	assert.NoError(t, err)
	assert.Equal(t, service.ErrInsufficientBalance.Error(), errorResponse.Message)
	helper.ClearAll(db)
}
func TestWithdrawal_ValidationError(t *testing.T) {
	helper.ClearAll(db)
	withdrawalRequest := model.Withdrawal{
		AccountNumber: "123456789",
		Nominal:       -100, //Invalid
	}

	requestBody, _ := json.Marshal(withdrawalRequest)
	resp, err := helper.MakeRequest(app, http.MethodPost, "/v1/tarik", string(requestBody), nil)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	assert.NoError(t, err)
	defer resp.Body.Close()

	var errorResponse response.ErrorDetails
	err = json.Unmarshal(body, &errorResponse)
	assert.NoError(t, err)
	assert.Contains(t, errorResponse.Message, "Nominal")

	helper.ClearAll(db)
}

func TestGetBalance_Success(t *testing.T) {
	helper.ClearAll(db)

	// 1. Create an account with a known balance.
	initialBalance := 750000.0
	existingAccount := model.Account{
		FullName:      "Balance Test User",
		IDNumber:      "4455667788990011",
		PhoneNumber:   "084455667788",
		AccountNumber: utils.GenerateAccountNumber(),
		Balance:       initialBalance,
	}
	err := helper.CreateTestAccount(db, &existingAccount)
	assert.NoError(t, err)

	// 2. Make the GET request to /saldo/{accountNumber}.
	resp, err := helper.MakeRequest(app, http.MethodGet, "/v1/saldo/"+existingAccount.AccountNumber, "", nil)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// 3. Read and parse the response body.
	body, err := io.ReadAll(resp.Body)
	assert.NoError(t, err)
	defer resp.Body.Close()

	var apiResponse response.SuccessWithData
	err = json.Unmarshal(body, &apiResponse)
	assert.NoError(t, err)

	// 4. Type assert and verify the balance.
	accountData, ok := apiResponse.Data.(map[string]interface{})
	assert.True(t, ok, "Data should be a map[string]interface{}")
	// Now you can access fields from accountData
	balance, ok := accountData["saldo"].(float64)
	assert.True(t, ok, "Data should be a float64 (balance)")
	assert.Equal(t, 750000.0, balance)
	assert.Equal(t, initialBalance, balance)

	helper.ClearAll(db) // Clean up
}

func TestGetBalance_AccountNotFound(t *testing.T) {
	helper.ClearAll(db)

	resp, err := helper.MakeRequest(app, http.MethodGet, "/v1/saldo/9999999999", "", nil) // Non-existent account
	assert.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	assert.NoError(t, err)
	defer resp.Body.Close()

	var errorResponse response.ErrorDetails
	err = json.Unmarshal(body, &errorResponse)
	assert.NoError(t, err)
	assert.Equal(t, service.ErrAccountNotFound.Error(), errorResponse.Message)
	helper.ClearAll(db)
}

func TestGetBalance_InvalidAccountNumber(t *testing.T) {
	helper.ClearAll(db)

	// 1. Make the GET request to /saldo/{accountNumber} with an invalid account number.
	resp, err := helper.MakeRequest(app, http.MethodGet, "/v1/saldo/invalid-account-number", "", nil)
	assert.NoError(t, err)

	// 2. Assert on the HTTP status code.
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

	helper.ClearAll(db)
}
