package service_test

import (
	"account-service/src/model"
	"account-service/src/service"
	"context"
	"errors"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/go-playground/validator/v10"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// setupMockDB sets up a mock database connection using sqlmock.
func setupMockDB() (*gorm.DB, sqlmock.Sqlmock, error) {
	// Use sqlmock.New() to create a mock database connection
	db, mock, err := sqlmock.New()
	if err != nil {
		return nil, nil, err
	}

	// Use gorm.Open with postgres.New to create a GORM DB instance connected to the mock DB
	dialector := postgres.New(postgres.Config{
		Conn:       db,
		DriverName: "postgres",
	})

	gormDB, err := gorm.Open(dialector, &gorm.Config{})
	if err != nil {
		return nil, nil, err
	}

	return gormDB, mock, nil
}

// TestCreateAccount tests the CreateAccount method.
func TestCreateAccount(t *testing.T) {
	db, mockDB, err := setupMockDB() // Use the mock DB setup
	assert.NoError(t, err)

	validate := validator.New()
	serviceInstance := service.NewAccountService(db, validate)
	mockLogger := logrus.New() // Mock logger
	serviceInstance.(*service.AccountService).Log = mockLogger

	tests := []struct {
		name          string
		request       *model.CreateAccount
		expectedError error
		setupMock     func(mock sqlmock.Sqlmock) // Use sqlmock for setup
	}{
		{
			name: "Successful account creation",
			request: &model.CreateAccount{
				FullName:    "John Doe",
				IDNumber:    "1234567890123456",
				PhoneNumber: "081234567890",
			},
			expectedError: nil,
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "accounts" WHERE id_number = $1 ORDER BY "accounts"."id" LIMIT $2`)).
					WithArgs("1234567890123456", 1).
					WillReturnError(gorm.ErrRecordNotFound)
				mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "accounts" WHERE phone_number = $1 ORDER BY "accounts"."id" LIMIT $2`)).
					WithArgs("081234567890", 1).
					WillReturnError(gorm.ErrRecordNotFound)

				mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "accounts" WHERE account_number = $1 ORDER BY "accounts"."id" LIMIT $2`)).
					WithArgs(sqlmock.AnyArg(), 1).
					WillReturnError(gorm.ErrRecordNotFound)

				mock.ExpectBegin()
				mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "accounts"`)).
					WithArgs(sqlmock.AnyArg(), "John Doe", "1234567890123456", "081234567890", 0.0, sqlmock.AnyArg(), sqlmock.AnyArg()).
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
				mock.ExpectCommit()
			},
		},
		{
			name: "Duplicate ID number",
			request: &model.CreateAccount{
				FullName:    "Jane Doe",
				IDNumber:    "1234567890123456",
				PhoneNumber: "081234567891",
			},
			expectedError: service.ErrDuplicateIDNumber,
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "accounts" WHERE id_number = $1 ORDER BY "accounts"."id" LIMIT $2`)).
					WithArgs("1234567890123456", 1).
					WillReturnRows(sqlmock.NewRows([]string{"id", "account_number", "full_name", "id_number", "phone_number", "balance"}).AddRow(1, "acc1", "John Doe", "1234567890123456", "081234567890", 0.0))
			},
		},
		{
			name: "Duplicate Phone Number",
			request: &model.CreateAccount{
				FullName:    "John Smith",
				IDNumber:    "1111111111111111",
				PhoneNumber: "081234567890",
			},
			expectedError: service.ErrDuplicatePhoneNumber,
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "accounts" WHERE id_number = $1 ORDER BY "accounts"."id" LIMIT $2`)).
					WithArgs("1111111111111111", 1).
					WillReturnError(gorm.ErrRecordNotFound)

				mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "accounts" WHERE phone_number = $1 ORDER BY "accounts"."id" LIMIT $2`)).
					WithArgs("081234567890", 1).
					WillReturnRows(sqlmock.NewRows([]string{"id", "account_number", "full_name", "id_number", "phone_number", "balance"}).AddRow(1, "acc1", "John Doe", "1111111111111111", "081234567890", 0.0))
			},
		},
		{
			name:          "Validation Error",
			request:       &model.CreateAccount{IDNumber: "2222222222222222", PhoneNumber: "081234567892"}, // Missing FullName
			expectedError: errors.New("validation"),                                                        // Expect validation error
			setupMock:     func(mock sqlmock.Sqlmock) {},
		},
		{
			name: "Database Error (Create)",
			request: &model.CreateAccount{
				FullName:    "John Doe",
				IDNumber:    "3333333333333333",
				PhoneNumber: "081234567894",
			},
			expectedError: errors.New("failed to create account"), // Expect wrapped error
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "accounts" WHERE id_number = $1 ORDER BY "accounts"."id" LIMIT $2`)).
					WithArgs("3333333333333333", 1).
					WillReturnError(gorm.ErrRecordNotFound)
				mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "accounts" WHERE phone_number = $1 ORDER BY "accounts"."id" LIMIT $2`)).
					WithArgs("081234567894", 1).
					WillReturnError(gorm.ErrRecordNotFound)

				mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "accounts" WHERE account_number = $1 ORDER BY "accounts"."id" LIMIT $2`)).
					WithArgs(sqlmock.AnyArg(), 1).
					WillReturnError(gorm.ErrRecordNotFound)

				mock.ExpectBegin()
				mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "accounts"`)).
					WithArgs(sqlmock.AnyArg(), "John Doe", "3333333333333333", "081234567894", 0.0, sqlmock.AnyArg(), sqlmock.AnyArg()).
					WillReturnError(errors.New("some database error"))
				mock.ExpectRollback()
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mock expectations
			if tt.setupMock != nil {
				tt.setupMock(mockDB)
			}

			// Use a valid context.
			ctx := context.Background()

			// Call the service method.  Pass the context.
			createdAccount, err := serviceInstance.CreateAccount(ctx, tt.request)

			// Assertions
			if tt.expectedError != nil {
				assert.ErrorContains(t, err, tt.expectedError.Error()) // Use ErrorContains
			} else {
				assert.NoError(t, err, "Should not return an error")
				assert.NotNil(t, createdAccount, "Created account should not be nil")
				assert.NotEmpty(t, createdAccount.AccountNumber, "Account number should be generated")
				assert.Equal(t, tt.request.FullName, createdAccount.FullName, "Full name should match")
				assert.Equal(t, tt.request.IDNumber, createdAccount.IDNumber, "ID number should match")
				assert.Equal(t, tt.request.PhoneNumber, createdAccount.PhoneNumber, "Phone number should match")
			}

			// Verify that all expectations were met
			assert.NoError(t, mockDB.ExpectationsWereMet())
		})
	}
}

func TestDeposit(t *testing.T) {
	db, mockDB, err := setupMockDB() // Use the mock DB setup
	assert.NoError(t, err)

	validate := validator.New()
	serviceInstance := service.NewAccountService(db, validate)

	mockLogger := logrus.New() // Mock logger
	serviceInstance.(*service.AccountService).Log = mockLogger

	tests := []struct {
		name          string
		request       *model.DepositRequest
		expectedError error
		setupMock     func(mock sqlmock.Sqlmock)
	}{
		{
			name: "Successful deposit",
			request: &model.DepositRequest{
				AccountNumber: "12345678",
				Nominal:       100,
			},
			expectedError: nil,
			setupMock: func(mock sqlmock.Sqlmock) {
				// Simulate finding the account
				mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "accounts" WHERE account_number = $1 ORDER BY "accounts"."id" LIMIT $2`)).
					WithArgs("12345678", 1).
					WillReturnRows(sqlmock.NewRows([]string{"id", "account_number", "full_name", "id_number", "phone_number", "balance"}).
						AddRow(1, "12345678", "Test User", "1122334455667788", "081111111111", 0))

				// Simulate creating the cash activity record
				mock.ExpectBegin() // Start a transaction
				mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO "cash_activities" ("account_id","reference_id","type","nominal","balance_before","balance_after","description","created_at") VALUES ($1,$2,$3,$4,$5,$6,$7,$8) RETURNING "id"`)).
					WithArgs(1, nil, "credit", 100.0, 0.0, 100.0, "", sqlmock.AnyArg()).
					WillReturnResult(sqlmock.NewResult(2, 1)) // Return a dummy activity ID
				mock.ExpectCommit() // Commit the transaction
			},
		},
		{
			name: "Account not found",
			request: &model.DepositRequest{
				AccountNumber: "99999999",
				Nominal:       100,
			},
			expectedError: service.ErrAccountNotFound,
			setupMock: func(mock sqlmock.Sqlmock) {
				// Simulate account not found
				mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "accounts" WHERE account_number = $1 ORDER BY "accounts"."id" LIMIT $2`)).
					WithArgs("99999999", 1).
					WillReturnError(gorm.ErrRecordNotFound)
			},
		},
		{
			name: "Invalid request (negative amount)",
			request: &model.DepositRequest{
				AccountNumber: "12345678",
				Nominal:       -50,
			},
			expectedError: errors.New("Key: 'DepositRequest.Nominal' Error:Field validation for 'Nominal' failed on the 'gt' tag"),
			setupMock:     func(mock sqlmock.Sqlmock) {}, // No DB calls
		},
		{
			name: "Database error during get account",
			request: &model.DepositRequest{
				AccountNumber: "12345678",
				Nominal:       100,
			},
			expectedError: errors.New("database error"),
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "accounts" WHERE account_number = $1 ORDER BY "accounts"."id" LIMIT $2`)).
					WithArgs("12345678", 1).
					WillReturnError(errors.New("database error"))
			},
		},
		{
			name: "Database error during create activity",
			request: &model.DepositRequest{
				AccountNumber: "12345678",
				Nominal:       100,
			},
			expectedError: errors.New("failed to record transaction"),
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "accounts" WHERE account_number = $1 ORDER BY "accounts"."id" LIMIT $2`)).
					WithArgs("12345678", 1).
					WillReturnRows(sqlmock.NewRows([]string{"id", "account_number", "full_name", "id_number", "phone_number", "balance"}).
						AddRow(1, "12345678", "Test User", "1122334455667788", "081111111111", 0))

				mock.ExpectBegin() // Start a transaction
				mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO "cash_activities" ("account_id","reference_id","type","nominal","balance_before","balance_after","description","created_at") VALUES ($1,$2,$3,$4,$5,$6,$7,$8) RETURNING "id"`)).
					WithArgs(1, nil, "credit", 100.0, 0.0, 100.0, "", sqlmock.AnyArg()).
					WillReturnError(errors.New("some database error"))
				mock.ExpectRollback() // Rollback the transaction
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mock expectations
			if tt.setupMock != nil {
				tt.setupMock(mockDB)
			}

			ctx := context.Background()
			err := serviceInstance.Deposit(ctx, tt.request)

			if tt.expectedError != nil {
				assert.ErrorContains(t, err, tt.expectedError.Error(), "Error message should match")
			} else {
				assert.NoError(t, err, "Should not return an error")
			}

			assert.NoError(t, mockDB.ExpectationsWereMet())
		})
	}
}

func TestWithdraw(t *testing.T) {
	db, mockDB, err := setupMockDB() // Use the mock DB setup
	assert.NoError(t, err)

	validate := validator.New()
	serviceInstance := service.NewAccountService(db, validate)
	mockLogger := logrus.New() // Mock logger
	serviceInstance.(*service.AccountService).Log = mockLogger

	tests := []struct {
		name          string
		request       *model.Withdrawal
		expectedError error
		setupMock     func(mock sqlmock.Sqlmock) // Use sqlmock for setup
	}{
		{
			name: "Successful withdrawal",
			request: &model.Withdrawal{
				AccountNumber: "78901234",
				Nominal:       200,
			},
			expectedError: nil,
			setupMock: func(mock sqlmock.Sqlmock) {

				// Simulate finding the account with sufficient balance
				mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "accounts" WHERE account_number = $1 ORDER BY "accounts"."id" LIMIT $2`)).
					WithArgs("78901234", 1).
					WillReturnRows(sqlmock.NewRows([]string{"id", "account_number", "full_name", "id_number", "phone_number", "balance"}).
						AddRow(1, "78901234", "Test User", "8765432109876543", "082222222222", 1000))

				//Simulate no previous activity
				mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "cash_activities" WHERE account_id = $1 ORDER BY created_at DESC LIMIT $2`)).
					WithArgs(1, 1).
					WillReturnError(gorm.ErrRecordNotFound)

				// Simulate creating the cash activity record
				mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "cash_activities"`)).
					WithArgs(1, nil, "debit", 200.0, 1000.0, 800.0, "", sqlmock.AnyArg()).
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(2))

			},
		},
		{
			name: "Account not found",
			request: &model.Withdrawal{
				AccountNumber: "99999999",
				Nominal:       100,
			},
			expectedError: service.ErrAccountNotFound,
			setupMock: func(mock sqlmock.Sqlmock) {

				// Simulate account not found
				mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "accounts" WHERE account_number = $1 ORDER BY "accounts"."id" LIMIT $2`)).
					WithArgs("99999999", 1).
					WillReturnError(gorm.ErrRecordNotFound)
			},
		},
		{
			name: "Insufficient balance",
			request: &model.Withdrawal{
				AccountNumber: "78901234",
				Nominal:       2000, // More than the balance
			},
			expectedError: service.ErrInsufficientBalance,
			setupMock: func(mock sqlmock.Sqlmock) {
				//Simulate account exist
				mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "accounts" WHERE account_number = $1 ORDER BY "accounts"."id" LIMIT $2`)).
					WithArgs("78901234", 1).
					WillReturnRows(sqlmock.NewRows([]string{"id", "account_number", "full_name", "id_number", "phone_number", "balance"}).
						AddRow(1, "78901234", "Test User", "8765432109876543", "082222222222", 1000))
			},
		},
		{
			name: "Invalid request (negative amount)",
			request: &model.Withdrawal{
				AccountNumber: "78901234",
				Nominal:       -50,
			},
			expectedError: errors.New("validation error"), // Expect validation error
			setupMock:     func(mock sqlmock.Sqlmock) {},
		},
		{
			name: "Database error during get account",
			request: &model.Withdrawal{
				AccountNumber: "78901234",
				Nominal:       100,
			},
			expectedError: errors.New("database error"), //Expect wrapped error
			setupMock: func(mock sqlmock.Sqlmock) {

				mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "accounts" WHERE account_number = $1 ORDER BY "accounts"."id" LIMIT $2`)).
					WithArgs("78901234", 1).
					WillReturnError(errors.New("some database error")) // Simulate a generic DB error

			},
		},
		{
			name: "Database error during create activity",
			request: &model.Withdrawal{
				AccountNumber: "78901234",
				Nominal:       100,
			},
			expectedError: errors.New("failed to record transaction"), //Expect wrapped error
			setupMock: func(mock sqlmock.Sqlmock) {

				// Simulate finding the account with sufficient balance
				mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "accounts" WHERE account_number = $1 ORDER BY "accounts"."id" LIMIT $2`)).
					WithArgs("78901234", 1).
					WillReturnRows(sqlmock.NewRows([]string{"id", "account_number", "full_name", "id_number", "phone_number", "balance"}).
						AddRow(1, "78901234", "Test User", "8765432109876543", "082222222222", 1000))

				//Simulate no previous activity
				mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "cash_activities" WHERE account_id = $1 ORDER BY created_at DESC LIMIT $2`)).
					WithArgs(1, 1).
					WillReturnError(gorm.ErrRecordNotFound)

				// Simulate creating the cash activity record
				mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "cash_activities"`)).
					WithArgs(1, nil, "debit", 100.0, 1000.0, 900.0, "", sqlmock.AnyArg()).
					WillReturnError(errors.New("some database error"))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setupMock != nil {
				tt.setupMock(mockDB)
			}
			ctx := context.Background()
			err := serviceInstance.Withdraw(ctx, tt.request)

			if tt.expectedError != nil {
				assert.ErrorContains(t, err, tt.expectedError.Error(), "Error message should match")
			} else {
				assert.NoError(t, err, "Should not return an error")
			}
			assert.NoError(t, mockDB.ExpectationsWereMet())
		})
	}
}

