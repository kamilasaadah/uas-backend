package middleware

import (
	"github.com/gofiber/fiber/v2"
)

func RequirePermission(permission string) fiber.Handler {
	return func(c *fiber.Ctx) error {

		raw := c.Locals("permissions")

		if raw == nil {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"code":    403,
				"message": "Forbidden",
				"error":   "No permissions found",
			})
		}

		rawList, ok := raw.([]interface{})
		if !ok {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"code":    403,
				"message": "Forbidden",
				"error":   "Invalid permission format",
			})
		}

		permList := make([]string, 0)
		for _, v := range rawList {
			if str, ok := v.(string); ok {
				permList = append(permList, str)
			}
		}

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
