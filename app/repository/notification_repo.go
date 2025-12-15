package repository

import (
	"context"
	"uas-backend/app/model"

	"github.com/jackc/pgx/v5/pgxpool"
)

type NotificationRepository interface {
	Create(ctx context.Context, n *model.Notification) error
}

type notificationRepository struct {
	db *pgxpool.Pool
}

func NewNotificationRepository(db *pgxpool.Pool) NotificationRepository {
	return &notificationRepository{db: db}
}

func (r *notificationRepository) Create(
	ctx context.Context,
	n *model.Notification,
) error {

	query := `
		INSERT INTO notifications (
			id, user_id, title, message, is_read, created_at
		) VALUES (
			gen_random_uuid(), $1, $2, $3, false, NOW()
		)
	`

	_, err := r.db.Exec(
		ctx,
		query,
		n.UserID,
		n.Title,
		n.Message,
	)

	return err
}
