package repository

import (
	"context"
	"database/sql"
	"e-commerce/pkg/utils/apperror"
	"e-commerce/user_service/internal/domain"
	"errors"

	"github.com/go-sql-driver/mysql"
)

type RepositoryImpl struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) domain.Repository {
	return &RepositoryImpl{db: db}
}

func (r *RepositoryImpl) CreateProfile(ctx context.Context, profile domain.Profile) error {
	query := `INSERT INTO user_profiles (usr_profile_id, user_id, username, full_name, phone_number, avatar_url) VALUES (?, ?, ?, ?, ?, ?)`
	_, err := r.db.ExecContext(ctx, query, profile.ID, profile.UserID, profile.Username, profile.FullName, profile.PhoneNumber, profile.AvatarURL)
	if err != nil {
		var mysqlErr *mysql.MySQLError
		if errors.As(err, &mysqlErr) {
			if mysqlErr.Number == 1062 {
				return apperror.New(apperror.ErrBadRequest, "profile pengguna tersebut sudah pernah dibuat")
			}
		}
		return apperror.New(apperror.ErrInternalServer, "gagal menyimpan data profile, karna gangguan sistem")
	}

	return nil
}

func (r *RepositoryImpl) UpdateProfile(ctx context.Context, profile domain.Profile) error {
	query := `UPDATE user_profiles SET username = ?, full_name = ?, phone_number = ?, avatar_url = ? WHERE user_id = ?`

	result, err := r.db.ExecContext(ctx, query, profile.Username, profile.FullName, profile.PhoneNumber, profile.AvatarURL, profile.UserID)
	if err != nil {
		return apperror.New(apperror.ErrInternalServer, "gagal memperbarui data profile, karna gangguan sistem")
	}

	rwAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rwAffected == 0 {
		apperror.New(apperror.ErrBadRequest, "gagal memperbarui data profile, user ID tidak valid")
	}

	return nil
}

func (r *RepositoryImpl) CreateAddress(ctx context.Context, address domain.Address) error {
	query := `INSERT INTO user_addresses (usr_address_id, user_id, recipient_name, phone, full_address, city, postal_code, is_default) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`

	_, err := r.db.ExecContext(ctx, query, address.ID, address.UserID, address.RecipientName, address.Phone, address.FullAddress, address.City, address.PostalCode, address.IsDefault)
	if err != nil {
		var mysqlErr *mysql.MySQLError
		if errors.As(err, &mysqlErr) {
			if mysqlErr.Number == 1062 {
				return apperror.New(apperror.ErrBadRequest, "alamat pengguna tersebut sudah pernah dibuat")
			}
		}
		return apperror.New(apperror.ErrInternalServer, "gagal menyimpan data alamat, karna gangguan sistem")
	}

	return nil
}
func (r *RepositoryImpl) UpdateAddress(ctx context.Context, address domain.Address) error {
	query := `UPDATE user_addresses SET recipient_name = ?, phone = ?, full_address = ?, city = ?, postal_code = ?, is_default = ? WHERE user_id = ?`

	result, err := r.db.ExecContext(ctx, query, address.RecipientName, address.Phone, address.FullAddress, address.City, address.PostalCode, address.IsDefault, address.UserID)
	if err != nil {
		return apperror.New(apperror.ErrInternalServer, "gagal memperbarui data alamat, karna gangguan sistem")
	}

	rwAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rwAffected == 0 {
		apperror.New(apperror.ErrBadRequest, "gagal memperbarui data alamat, user ID tidak valid")
	}

	return nil
}

func (r *RepositoryImpl) GetAddressByUserID(ctx context.Context, userID string) ([]domain.Address, error) {
	query := `SELECT usr_address_id, user_id, recipient_name, phone, full_address, city, postal_code, is_default FROM user_addresses WHERE user_id = ?`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return []domain.Address{}, apperror.New(apperror.ErrInternalServer, "gagal mengambil data alamat, karna gangguan sistem")
	}
	defer rows.Close()

	userAddresses := make([]domain.Address, 0, 5)
	for rows.Next() {
		var userAddress domain.Address

		err := rows.Scan(&userAddress.ID, &userAddress.UserID, &userAddress.RecipientName, &userAddress.Phone, &userAddress.FullAddress, &userAddress.City, &userAddress.PostalCode, &userAddress.IsDefault)
		if err != nil {
			return []domain.Address{}, err
		}

		userAddresses = append(userAddresses, userAddress)
	}

	if err := rows.Err(); err != nil {
		return []domain.Address{}, err
	}

	return userAddresses, nil
}
