package middleware

import (
	"strings"
	"uas-backend/app/model"
	"uas-backend/config"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
)

func JWTAuth() fiber.Handler {
	return func(c *fiber.Ctx) error {

		// Ambil token dari header
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"code":    401,
				"message": "Unauthorized",
				"error":   "Missing Authorization header",
			})
		}

		// Format harus "Bearer token"
		tokenString := strings.Replace(authHeader, "Bearer ", "", 1)
		if tokenString == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"code":    401,
				"message": "Unauthorized",
				"error":   "Invalid token format",
			})
		}

		// Parse token
		claims := &model.JWTClaims{}
		token, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (interface{}, error) {
			return []byte(config.JWTSecret()), nil
		})

		if err != nil || !token.Valid {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"code":    401,
				"message": "Unauthorized",
				"error":   "Invalid or expired token",
			})
		}

		// SIMPAN SEMUA CLAIMS DI SINI ⬇
		c.Locals("user", claims)

		// Simpan user_id, role_id ke context
		c.Locals("user_id", claims.UserID)
		c.Locals("role", claims.Role)
		c.Locals("permissions", claims.Permissions)

		return c.Next()
	}
}
