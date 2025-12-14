package repository

import (
	"context"
	"errors"

	"uas-backend/app/model"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type AchievementRepository interface {
	Create(ctx context.Context, a *model.Achievement) (primitive.ObjectID, error)
	GetByID(ctx context.Context, id primitive.ObjectID) (*model.Achievement, error)
	AddAttachment(ctx context.Context, id primitive.ObjectID, att model.Attachment) error
	Update(ctx context.Context, a *model.Achievement) error
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

func (r *achievementRepository) GetByID(
	ctx context.Context,
	id primitive.ObjectID,
) (*model.Achievement, error) {

	var achievement model.Achievement
	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&achievement)
	if err != nil {
		return nil, err
	}
	return &achievement, nil
}

func (r *achievementRepository) AddAttachment(
	ctx context.Context,
	id primitive.ObjectID,
	att model.Attachment,
) error {

	_, err := r.collection.UpdateOne(
		ctx,
		bson.M{"_id": id},
		bson.M{
			"$push": bson.M{"attachments": att},
			"$set":  bson.M{"updatedAt": att.UploadedAt},
		},
	)
	return err
}

func (r *achievementRepository) Update(ctx context.Context, a *model.Achievement) error {
	_, err := r.collection.UpdateOne(
		ctx,
		bson.M{"_id": a.ID},
		bson.M{
			"$set": a,
		},
	)
	return err
}
