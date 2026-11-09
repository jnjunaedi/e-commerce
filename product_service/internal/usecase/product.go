package usecase

import (
	"context"
	"e-commerce/pkg/utils"
	"e-commerce/pkg/utils/apperror"
	"e-commerce/product_service/internal/domain"
	"time"
)

type UsecaseImpl struct {
	repo        domain.Repository
	idGenerator utils.IDGenerator
}

func NewUsecase(repo domain.Repository, idGenerator utils.IDGenerator) domain.Usecase {
	return &UsecaseImpl{repo: repo, idGenerator: idGenerator}
}

func (u *UsecaseImpl) RegisterProduct(ctx context.Context, p domain.Product) (domain.Product, error) {
	productID, err := u.idGenerator.GenerateUUID()
	if err != nil {
		return domain.Product{}, apperror.New(apperror.ErrInternalServer, "error internal server")
	}

	newProduct := domain.Product{
		ID:          productID,
		ProductCode: p.ProductCode,
		ProductName: p.ProductName,
		CategoryID:  p.CategoryID,
		Price:       p.Price,
		Stock:       p.Stock,
		Description: p.Description,
		ImageURLs:   p.ImageURLs,
	}

	err = u.repo.Create(ctx, newProduct)
	if err != nil {
		return domain.Product{}, err
	}

	return newProduct, nil
}

func (u *UsecaseImpl) FindProduct(ctx context.Context, id string) (domain.Product, error) {
	return u.repo.GetByID(ctx, id)
}

func (u *UsecaseImpl) DeductRegularStock(ctx context.Context, productID string, qty int32) error {
	return u.repo.UpdateRegularStock(ctx, productID, qty)
}

func (u *UsecaseImpl) StartFlashSaleEvent(ctx context.Context, fs domain.ProductFlashSale) (domain.ProductFlashSale, error) {
	now := time.Now().Unix()

	flashSaleID, err := u.idGenerator.GenerateUUID()
	if err != nil {
		return domain.ProductFlashSale{}, err
	}

	if fs.StartTime <= now {
		return domain.ProductFlashSale{}, apperror.New(apperror.ErrBadRequest, "waktu mulai harus berada di masa depan")
	}

	if fs.EndTime <= fs.StartTime {
		return domain.ProductFlashSale{}, apperror.New(apperror.ErrBadRequest, "waktu selesai harus lebih besar dari waktu mulai")
	}

	newProductFs := domain.ProductFlashSale{ID: flashSaleID, ProductID: fs.ProductID, FlashSalePrice: fs.FlashSalePrice, FlashSaleStock: fs.FlashSaleStock, MaxPurchasePerUser: fs.MaxPurchasePerUser, StartTime: fs.StartTime, EndTime: fs.EndTime, Version: 0}

	err = u.repo.CreateFlashSale(ctx, newProductFs)
	if err != nil {
		return domain.ProductFlashSale{}, err
	}

	return newProductFs, nil
}

func (u *UsecaseImpl) FindFlashSale(ctx context.Context, id string) (domain.ProductFlashSale, error) {
	return u.repo.GetFlashSaleByID(ctx, id)
}

func (u *UsecaseImpl) DeductFlashSaleStock(ctx context.Context, productID string, flashSaleID string, qty int32) (bool, string, error) {
	maxRetries := 3
	for i := 0; i < maxRetries; i++ {
		fs, err := u.repo.GetFlashSaleByID(ctx, flashSaleID)
		if err != nil {
			return false, "", err
		}

		now := time.Now().Unix()
		if now < fs.StartTime {
			return false, "event flash sale belum dimulai", nil
		}
		if now > fs.EndTime {
			return false, "event flash sale sudah berakhir", nil
		}
		if fs.FlashSaleStock < qty {
			return false, "stok produk flash sale sudah habis", nil
		}
		if qty > fs.MaxPurchasePerUser {
			return false, "pembelian melebihi batas maksimal per user", nil
		}

		err = u.repo.UpdateFlashSaleStock(ctx, flashSaleID, qty, fs.Version)
		if err == nil {
			return true, "sukses memotong stok", nil
		}

		if err.Error() == "optimistic_lock_failed" {
			time.Sleep(10 * time.Millisecond)
			continue
		}

		return false, "", err
	}

	return false, "gagal mendapatkan slot transaksi, silakan coba lagi", nil
}

func (u *UsecaseImpl) StreamStock(ctx context.Context, flashSaleID string) (<-chan int32, error) {
	_, err := u.repo.GetFlashSaleByID(ctx, flashSaleID)
	if err != nil {
		return nil, err
	}

	return u.repo.SubscribeStockChanges(ctx, flashSaleID)
}
