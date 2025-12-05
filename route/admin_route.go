package route

import (
	"context"
	"uas-backend/app/model"
	"uas-backend/app/repository"
	"uas-backend/app/service"
	"uas-backend/database"
	"uas-backend/middleware"

	"github.com/gofiber/fiber/v2"
)

func AdminRoutes(r fiber.Router) {

	userRepo := repository.NewUserRepository(database.PG)
	studentRepo := repository.NewStudentRepository(database.PG)
	lecturerRepo := repository.NewLecturerRepository(database.PG)

	userService := service.NewUserService(userRepo, studentRepo, lecturerRepo)

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

	// 4. GET ALL USERS (GET /users/)
	admin.Get("/users", func(c *fiber.Ctx) error {
		users, err := userService.GetAllUsers(context.Background())
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"message": "failed to get users"})
		}
		return c.JSON(fiber.Map{"data": users})
	})

	// 5. GET USERS BY ID (GET /users/:id)
	admin.Get("/users/:id", func(c *fiber.Ctx) error {
		id := c.Params("id")

		user, err := userService.GetUserByID(context.Background(), id)
		if err != nil {
			return c.Status(404).JSON(fiber.Map{"message": "user not found"})
		}

		return c.JSON(fiber.Map{"data": user})
	})

}
