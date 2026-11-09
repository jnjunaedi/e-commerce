package repository

import (
	"context"
	"database/sql"
	"e-commerce/pkg/utils"
	"e-commerce/pkg/utils/apperror"
	"e-commerce/product_service/internal/domain"
	"errors"
	"sync"
	"time"

	"github.com/go-sql-driver/mysql"
)

type RepositoryImpl struct {
	db          *sql.DB
	mu          sync.RWMutex
	listeners   map[string]map[chan int32]struct{}
	idGenerator utils.IDGenerator
}

func NewRepository(db *sql.DB, idGenerator utils.IDGenerator) domain.Repository {
	return &RepositoryImpl{db: db, listeners: make(map[string]map[chan int32]struct{}), idGenerator: idGenerator}
}

func (r *RepositoryImpl) SubscribeStockChanges(ctx context.Context, flashSaleID string) (<-chan int32, error) {
	r.mu.Lock()

	if r.listeners == nil {
		r.listeners = make(map[string]map[chan int32]struct{})
	}

	if _, exixts := r.listeners[flashSaleID]; !exixts {
		r.listeners[flashSaleID] = make(map[chan int32]struct{})
	}

	ch := make(chan int32, 10)
	r.listeners[flashSaleID][ch] = struct{}{}

	r.mu.Unlock()

	go func() {
		<-ctx.Done()
		r.mu.Lock()

		if listeners, exists := r.listeners[flashSaleID]; exists {
			delete(listeners, ch)
			close(ch)

			if len(listeners) == 0 {
				delete(r.listeners, flashSaleID)
			}
		}
		r.mu.Unlock()
	}()

	return ch, nil
}

func (r *RepositoryImpl) BroadcastStockUpdate(flashSaleID string, latestStock int32) {
	r.mu.RLock()

	listeners, exists := r.listeners[flashSaleID]
	if !exists || len(listeners) == 0 {
		r.mu.RUnlock()
		return
	}

	channels := make([]chan int32, 0, len(listeners))
	for ch := range listeners {
		channels = append(channels, ch)
	}
	r.mu.RUnlock()

	go func(chans []chan int32, stock int32) {
		for _, ch := range chans {
			select {
			case ch <- stock:
			default:
			}
		}
	}(channels, latestStock)
}

func (r *RepositoryImpl) Create(ctx context.Context, p domain.Product) error {
	tx, err := r.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if err != nil {
		return err
	}
	defer tx.Rollback()

	querySaveProduct := `INSERT INTO products (product_id, product_code, product_name, category_id, price, stock, description) VALUES (?, ?, ?, ?, ?, ?, ?)`
	_, err = tx.ExecContext(ctx, querySaveProduct, p.ID, p.ProductCode, p.ProductName, p.CategoryID, p.Price, p.Stock, p.Description)
	if err != nil {
		var mysqlErr *mysql.MySQLError
		if errors.As(err, &mysqlErr) {
			if mysqlErr.Number == 1062 {
				return apperror.New(apperror.ErrBadRequest, "produk sudah pernah dibuat")
			}
		}
		return apperror.New(apperror.ErrInternalServer, "gagal menyimpan data produk, karna gangguan sistem")
	}

	querySaveImage := `INSERT INTO product_images (image_id, product_id, image_url) VALUES (?, ?, ?)`
	for _, url := range p.ImageURLs {
		genUUID, err := r.idGenerator.GenerateUUID()
		if err != nil {
			return err
		}
		_, err = tx.ExecContext(ctx, querySaveImage, genUUID, p.ID, url)
		if err != nil {
			var mysqlErr *mysql.MySQLError
			if errors.As(err, &mysqlErr) {
				if mysqlErr.Number == 1062 {
					return apperror.New(apperror.ErrBadRequest, "gambar tersebut sudah pernah dibuat")
				}
			}
			return apperror.New(apperror.ErrInternalServer, "gagal menyimpan gambar produk, karna gangguan sistem")
		}
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	return nil
}

func (r *RepositoryImpl) GetByID(ctx context.Context, id string) (domain.Product, error) {
	queryProd := `SELECT product_id, product_code, product_name, category_id, price, stock, description, created_at, updated_at FROM products WHERE product_id = ?`

	var product domain.Product
	err := r.db.QueryRowContext(ctx, queryProd, id).Scan(&product.ID, &product.ProductCode, &product.ProductName, &product.CategoryID, &product.Price, &product.Stock, &product.Description, &product.CreatedAt, &product.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return domain.Product{}, apperror.New(apperror.ErrNotFound, "produk tidak ditemukan")
		}
		return domain.Product{}, err
	}

	queryProdImage := `SELECT image_url FROM product_images WHERE product_id = ?`
	rows, err := r.db.QueryContext(ctx, queryProdImage, id)
	if err != nil {
		return domain.Product{}, err
	}
	defer rows.Close()

	for rows.Next() {
		var imageURL string

		if err := rows.Scan(&imageURL); err != nil {
			return domain.Product{}, err
		}
		product.ImageURLs = append(product.ImageURLs, imageURL)
	}

	if err := rows.Err(); err != nil {
		return domain.Product{}, err
	}

	return product, nil
}

