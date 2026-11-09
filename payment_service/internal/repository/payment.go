package repository

import (
	"context"
	"database/sql"
	"e-commerce/payment_service/internal/domain"
	"e-commerce/pkg/utils/apperror"
	"errors"
	"time"

	"github.com/go-sql-driver/mysql"
)

type RepositoryImpl struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) domain.Repository {
	return &RepositoryImpl{db: db}
}

func (r *RepositoryImpl) getExecutor(ctx context.Context) interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
} {
	if tx, ok := ctx.Value(txKey{}).(*sql.Tx); ok {
		return tx
	}
	return r.db
}

func (r *RepositoryImpl) CreateOutbox(ctx context.Context, id, eventType string, payload []byte) error {
	query := `INSERT INTO outbox (id, event_type, payload, status) VALUES (?, ?, ?, 'PENDING')`

	_, err := r.getExecutor(ctx).ExecContext(ctx, query, id, eventType, string(payload))
	if err != nil {
		return apperror.New(apperror.ErrInternalServer, "gagal menyimpan data outbox, karena gangguan sistem")
	}

	return nil
}

func (r *RepositoryImpl) Create(ctx context.Context, p domain.Payment) error {
	query := `INSERT INTO payments (payment_id, order_id, payment_method, payment_amount, payment_status, payment_time) VALUES (?, ?, ?, ?, ?, ?)`

	var paymentTimeInterface any = nil
	if p.PaidAt > 0 {
		paymentTimeInterface = time.Unix(p.PaidAt, 0)
	}

	_, err := r.db.ExecContext(ctx, query, p.ID, p.OrderID, p.PaymentMethod, p.PaymentAmount, p.PaymentStatus, paymentTimeInterface)
	if err != nil {
		var mysqlErr *mysql.MySQLError
		if errors.As(err, &mysqlErr) {
			if mysqlErr.Number == 1062 {
				return apperror.New(apperror.ErrBadRequest, "riwayat pembayaran tersebut sudah pernah dibuat")
			}
		}
		return apperror.New(apperror.ErrInternalServer, "gagal menyimpan riwayat pembayaran, dikarenakan sedang ada gangguan dari sistem")
	}

	return nil
}

func (r *RepositoryImpl) GetByOrderID(ctx context.Context, orderID string) (domain.Payment, error) {
	query := `SELECT payment_id, order_id, payment_method, payment_amount, payment_status, payment_time, created_at, updated_at FROM payments WHERE order_id = ?`

	var payment domain.Payment
	var paymentTime time.Time

	err := r.db.QueryRowContext(ctx, query, orderID).Scan(&payment.ID, &payment.OrderID, &payment.PaymentMethod, &payment.PaymentAmount, &payment.PaymentStatus, &paymentTime, &payment.CreatedAt, &payment.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return domain.Payment{}, apperror.New(apperror.ErrNotFound, "Detail pembayaran tidak ditemukan")
		}
		return domain.Payment{}, apperror.New(apperror.ErrInternalServer, "gagal mengambil info riwayat pembayaran, dikarenakan sedang ada gangguan dari sistem")
	}

	payment.PaidAt = paymentTime.Unix()

	return payment, nil

}

func (r *RepositoryImpl) UpdateStatus(ctx context.Context, orderID, status string, paidAt int64) error {
	query := `UPDATE payments SET payment_status = ?, payment_time = COALESCE(?, payment_time) WHERE order_id = ?`

	var paymentTimeInterface any = nil
	if paidAt > 0 {
		paymentTimeInterface = time.Unix(paidAt, 0)
	}

	result, err := r.db.ExecContext(ctx, query, status, paymentTimeInterface, orderID)
	if err != nil {
		return apperror.New(apperror.ErrInternalServer, "gagal memperbarui status pembayaran, dikarenakan sedang ada gangguan dari sistem")
	}

	rwAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rwAffected <= 0 {
		return apperror.New(apperror.ErrBadRequest, "gagal memperbarui status pembayaran, order ID tidak valid atau status sama")
	}

	return nil
}
