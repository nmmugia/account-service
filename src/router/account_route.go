package router

import (
	"account-service/src/controller"
	"account-service/src/service"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
)

func AccountRoutes(v1 fiber.Router, a service.AccountServices, v *validator.Validate) {
	accountController := controller.NewAccountController(a, v)

	v1.Post("/tabung", accountController.Deposit)
	v1.Post("/tarik", accountController.Withdrawal)
	v1.Post("/daftar", accountController.Register)
	v1.Get("/saldo/:accountNumber", accountController.GetBalance)
}
