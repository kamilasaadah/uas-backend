package middleware

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
)

func RequirePermission(permission string) fiber.Handler {
	return func(c *fiber.Ctx) error {

		raw := c.Locals("permissions")

		fmt.Println("=== DEBUG PERMISSION ===")
		fmt.Println("REQUIRED :", permission)
		fmt.Printf("RAW      : %#v\n", raw)

		if raw == nil {
			fmt.Println("❌ permissions NIL")
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

		fmt.Println("PERM LIST:", permList)

		// cek permission
		for _, p := range permList {
			if p == permission {
				fmt.Println("✅ PERMISSION MATCH")
				return c.Next()
			}
		}

		fmt.Println("❌ NO MATCH")

		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"code":    403,
			"message": "Forbidden",
			"error":   "Insufficient permissions",
		})
	}
}

func RequireAnyPermission(perms ...string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		raw := c.Locals("permissions")
		if raw == nil {
			return fiber.ErrForbidden
		}

		var userPerms []string
		switch p := raw.(type) {
		case []string:
			userPerms = p
		case []interface{}:
			for _, x := range p {
				if s, ok := x.(string); ok {
					userPerms = append(userPerms, s)
				}
			}
		default:
			return fiber.ErrForbidden
		}

		for _, up := range userPerms {
			for _, rp := range perms {
				if up == rp {
					return c.Next()
				}
			}
		}

		return fiber.ErrForbidden
	}
}
