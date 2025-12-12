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

		var permList []string

		// CASE 1: []string (JWTClaims)
		if list, ok := raw.([]string); ok {
			permList = list
		} else

		// CASE 2: []interface{} (kalau dari MapClaims)
		if rawList, ok := raw.([]interface{}); ok {
			for _, v := range rawList {
				if str, ok := v.(string); ok {
					permList = append(permList, str)
				}
			}
		} else {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"code":    403,
				"message": "Forbidden",
				"error":   "Invalid permission format",
			})
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
