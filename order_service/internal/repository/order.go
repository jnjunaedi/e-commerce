package repository

import (
	"context"
	"database/sql"
	"e-commerce/order_service/internal/domain"
	"e-commerce/pkg/utils"
	"e-commerce/pkg/utils/apperror"
	"errors"

	"github.com/go-sql-driver/mysql"
)

type RepositoryImpl struct {
	db          *sql.DB
	idGenerator utils.IDGenerator
}

func NewRepository(db *sql.DB, idGenerator utils.IDGenerator) domain.Repository {
	return &RepositoryImpl{db: db, idGenerator: idGenerator}
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

func (r *RepositoryImpl) Create(ctx context.Context, order domain.Order) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	querySaveOrder := `INSERT INTO orders (order_id, user_id, total_price, order_status, idempotency_key) VALUES (?, ?, ?, ?, ?)`
	_, err = tx.ExecContext(ctx, querySaveOrder, order.ID, order.UserID, order.TotalPrice, order.OrderStatus, order.IdempotencyKey)
	if err != nil {
		var mysqlErr *mysql.MySQLError
		if errors.As(err, &mysqlErr) && mysqlErr.Number == 1062 {
			return apperror.New(apperror.ErrBadRequest, "pesanan sedang diproses, mohon jangan klik dua kali!")
		}
		return apperror.New(apperror.ErrInternalServer, "gagal menyimpan riwayat pesanan, karna gangguan sistem")
	}

	querySaveOrderItem := `INSERT INTO order_items (item_id, order_id, product_id, flash_sale_id, quantity, unit_price, subtotal) VALUES (?, ?, ?, ?, ?, ?, ?)`
	stmtOdrItem, err := tx.PrepareContext(ctx, querySaveOrderItem)
	if err != nil {
		return err
	}
	defer stmtOdrItem.Close()

	for _, item := range order.Items {
		genUUID, err := r.idGenerator.GenerateUUID()
		if err != nil {
			return err
		}

		i := item.FlashSaleID
		var fsID *string = nil
		if *i != "" {
			fsID = i
		} else {
			fsID = nil
		}

		_, err = stmtOdrItem.ExecContext(ctx, genUUID, order.ID, item.ProductID, fsID, item.Quantity, item.UnitPrice, item.Subtotal)
		if err != nil {
			var mysqlErr *mysql.MySQLError
			if errors.As(err, &mysqlErr) {
				if mysqlErr.Number == 1062 {
					return apperror.New(apperror.ErrBadRequest, "barang pesanan tersebut sudah pernah dibuat")
				}
			}
			return apperror.New(apperror.ErrInternalServer, "gagal menyimpan riwayat barang pesanan, karna gangguan sistem")
		}
	}

	if err := tx.Commit(); err != nil {
		return err
	}
	return nil
}

func (r *RepositoryImpl) GetByID(ctx context.Context, orderID string) (domain.Order, error) {
	query := `SELECT order_id, user_id, total_price, order_status, idempotency_key, created_at FROM orders WHERE order_id = ?`

	var order domain.Order

	err := r.db.QueryRowContext(ctx, query, orderID).Scan(&order.ID, &order.UserID, &order.TotalPrice, &order.OrderStatus, &order.IdempotencyKey, &order.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return domain.Order{}, apperror.New(apperror.ErrNotFound, "riwayat order tidak ditemukan")
		}
		return domain.Order{}, apperror.New(apperror.ErrInternalServer, "gagal mengambil info riwayat pesanan, karna gangguan sistem")
	}

	return order, nil
}

func (r *RepositoryImpl) CheckByIdempotencyKey(ctx context.Context, key string) (bool, error) {
	query := `SELECT 1 FROM orders WHERE idempotency_key = ? LIMIT 1`

	var dummyInt int
	err := r.db.QueryRowContext(ctx, query, key).Scan(&dummyInt)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, err
	}

	return true, nil
}

func (r *RepositoryImpl) UpdateStatus(ctx context.Context, orderID string, status string) error {
	query := `UPDATE orders SET order_status = ? WHERE order_id = ?`

	result, err := r.db.ExecContext(ctx, query, status, orderID)
	if err != nil {
		return apperror.New(apperror.ErrInternalServer, "gagal memperbarui status pesanan, karna gangguan sistem")
	}

	rwAffectd, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rwAffectd <= 0 {
		return apperror.New(apperror.ErrBadRequest, "gagal memperbarui status pesanan, order ID tidak valid atau status sama")
	}

	return nil
}

func (r *RepositoryImpl) DeductLocalFlashSaleStock(ctx context.Context, flashSaleID string, qty int32) (int64, error) {
	query := `
		UPDATE flash_sale_caches 
		SET stock = stock - ?, updated_at = NOW() 
		WHERE flash_sale_id = ? AND stock >= ?`

	result, err := r.db.ExecContext(ctx, query, qty, flashSaleID, qty)
	if err != nil {
		return 0, err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return 0, err
	}

	return rowsAffected, nil
}
