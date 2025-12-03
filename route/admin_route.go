package route

import (
	"context"
	"uas-backend/app/model"
	"uas-backend/app/repository"
	"uas-backend/app/service"
	"uas-backend/middleware"

	"github.com/gofiber/fiber/v2"
)

func AdminRoutes(r fiber.Router) {

	userRepo := repository.NewUserRepository()
	userService := service.NewUserService(userRepo)

	admin := r.Group("/users",
		middleware.JWTAuth(),                        // cek token
		middleware.RequirePermission("user:manage"), // cek permission
	)

	// CREATE USER
	admin.Post("/", func(c *fiber.Ctx) error {

		var req model.CreateUserRequest
		if err := c.BodyParser(&req); err != nil {
			return c.Status(400).JSON(fiber.Map{
				"message": "invalid request data",
			})
		}

		newUser, err := userService.CreateUser(context.Background(), req)
		if err != nil {
			return c.Status(400).JSON(fiber.Map{
				"message": err.Error(),
			})
		}

		return c.JSON(fiber.Map{
			"message": "user created successfully",
			"data":    newUser,
		})
	})

	// 2. UPDATE USER (PUT /users/:id)
	admin.Put("/:id", func(c *fiber.Ctx) error {

		id := c.Params("id")
		var req model.UpdateUserRequest

		if err := c.BodyParser(&req); err != nil {
			return c.Status(400).JSON(fiber.Map{
				"message": "invalid request data",
			})
		}

		if err := userService.UpdateUser(context.Background(), id, req); err != nil {
			return c.Status(400).JSON(fiber.Map{
				"message": err.Error(),
			})
		}

		return c.JSON(fiber.Map{
			"message": "user updated successfully",
		})
	})

	// 3. ASSIGN ROLE (PUT /users/:id/role)
	admin.Put("/:id/role", func(c *fiber.Ctx) error {

		id := c.Params("id")
		var req model.UpdateUserRoleRequest

		if err := c.BodyParser(&req); err != nil {
			return c.Status(400).JSON(fiber.Map{
				"message": "invalid request data",
			})
		}

		if err := userService.AssignRole(context.Background(), id, req.RoleID); err != nil {
			return c.Status(400).JSON(fiber.Map{
				"message": err.Error(),
			})
		}

		return c.JSON(fiber.Map{
			"message": "role updated successfully",
		})
	})
}
