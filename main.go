package main

import (
	"log"

	"uas-backend/config"
	"uas-backend/database"
	"uas-backend/route"
)

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

	// Setup Routes
	route.SetupRoutes(app)

	// Run Server
	port := config.AppPort()
	log.Printf("Server running at http://localhost:%s", port)
	log.Fatal(app.Listen(":" + port))
}
