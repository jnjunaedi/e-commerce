package usecase_test

import (
	"context"
	mId "e-commerce/pkg/utils/mocks"
	"e-commerce/product_service/internal/domain"
	"e-commerce/product_service/internal/repository/mocks"
	"e-commerce/product_service/internal/usecase"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestProductUsecase_RegisterProduct(t *testing.T) {
	mockRepo := mocks.NewRepository(t)
	mockId := mId.NewIDGenerator(t)

	prod := domain.Product{
		ProductCode: "LPT-1001",
		ProductName: "Laptop Thinkpad X1 Carbon",
		Price:       2500000,
		Stock:       100,
		Description: "Laptop powerfull dengan spek ekstrim",
		ImageURLs:   []string{"image1.png", "image2.png"},
	}

	idDummy := "prod-uuid-112"

	mockId.On("GenerateUUID").Return(idDummy, nil)

	mockRepo.On("Create", mock.Anything, mock.MatchedBy(func(p domain.Product) bool {
		return p.ID == idDummy && prod.ProductCode == p.ProductCode && prod.ProductName == p.ProductName && prod.Price == p.Price && prod.Stock == p.Stock && prod.Description == p.Description && len(p.ImageURLs) > 0
	})).Return(nil)

	uc := usecase.NewUsecase(mockRepo, mockId)
	ctx := context.Background()

	res, err := uc.RegisterProduct(ctx, prod)

	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.Equal(t, prod.ProductName, res.ProductName)
	assert.Equal(t, prod.Price, res.Price)
	assert.Equal(t, prod.Stock, res.Stock)

	mockRepo.AssertExpectations(t)
	mockId.AssertExpectations(t)
}

func TestProductUsecase_FindProductById(t *testing.T) {
	mockRepo := mocks.NewRepository(t)
	mockId := mId.NewIDGenerator(t)

	prod := domain.Product{
		ProductCode: "LPT-1001",
		ProductName: "Laptop Thinkpad X1 Carbon",
		Price:       2500000,
		Stock:       100,
		Description: "Laptop powerfull dengan spek ekstrim",
		ImageURLs:   []string{"image1.png", "image2.png"},
	}

	mockRepo.On("GetByID", mock.Anything, mock.Anything).Return(prod, nil)

	uc := usecase.NewUsecase(mockRepo, mockId)
	ctx := context.Background()

	res, err := uc.FindProduct(ctx, "PROD-1001")

	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.Equal(t, prod.ProductCode, res.ProductCode)

	mockRepo.AssertExpectations(t)
	mockId.AssertExpectations(t)
}

func TestProductUsecase_DeductRegularStock(t *testing.T) {
	mockRepo := mocks.NewRepository(t)
	mockId := mId.NewIDGenerator(t)

	prod := domain.Product{
		ProductCode: "LPT-1001",
		ProductName: "Laptop Thinkpad X1 Carbon",
		Price:       2500000,
		Stock:       100,
		Description: "Laptop powerfull dengan spek ekstrim",
		ImageURLs:   []string{"image1.png", "image2.png"},
	}

	mockRepo.On("UpdateRegularStock", mock.Anything, mock.Anything, mock.AnythingOfType("int32")).Return(nil)

	uc := usecase.NewUsecase(mockRepo, mockId)
	ctx := context.Background()

	err := uc.DeductRegularStock(ctx, prod.ID, 10)

	assert.NoError(t, err)

	mockRepo.AssertExpectations(t)
	mockId.AssertExpectations(t)
}

func TestProductUsecase_StartFlashSale(t *testing.T) {
	mockRepo := mocks.NewRepository(t)
	mockId := mId.NewIDGenerator(t)

	prodFS := domain.ProductFlashSale{
		ProductID:          "PRD-100",
		FlashSalePrice:     12000,
		FlashSaleStock:     10,
		MaxPurchasePerUser: 1,
		StartTime:          time.Now().Add(5 * time.Second).Unix(),
		EndTime:            time.Now().Add(10 * time.Second).Unix(),
		Version:            0,
	}

	idDummy := "prod-uuid-112"

	mockId.On("GenerateUUID").Return(idDummy, nil)

	mockRepo.On("CreateFlashSale", mock.Anything, mock.MatchedBy(func(p domain.ProductFlashSale) bool {
		return p.ID == idDummy && prodFS.FlashSalePrice == p.FlashSalePrice && prodFS.FlashSaleStock == p.FlashSaleStock && p.MaxPurchasePerUser == 1 && prodFS.StartTime == p.StartTime && prodFS.EndTime == p.EndTime && prodFS.Version == p.Version
	})).Return(nil)

	uc := usecase.NewUsecase(mockRepo, mockId)
	ctx := context.Background()

	res, err := uc.StartFlashSaleEvent(ctx, prodFS)

	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.Equal(t, prodFS.FlashSalePrice, res.FlashSalePrice)
	assert.Equal(t, prodFS.FlashSaleStock, res.FlashSaleStock)
	assert.Equal(t, prodFS.MaxPurchasePerUser, res.MaxPurchasePerUser)

	mockRepo.AssertExpectations(t)
	mockId.AssertExpectations(t)
}

func TestProductUsecase_FindProductFlashSaleById(t *testing.T) {
	mockRepo := mocks.NewRepository(t)
	mockId := mId.NewIDGenerator(t)

	prodFS := domain.ProductFlashSale{
		ProductID:          "PRD-100",
		FlashSalePrice:     12000,
		FlashSaleStock:     10,
		MaxPurchasePerUser: 1,
		StartTime:          time.Now().Unix(),
		EndTime:            1798736340,
		Version:            0,
	}

	mockRepo.On("GetFlashSaleByID", mock.Anything, mock.Anything).Return(prodFS, nil)

	uc := usecase.NewUsecase(mockRepo, mockId)
	ctx := context.Background()

	res, err := uc.FindFlashSale(ctx, "PRD-100")
	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.Equal(t, prodFS.FlashSalePrice, res.FlashSalePrice)
	assert.Equal(t, prodFS.FlashSaleStock, res.FlashSaleStock)
	assert.Equal(t, prodFS.MaxPurchasePerUser, res.MaxPurchasePerUser)

	mockRepo.AssertExpectations(t)
	mockId.AssertExpectations(t)
}

func TestProductUsecase_DeductFlashSaleStock(t *testing.T) {
	mockRepo := mocks.NewRepository(t)
	mockId := mId.NewIDGenerator(t)

	prod := domain.Product{
		ProductCode: "LPT-1001",
		ProductName: "Laptop Thinkpad X1 Carbon",
		Price:       2500000,
		Stock:       100,
		Description: "Laptop powerfull dengan spek ekstrim",
		ImageURLs:   []string{"image1.png", "image2.png"},
	}

	prodFS := domain.ProductFlashSale{
		ProductID:          "PRD-100",
		FlashSalePrice:     12000,
		FlashSaleStock:     10,
		MaxPurchasePerUser: 1,
		StartTime:          1793157000,
		EndTime:            1798736340,
		Version:            0,
	}

	mockRepo.On("GetFlashSaleByID", mock.Anything, mock.Anything).Return(prodFS, nil).Once()

	mockRepo.On("UpdateFlashSaleStock", mock.Anything, mock.Anything, mock.AnythingOfType("int32"), mock.AnythingOfType("int32")).Return(nil).Once()

	uc := usecase.NewUsecase(mockRepo, mockId)
	ctx := context.Background()

	ok, msg, err := uc.DeductFlashSaleStock(ctx, prod.ID, prodFS.ID, 1)

	assert.NoError(t, err)
	assert.True(t, ok)
	assert.Equal(t, "sukses memotong stok", msg)

	mockRepo.AssertExpectations(t)
	mockId.AssertExpectations(t)
}

func TestProductUsecase_StreamStock(t *testing.T) {
	mockRepo := mocks.NewRepository(t)
	mockId := mId.NewIDGenerator(t)

	prodFS := domain.ProductFlashSale{
		ID:                 "PRDFS-100",
		ProductID:          "PRD-100",
		FlashSalePrice:     12000,
		FlashSaleStock:     10,
		MaxPurchasePerUser: 1,
		StartTime:          1793157000,
		EndTime:            1798736340,
		Version:            0,
	}

	stChan := make(chan int32, 10)

	mockRepo.On("GetFlashSaleByID", mock.Anything, mock.Anything).Return(prodFS, nil)

	mockRepo.On("SubscribeStockChanges", mock.Anything, mock.Anything).Return((<-chan int32)(stChan), nil)

	uc := usecase.NewUsecase(mockRepo, mockId)
	ctx := context.Background()

	ch, err := uc.StreamStock(ctx, "PRDFS-100")

	assert.NoError(t, err)
	assert.NotNil(t, ch)

	go func() {
		stChan <- 5
	}()

	select {
	case res := <-ch:
		assert.Equal(t, int32(5), res)
	case <-time.After(1 * time.Second):
		assert.Fail(t, "Timeout menunggu data stock")
	}

	mockRepo.AssertExpectations(t)
	mockId.AssertExpectations(t)
}
