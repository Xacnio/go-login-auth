package middleware

import (
	"github.com/gofiber/fiber/v2"
	"go-login-auth/utils"
)

func UserAuthorization(c *fiber.Ctx) error {
	metadata, err := utils.ExtractTokenMetadata(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"message": "unauthorized"})
	}
	userid, err := utils.FetchAuthData(metadata)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"message": err.Error()})
	}
	c.Locals("User-ID", userid)

	return c.Next()
}
