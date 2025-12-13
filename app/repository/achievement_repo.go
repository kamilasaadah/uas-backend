package repository

import (
	"context"

	"uas-backend/app/model"

	"go.mongodb.org/mongo-driver/mongo"
)

type AchievementRepository interface {
	Create(ctx context.Context, a *model.Achievement) error
}

type achievementRepository struct {
	collection *mongo.Collection
}

func NewAchievementRepository(db *mongo.Database) AchievementRepository {
	return &achievementRepository{
		collection: db.Collection("achievements"),
	}
}

func (r *achievementRepository) Create(ctx context.Context, a *model.Achievement) error {
	_, err := r.collection.InsertOne(ctx, a)
	return err
}
