package database

import (
	"context"
	"log"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"uas-backend/config"
)

var MongoClient *mongo.Client
var MongoDB *mongo.Database

func ConnectMongo() {
	uri := config.MongoURI()
	dbName := config.MongoDB()

	clientOpts := options.Client().ApplyURI(uri)

	client, err := mongo.Connect(context.Background(), clientOpts)
	if err != nil {
		log.Fatalf("❌ Failed to connect MongoDB: %v", err)
	}

	MongoClient = client
	MongoDB = client.Database(dbName)

	log.Println("✅ Connected to MongoDB")
}
