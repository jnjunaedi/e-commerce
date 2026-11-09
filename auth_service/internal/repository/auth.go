package repository

import (
	"context"
	"database/sql"
	"e-commerce/auth_service/internal/domain"
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

func (r *RepositoryImpl) Create(ctx context.Context, auth domain.Auth) error {
	query := `INSERT INTO auth_credentials (auth_id, email, username, password_hash) VALUES (?, ?, ?, ?)`

	_, err := r.db.ExecContext(ctx, query, auth.ID, auth.Email, auth.Username, auth.PasswordHash)
	if err != nil {
		var mysqlErr *mysql.MySQLError
		if errors.As(err, &mysqlErr) {
			if mysqlErr.Number == 1062 {
				return apperror.New(apperror.ErrBadRequest, "data creadentials sudah pernah dibuat")
			}
		}
		return apperror.New(apperror.ErrInternalServer, "gagal menyimpan data credentials, karna gangguan sistem")
	}

	return nil
}

func (r *RepositoryImpl) FindByEmail(ctx context.Context, email string) (domain.Auth, error) {
	query := `SELECT auth_id, email, username, password_hash, created_at, updated_at FROM auth_credentials WHERE email = ?`

	var auth domain.Auth

	err := r.db.QueryRowContext(ctx, query, email).Scan(&auth.ID, &auth.Email, &auth.Username, &auth.PasswordHash, &auth.CreatedAt, &auth.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return domain.Auth{}, apperror.New(apperror.ErrNotFound, "email atau password yang anda masukkan salah")
		}
		return domain.Auth{}, apperror.New(apperror.ErrInternalServer, "gagal mengambil data creadentials, karena gangguan sistem")
	}

	return auth, nil
}

func (r *RepositoryImpl) ExistsByEmailOrUsername(ctx context.Context, email, username string) (bool, error) {
	query := `SELECT 1 FROM auth_credentials WHERE email = ? OR username = ? LIMIT 1`

	var dummyInt int
	err := r.db.QueryRowContext(ctx, query, email, username).Scan(&dummyInt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}
		return false, apperror.New(apperror.ErrInternalServer, "gagal mengecek ketersediaan akun karena gangguan sistem")
	}

	return true, nil
}
