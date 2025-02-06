package response

import (
	"github.com/gofiber/fiber/v2"
	"github.com/sirupsen/logrus"
)

func Error(c *fiber.Ctx, err *fiber.Error, details interface{}) error {
	var errRes error
	if details != nil {
		errRes = c.Status(err.Code).JSON(ErrorDetails{
			Code:    err.Code,
			Status:  "error",
			Message: err.Message,
			Errors:  details,
		})
	} else {
		errRes = c.Status(err.Code).JSON(Common{
			Code:    err.Code,
			Status:  "error",
			Message: err.Message,
		})
	}

	if errRes != nil {
		logrus.Errorf("Failed to send error response : %+v", errRes)
	}

	return errRes
}

func ErrorCustom(c *fiber.Ctx, statusCode int, message string, details interface{}) error {
	var errRes error
	if details != nil {
		errRes = c.Status(statusCode).JSON(ErrorDetails{
			Code:    statusCode,
			Status:  "error",
			Message: message,
			Errors:  details,
		})
	} else {
		errRes = c.Status(statusCode).JSON(Common{
			Code:    statusCode,
			Status:  "error",
			Message: message,
		})
	}

	if errRes != nil {
		logrus.Errorf("Failed to send error response : %+v", errRes)
	}

	return errRes
}
