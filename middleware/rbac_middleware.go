package middleware

import (
	"github.com/gofiber/fiber/v2"
)

func RequirePermission(permission string) fiber.Handler {
	return func(c *fiber.Ctx) error {

		permissions := c.Locals("permissions")

		if permissions == nil {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"code":    403,
				"message": "Forbidden",
				"error":   "No permissions found",
			})
		}

		permList := permissions.([]string)

		// cek permission
		for _, p := range permList {
			if p == permission {
				return c.Next()
			}
		}

		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"code":    403,
			"message": "Forbidden",
			"error":   "Insufficient permissions",
		})
	}
}
