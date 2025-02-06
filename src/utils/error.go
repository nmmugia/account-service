package utils

import (
	"account-service/src/response"
	"errors"

	"github.com/gofiber/fiber/v2"
)

func ErrorHandler(c *fiber.Ctx, err error) error {
	if errorsMap := CustomErrorMessages(err); len(errorsMap) > 0 {
		return response.ErrorCustom(c, fiber.StatusBadRequest, "Bad Request", errorsMap)
	}

	var fiberErr *fiber.Error
	if errors.As(err, &fiberErr) {
		return response.ErrorCustom(c, fiberErr.Code, fiberErr.Message, nil)
	}

	return response.ErrorCustom(c, fiber.StatusInternalServerError, "Internal Server Error", nil)
}

func NotFoundHandler(c *fiber.Ctx) error {
	return response.ErrorCustom(c, fiber.StatusNotFound, "Endpoint Not Found", nil)
}
