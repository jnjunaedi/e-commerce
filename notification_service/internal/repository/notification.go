package repository

import (
	"context"
	"database/sql"
	"e-commerce/notification_service/internal/domain"
	"e-commerce/pkg/utils/apperror"
	"errors"

	"github.com/go-sql-driver/mysql"
)

type RepositoryImpl struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) domain.Repository {
	return &RepositoryImpl{db: db}
}

func (r *RepositoryImpl) Create(ctx context.Context, n domain.Notification) error {
	query := `INSERT INTO notifications (notification_id, user_id, order_id, notification_type, title, message, status) VALUES (?, ?, ?, ?, ?, ?, ?)`
	_, err := r.db.ExecContext(ctx, query, n.ID, n.UserID, n.OrderID, n.NotificationType, n.Title, n.Message, n.Status)
	if err != nil {
		var mysqlErr *mysql.MySQLError
		if errors.As(err, &mysqlErr) {
			if mysqlErr.Number == 1062 {
				return apperror.New(apperror.ErrBadRequest, "data notification sudah pernah dibuat")
			}
		}
		return apperror.New(apperror.ErrInternalServer, "gagal menyimpan data notification, karna gangguan sistem")
	}
	return nil
}
func (r RepositoryImpl) UpdateStatus(ctx context.Context, id string, status string) error {
	query := `UPDATE notifications SET status = ? WHERE notification_id = ?`
	_, err := r.db.ExecContext(ctx, query, status, id)
	if err != nil {
		return apperror.New(apperror.ErrInternalServer, "gagal memperbarui status notification, karna gangguan sistem")
	}
	return nil
}
