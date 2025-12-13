package middleware

import (
	"strings"

	"uas-backend/app/model"
	"uas-backend/app/repository"
	"uas-backend/config"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
)

func JWTAuth(userRepo repository.UserRepository) fiber.Handler {
	return func(c *fiber.Ctx) error {

		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return fiber.ErrUnauthorized
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")

		claims := &model.JWTClaims{}
		token, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (interface{}, error) {
			return []byte(config.JWTSecret()), nil
		})

		if err != nil || !token.Valid {
			return fiber.ErrUnauthorized
		}

		// 🔥 ambil permission dari DB via repository
		perms, err := userRepo.GetUserPermissions(claims.UserID)
		if err != nil {
			return fiber.ErrForbidden
		}

		c.Locals("user", claims)
		c.Locals("user_id", claims.UserID)
		c.Locals("role", claims.Role)
		c.Locals("permissions", perms)

		return c.Next()
	}
}
