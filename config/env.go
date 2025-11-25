package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

func LoadEnv() {
	err := godotenv.Load()
	if err != nil {
		log.Println("Warning: .env file not found, using system environment variables")
	}
}

func AppPort() string {
	return os.Getenv("APP_PORT")
}

func DBDsn() string {
	return os.Getenv("DB_DSN")
}

func MongoURI() string {
	return os.Getenv("MONGO_URI")
}

func MongoDB() string {
	return os.Getenv("MONGO_DBNAME")
}

func JWTSecret() string {
	return os.Getenv("JWT_SECRET")
}
