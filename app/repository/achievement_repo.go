package repository

import (
	"context"
	"errors"

	"uas-backend/app/model"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type AchievementRepository interface {
	Create(ctx context.Context, a *model.Achievement) (primitive.ObjectID, error)
}

type achievementRepository struct {
	collection *mongo.Collection
}

func NewAchievementRepository(db *mongo.Database) AchievementRepository {
	return &achievementRepository{
		collection: db.Collection("achievements"),
	}
}

func (r *achievementRepository) Create(ctx context.Context, a *model.Achievement) (primitive.ObjectID, error) {
	res, err := r.collection.InsertOne(ctx, a)
	if err != nil {
		return primitive.NilObjectID, err
	}

	oid, ok := res.InsertedID.(primitive.ObjectID)
	if !ok {
		return primitive.NilObjectID, errors.New("failed to cast inserted ID to ObjectID")
	}

	a.ID = oid
	return oid, nil
}
