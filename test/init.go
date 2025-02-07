package test

import (
	"account-service/src/database"
	"account-service/src/router"
	"account-service/src/utils"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

var App = fiber.New(fiber.Config{
	CaseSensitive: true,
	ErrorHandler:  utils.ErrorHandler,
})
var DB *gorm.DB
var Log = utils.Log

func init() {
	// TODO: You can modify host and database configuration for tests
	DB = database.Connect()
	router.Routes(App, DB)
	App.Use(utils.NotFoundHandler)
}