func TestGetBalance(t *testing.T) {
	db, mockDB, err := setupMockDB() // Use the mock DB setup
	assert.NoError(t, err)

	validate := validator.New()
	serviceInstance := service.NewAccountService(db, validate)
	mockLogger := logrus.New() // Mock logger
	serviceInstance.(*service.AccountService).Log = mockLogger

	tests := []struct {
		name            string
		accountNumber   string
		expectedBalance float64
		expectedError   error
		setupMock       func(mock sqlmock.Sqlmock) // Use sqlmock for setup
	}{
		{
			name:            "Successful balance retrieval",
			accountNumber:   "56789012",
			expectedBalance: 500,
			expectedError:   nil,
			setupMock: func(mock sqlmock.Sqlmock) {
				//Simulate Account Exist
				mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "accounts" WHERE account_number = $1 ORDER BY "accounts"."id" LIMIT $2`)).
					WithArgs("56789012").
					WillReturnRows(sqlmock.NewRows([]string{"id", "account_number", "full_name", "id_number", "phone_number", "balance"}).
						AddRow(1, "56789012", "Balance Test User", "5555555555555555", "085555555555", 500))
			},
		},
		{
			name:          "Account not found",
			accountNumber: "99999999",
			expectedError: service.ErrAccountNotFound,
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "accounts" WHERE account_number = $1 ORDER BY "accounts"."id" LIMIT $2`)).
					WithArgs("99999999").
					WillReturnError(gorm.ErrRecordNotFound)
			},
		},
		{
			name:          "Database Error",
			accountNumber: "56789012",
			expectedError: errors.New("database error"), // Expect wrapped error
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "accounts" WHERE account_number = $1 ORDER BY "accounts"."id" LIMIT $2`)).
					WithArgs("56789012").
					WillReturnError(errors.New("some database error")) // Simulate a generic DB error
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setupMock != nil {
				tt.setupMock(mockDB)
			}
			ctx := context.Background()
			account, err := serviceInstance.GetBalance(ctx, tt.accountNumber)

			if tt.expectedError != nil {
				assert.ErrorIs(t, err, tt.expectedError, "Error message should match") // Use ErrorIs
				assert.Nil(t, account)
			} else {
				assert.NoError(t, err, "Should not return error")
				assert.NotNil(t, account, "Account should not be nil")
				assert.Equal(t, tt.expectedBalance, account.Balance, "Balance should match")
			}
			assert.NoError(t, mockDB.ExpectationsWereMet())
		})
	}
}

// Add tests for Mutation when you implement it.
