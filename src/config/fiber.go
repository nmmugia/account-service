package config

import (
	"account-service/src/utils"

	"github.com/bytedance/sonic"
	"github.com/gofiber/fiber/v2"
)

func FiberConfig() fiber.Config {
	return fiber.Config{
		Prefork:       IsProd,
		CaseSensitive: true,
		ServerHeader:  "Fiber",
		AppName:       AppName,
		ErrorHandler:  utils.ErrorHandler,
		JSONEncoder:   sonic.Marshal,
		JSONDecoder:   sonic.Unmarshal,
	}
}
