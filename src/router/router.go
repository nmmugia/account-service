package router

import (
	"account-service/src/config"
	"account-service/src/service"
	"account-service/src/utils"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

func Routes(app *fiber.App, db *gorm.DB) {
	validate := utils.Validator()

	healthCheckService := service.NewHealthCheckService(db)
	accountService := service.NewAccountService(db, validate)

	v1 := app.Group("/v1")

	HealthCheckRoutes(v1, healthCheckService)
	AccountRoutes(v1, accountService, validate)
	// add another routes here...

	if !config.IsProd {
		DocsRoutes(v1)
	}
}