func (r *RepositoryImpl) UpdateRegularStock(ctx context.Context, id string, quantity int32) error {
	tx, err := r.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var prodStock int32
	queryCheckStock := `SELECT stock FROM products WHERE product_id = ? FOR UPDATE`
	err = tx.QueryRowContext(ctx, queryCheckStock, id).Scan(&prodStock)
	if err != nil {
		if err == sql.ErrNoRows {
			return apperror.New(apperror.ErrBadRequest, "produk tidak ditemukan")
		}
		return apperror.New(apperror.ErrInternalServer, "gagal mengambil stok produk, karna gangguan sistem")
	}

	if prodStock < quantity {
		return apperror.New(apperror.ErrBadRequest, "stok produk tidak mencukupi")
	}

	resultStock := prodStock - quantity

	queryUpdateStock := `UPDATE products SET stock = ? WHERE product_id = ?`
	result, err := tx.ExecContext(ctx, queryUpdateStock, resultStock, id)
	if err != nil {
		return apperror.New(apperror.ErrInternalServer, "gagal memperbarui stok produk, karna gangguan sistem")
	}

	rwAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rwAffected <= 0 {
		return apperror.New(apperror.ErrBadRequest, "gagal memperbarui stock, product ID tidak valid atau stock sama")
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	return nil
}

func (r *RepositoryImpl) CreateFlashSale(ctx context.Context, fs domain.ProductFlashSale) error {
	query := `INSERT INTO product_flash_sales (flash_sale_id, product_id, flash_sale_price, flash_sale_stock, max_purchase_per_user, start_time, end_time, version) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`

	startTimeConvert := time.Unix(fs.StartTime, 0)
	endTimeConvert := time.Unix(fs.EndTime, 0)

	_, err := r.db.ExecContext(ctx, query, fs.ID, fs.ProductID, fs.FlashSalePrice, fs.FlashSaleStock, fs.MaxPurchasePerUser, startTimeConvert, endTimeConvert, fs.Version)
	if err != nil {
		var mysqlErr *mysql.MySQLError
		if errors.As(err, &mysqlErr) {
			if mysqlErr.Number == 1062 {
				return apperror.New(apperror.ErrBadRequest, "produk flash sale sudah pernah dibuat")
			}
		}
		return apperror.New(apperror.ErrInternalServer, "gagal menyimpan produk flash sale, karna gangguan sistem")
	}

	return nil
}

func (r *RepositoryImpl) GetFlashSaleByID(ctx context.Context, id string) (domain.ProductFlashSale, error) {
	query := `SELECT flash_sale_id, product_id, flash_sale_price, flash_sale_stock, max_purchase_per_user, start_time, end_time, version FROM product_flash_sales WHERE flash_sale_id = ?`

	var fsProduct domain.ProductFlashSale
	var startTime, endTime time.Time

	err := r.db.QueryRowContext(ctx, query, id).Scan(&fsProduct.ID, &fsProduct.ProductID, &fsProduct.FlashSalePrice, &fsProduct.FlashSaleStock, &fsProduct.MaxPurchasePerUser, &startTime, &endTime, &fsProduct.Version)
	if err != nil {
		if err == sql.ErrNoRows {
			return domain.ProductFlashSale{}, apperror.New(apperror.ErrNotFound, "Product Flash Sale tidak ditemukan")
		}
		return domain.ProductFlashSale{}, apperror.New(apperror.ErrInternalServer, "gagal mengambil stok produk flash sale, karna gangguan sistem")
	}

	fsProduct.StartTime = startTime.Unix()
	fsProduct.EndTime = endTime.Unix()

	return fsProduct, nil
}

func (r *RepositoryImpl) UpdateFlashSaleStock(ctx context.Context, id string, quantity int32, currentVersion int32) error {
	query := `UPDATE product_flash_sales 
	          SET flash_sale_stock = flash_sale_stock - ?, version = version + 1 
	          WHERE flash_sale_id = ? AND version = ? AND flash_sale_stock >= ?`

	result, err := r.db.ExecContext(ctx, query, quantity, id, currentVersion, quantity)
	if err != nil {
		return apperror.New(apperror.ErrInternalServer, "gagal memperbarui stok flash sale, karna gangguan sistem")
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return errors.New("optimistic_lock_failed")
	}

	var newStock int32
	err = r.db.QueryRowContext(ctx, "SELECT flash_sale_stock FROM product_flash_sales WHERE flash_sale_id = ?", id).Scan(&newStock)
	if err != nil {
		if err == sql.ErrNoRows {
			return apperror.New(apperror.ErrNotFound, "produk Flash Sale tidak ditemukan")
		}
		return apperror.New(apperror.ErrInternalServer, "gagal mengambil stok produk flash sale, karna gangguan sistem")
	}

	r.BroadcastStockUpdate(id, newStock)

	return nil
}
