package controller

import (
	"account-service/src/model"
	"account-service/src/response"
	"account-service/src/service"
	"strconv"

	"github.com/go-playground/validator/v10"

	"github.com/gofiber/fiber/v2"
)

type AccountController struct {
	AccountService service.AccountServices
	Validator      *validator.Validate
}

func NewAccountController(accountService service.AccountServices, validator *validator.Validate) *AccountController {
	return &AccountController{
		AccountService: accountService,
		Validator:      validator,
	}
}

// @Tags         Accounts
// @Summary      Register a new customer (Nasabah)
// @Description  API for registering a new customer.
// @Accept       json
// @Produce      json
// @Param        request  body  model.CreateAccount  true  "Request body"
// @Success      201  {object}  response.SuccessWithData
// @Failure      400  {object}  response.ErrorDetails
// @Failure      409  {object}  response.ErrorDetails
// @Router       /daftar [post]
func (accountController *AccountController) Register(c *fiber.Ctx) error {
	req := new(model.CreateAccount)
	if err := c.BodyParser(req); err != nil {
		return response.ErrorCustom(c, fiber.StatusBadRequest, err.Error(), "Invalid request body")
	}

	account, err := accountController.AccountService.CreateAccount(c.Context(), req)
	if err != nil {
		return response.Error(c, err, nil)
	}

	return c.Status(fiber.StatusCreated).JSON(response.SuccessWithData{
		Code:    fiber.StatusCreated,
		Status:  "success",
		Message: "Account registration successful",
		Data:    account,
	})
}

// @Tags         Accounts
// @Summary      Deposit to an account (Tabung)
// @Description  API for depositing money into an account.
// @Accept       json
// @Produce      json
// @Param        request  body  model.DepositRequest  true  "Request body"
// @Success      200  {object}  response.SuccessWithData
// @Failure      400  {object}  response.ErrorDetails
// @Failure      404  {object}  response.ErrorDetails
// @Router       /tabung [post]
func (accountController *AccountController) Deposit(c *fiber.Ctx) error {
	req := new(model.DepositRequest)
	if err := c.BodyParser(req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid request body")
	}

	err := accountController.AccountService.Deposit(c.Context(), req)
	if err != nil {
		return response.Error(c, err, nil)
	}

	account, err := accountController.AccountService.GetBalance(c.Context(), req.AccountNumber)
	if err != nil {
		return response.Error(c, err, nil)
	}

	return c.Status(fiber.StatusOK).JSON(response.SuccessWithData{
		Code:    fiber.StatusOK,
		Status:  "success",
		Message: "Deposit successful",
		Data:    map[string]interface{}{"saldo": account.Balance},
	})
}

// @Tags         Accounts
// @Summary      Withdraw from an account (Tarik)
// @Description  API for withdrawing money from an account.
// @Accept       json
// @Produce      json
// @Param        request  body  model.Withdrawal  true  "Request body"
// @Success      200  {object}  response.SuccessWithData
// @Failure      400  {object}  response.ErrorDetails
// @Failure      404  {object}  response.ErrorDetails
// @Router       /tarik [post]
func (accountController *AccountController) Withdrawal(c *fiber.Ctx) error {
	req := new(model.Withdrawal)
	if err := c.BodyParser(req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid request body")
	}

	err := accountController.AccountService.Withdraw(c.Context(), req)
	if err != nil {
		return response.Error(c, err, nil)
	}

	account, err := accountController.AccountService.GetBalance(c.Context(), req.AccountNumber)
	if err != nil {
		return response.Error(c, err, nil)
	}

	return c.Status(fiber.StatusOK).JSON(response.SuccessWithData{
		Code:    fiber.StatusOK,
		Status:  "success",
		Message: "Withdrawal successful",
		Data:    map[string]interface{}{"saldo": account.Balance},
	})
}

// @Tags         Accounts
// @Summary      Get account balance (Saldo)
// @Description  API for checking the balance of an account.
// @Produce      json
// @Param        accountNumber  path  string  true  "Account number"
// @Success      200  {object}  response.SuccessWithData
// @Failure      400  {object}  response.ErrorDetails
// @Failure      404  {object}  response.ErrorDetails
// @Router       /saldo/{accountNumber} [get]
func (accountController *AccountController) GetBalance(c *fiber.Ctx) error {
	accountNumber := c.Params("accountNumber")

	// Convert accountNumber to uint
	if _, err := strconv.ParseUint(accountNumber, 10, 64); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid account number")
	}

	account, err := accountController.AccountService.GetBalance(c.Context(), accountNumber)
	if err != nil {
		return response.Error(c, err, nil)
	}

	return c.Status(fiber.StatusOK).JSON(response.SuccessWithData{
		Code:    fiber.StatusOK,
		Status:  "success",
		Message: "Get balance successful",
		Data:    map[string]interface{}{"saldo": account.Balance},
	})
}
