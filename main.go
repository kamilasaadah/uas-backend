package main

import (
	"log"

	"uas-backend/config"
	"uas-backend/database"
	"uas-backend/route"

	// 🔽 Swagger
	_ "uas-backend/docs"

	fiberSwagger "github.com/swaggo/fiber-swagger"
)

/*
   ===== Swagger Global Annotation =====
*/

// @title UAS Backend API
// @version 1.0
// @description API Backend untuk UAS
// @host localhost:3000
// @BasePath /api
// @schemes http

func main() {
	// Load environment
	config.LoadEnv()

	// Initialize logger
	config.InitLogger()

	// Connect Databases
	database.ConnectPostgres()
	database.ConnectMongo()

	// Setup Fiber App
	app := config.NewFiber()

	// 🔽 Swagger route
	app.Get("/swagger/*", fiberSwagger.WrapHandler)

	// Setup Routes
	route.SetupRoutes(app)

	// Run Server
	port := config.AppPort()
	log.Printf("Server running at http://localhost:%s", port)
	log.Fatal(app.Listen(":" + port))
}
