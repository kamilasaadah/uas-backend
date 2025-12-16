package repository

import (
	"context"
	"errors"
	"time"

	"uas-backend/app/model"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type AchievementRepository interface {
	Create(ctx context.Context, a *model.Achievement) (primitive.ObjectID, error)
	GetByID(ctx context.Context, id primitive.ObjectID) (*model.Achievement, error)
	AddAttachment(ctx context.Context, id primitive.ObjectID, att model.Attachment) error
	Update(ctx context.Context, a *model.Achievement) error
	SoftDelete(ctx context.Context, id primitive.ObjectID) error
	FindByStudentIDs(ctx context.Context, studentIDs []string) ([]model.Achievement, error)
	FindAll(ctx context.Context) ([]model.Achievement, error)
	FindByIDs(ctx context.Context, ids []primitive.ObjectID) ([]model.Achievement, error)
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

func (r *achievementRepository) SoftDelete(ctx context.Context, id primitive.ObjectID) error {
	_, err := r.collection.UpdateOne(
		ctx,
		bson.M{"_id": id},
		bson.M{
			"$set": bson.M{
				"is_deleted": true,
				"status":     "deleted",
				"updatedAt":  time.Now(),
			},
		},
	)
	return err
}

// helper filter: ambil data yang BELUM di-delete
func notDeletedFilter() bson.M {
	return bson.M{
		"$or": []bson.M{
			{"is_deleted": false},
			{"is_deleted": bson.M{"$exists": false}},
		},
	}
}

func (r *achievementRepository) FindByStudentIDs(
	ctx context.Context,
	studentIDs []string,
) ([]model.Achievement, error) {

	// ⛔ jangan query Mongo dengan $in: []
	if len(studentIDs) == 0 {
		return []model.Achievement{}, nil
	}

	filter := bson.M{
		"studentId": bson.M{"$in": studentIDs},
	}

	// gabung dengan not-deleted
	for k, v := range notDeletedFilter() {
		filter[k] = v
	}

	cursor, err := r.collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}

	var result []model.Achievement
	if err := cursor.All(ctx, &result); err != nil {
		return nil, err
	}

	return result, nil
}

func (r *achievementRepository) FindAll(
	ctx context.Context,
) ([]model.Achievement, error) {

	opts := options.Find().
		SetSort(bson.D{{Key: "createdAt", Value: -1}})

	cursor, err := r.collection.Find(
		ctx,
		notDeletedFilter(), // ✅ FIX UTAMA
		opts,
	)
	if err != nil {
		return nil, err
	}

	var result []model.Achievement
	if err := cursor.All(ctx, &result); err != nil {
		return nil, err
	}

	return result, nil
}

func (r *achievementRepository) FindByIDs(
	ctx context.Context,
	ids []primitive.ObjectID,
) ([]model.Achievement, error) {

	// ⛔ jangan query Mongo dengan $in: []
	if len(ids) == 0 {
		return []model.Achievement{}, nil
	}

	filter := bson.M{
		"_id": bson.M{"$in": ids},
	}

	// gabung dengan not-deleted
	for k, v := range notDeletedFilter() {
		filter[k] = v
	}

	cursor, err := r.collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}

	var result []model.Achievement
	if err := cursor.All(ctx, &result); err != nil {
		return nil, err
	}

	return result, nil
}
