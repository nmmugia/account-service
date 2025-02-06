package integration

import (
	"account-service/src/controller" // Import your controller
	"account-service/src/database"   // Import your database package
	"account-service/src/model"
	"account-service/src/response"
	"account-service/src/service"
	"account-service/test/fixture"
	"account-service/test/helper"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

var (
	app *fiber.App
	db  *gorm.DB // Database connection for tests
)

// setup sets up the test environment before *all* tests in the package.
func setup() {
	// Replace with your actual test database connection details
	db = database.Connect("localhost", "account_service_test") // Connect to your *TEST* database.
	app = helper.NewTestServer(db)                             // Create a Fiber app instance

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
	helper.ClearAll(db) //Clear all data

	requestBody := fixture.ValidCreateAccount //Valid Request

	resp, err := helper.MakeRequest(app, http.MethodPost, "/daftar", string(requestBody), nil)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	// Read and parse the response body
	body, err := io.ReadAll(resp.Body)
	assert.NoError(t, err)
	defer resp.Body.Close()

	var apiResponse response.SuccessWithAccount
	err = json.Unmarshal(body, &apiResponse)
	assert.NoError(t, err)

	// Assertions on the response body
	assert.Equal(t, fiber.StatusCreated, apiResponse.Code)
	assert.Equal(t, "success", apiResponse.Status)
	assert.NotEmpty(t, apiResponse.AccountNumber)

	// Check if the account was actually created in the database
	var createdAccount model.Account
	err = db.Where("account_number = ?", apiResponse.AccountNumber).First(&createdAccount).Error
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
		AccountNumber: helper.GenerateAccountNumber(),
	}
	err := helper.CreateTestAccount(db, &existingAccount) // Use helper function
	assert.NoError(t, err)

	requestBody, _ := json.Marshal(model.CreateAccount{
		FullName:    "New User",
		IDNumber:    "1234567890123456", // Duplicate ID
		PhoneNumber: "081234567890",
	})

	resp, err := helper.MakeRequest(app, http.MethodPost, "/daftar", string(requestBody), nil)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusConflict, resp.StatusCode) // Expect a conflict

	body, err := io.ReadAll(resp.Body)
	assert.NoError(t, err)
	defer resp.Body.Close()

	var errorResponse response.ErrorResponse
	err = json.Unmarshal(body, &errorResponse)
	assert.NoError(t, err)
	assert.Equal(t, service.ErrDuplicateIDNumber.Error(), errorResponse.Remark)

	helper.ClearAll(db) //Clear all data

}

func TestRegisterAccount_DuplicatePhoneNumber(t *testing.T) {
	helper.ClearAll(db)

	existingAccount := model.Account{
		FullName:      "Existing User",
		IDNumber:      "1111111111111111",
		PhoneNumber:   "081234567890", // Same Phone Number
		AccountNumber: helper.GenerateAccountNumber(),
	}
	err := helper.CreateTestAccount(db, &existingAccount) // Use helper function
	assert.NoError(t, err)

	requestBody, _ := json.Marshal(model.CreateAccount{
		FullName:    "New User",
		IDNumber:    "1234567890123456",
		PhoneNumber: "081234567890", // Duplicate Phone Number
	})

	resp, err := helper.MakeRequest(app, http.MethodPost, "/daftar", string(requestBody), nil)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusConflict, resp.StatusCode) // Expect a conflict

	body, err := io.ReadAll(resp.Body)
	assert.NoError(t, err)
	defer resp.Body.Close()

	var errorResponse response.ErrorResponse
	err = json.Unmarshal(body, &errorResponse)
	assert.NoError(t, err)
	assert.Equal(t, service.ErrDuplicatePhoneNumber.Error(), errorResponse.Remark)
	helper.ClearAll(db)
}
func TestRegisterAccount_ValidationError(t *testing.T) {
	helper.ClearAll(db)
	invalidRequest := model.CreateAccount{ // No FullName
		IDNumber:    "1234567890123456",
		PhoneNumber: "081234567890",
	}
	requestBody, _ := json.Marshal(invalidRequest)

	resp, err := helper.MakeRequest(app, http.MethodPost, "/daftar", string(requestBody), nil)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	assert.NoError(t, err)
	defer resp.Body.Close()

	var errorResponse response.ErrorResponse
	err = json.Unmarshal(body, &errorResponse)
	assert.NoError(t, err)
	assert.Contains(t, errorResponse.Remark, "FullName") // Check for field name in error

	helper.ClearAll(db)
}
