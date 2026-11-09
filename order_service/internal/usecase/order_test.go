package usecase_test

import (
	"context"
	"e-commerce/order_service/infrastructure/client"
	mGrpc "e-commerce/order_service/infrastructure/client/mocks"
	mKafka "e-commerce/order_service/infrastructure/kafka/mocks"
	"e-commerce/order_service/internal/domain"
	mRepo "e-commerce/order_service/internal/repository/mocks"
	mTxManager "e-commerce/order_service/internal/repository/mocks"
	"e-commerce/order_service/internal/usecase"
	mId "e-commerce/pkg/utils/mocks"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestOrderUsecase_CreateCheckoutFlashSale(t *testing.T) {
	mockRepo := mRepo.NewRepository(t)
	mockClient := mGrpc.NewProductGrpcClient(t)
	mockKafka := mKafka.NewOrderProducer(t)
	mockId := mId.NewIDGenerator(t)
	mockTxManager := mTxManager.NewTxManager(t)

	fsID := "FS-112"
	orderID := "order-uuid-111"
	userID := "usr-uuid-123"
	idempotencyKey := "ifnx-regular-1002"
	itemID := "item-uuid-222"
	outboxID := "outbox-uuid-123"

	orderInput := domain.Order{
		UserID:         userID,
		TotalPrice:     500000,
		IdempotencyKey: idempotencyKey,
		Items: []domain.OrderItem{
			{
				OrderID:     "ODR-112",
				ProductID:   "PRD-122",
				FlashSaleID: &fsID,
				Quantity:    5,
				UnitPrice:   100000,
				Subtotal:    500000,
			},
		},
	}

	fsProd := client.FlashSaleProductResult{
		FlashSaleID:        "FS-112",
		ProductID:          "PRD-122",
		FlashSalePrice:     100000,
		FlashSaleStock:     10,
		MaxPurchasePerUser: 5,
	}

	newOrder := domain.Order{
		ID:             orderID,
		UserID:         userID,
		TotalPrice:     500000,
		OrderStatus:    "PENDING_PAYMENT",
		IdempotencyKey: idempotencyKey,
		Items: []domain.OrderItem{
			{
				ID:          itemID,
				OrderID:     orderID,
				ProductID:   "PRD-122",
				FlashSaleID: &fsID,
				Quantity:    5,
				UnitPrice:   100000,
				Subtotal:    500000,
			},
		},
		CreatedAt: time.Date(1, time.January, 1, 0, 0, 0, 0, time.UTC),
	}

	mockRepo.On("CheckByIdempotencyKey", mock.Anything, idempotencyKey).Return(false, nil).Once()

	mockId.On("GenerateUUID").Return(orderID, nil).Once()

	mockTxManager.On("WithTransaction", mock.Anything, mock.Anything).Return(func(ctx context.Context, fn func(ctx context.Context) error) error {
		return fn(ctx)
	}).Once()

	mockRepo.On("DeductLocalFlashSaleStock", mock.Anything, mock.Anything, mock.Anything).Return(int64(10), nil)

	mockClient.On("FindFlashSale", mock.Anything, mock.Anything).Return(fsProd, nil).Once()

	mockId.On("GenerateUUID").Return(outboxID, nil).Once()

	mockRepo.On("CreateOutbox", mock.Anything, outboxID, "DeductFlashSaleStock", mock.Anything).Return(nil).Once()

	mockId.On("GenerateUUID").Return(itemID, nil).Once()

	mockRepo.On("Create", mock.Anything, mock.MatchedBy(func(o domain.Order) bool {
		return o.ID == orderID && o.TotalPrice == orderInput.TotalPrice && len(o.Items) == 1 && o.Items[0].ID == itemID
	})).Return(nil).Once()

	uc := usecase.NewUsecase(mockRepo, mockClient, mockTxManager, mockKafka, mockId)
	ctx := context.Background()

	res, err := uc.CreateCheckout(ctx, orderInput)

	assert.NoError(t, err)
	assert.Equal(t, newOrder, res)
	assert.Len(t, res.Items, 1)

	mockRepo.AssertExpectations(t)
	mockClient.AssertExpectations(t)
	mockKafka.AssertExpectations(t)
	mockId.AssertExpectations(t)
	mockTxManager.AssertExpectations(t)
}

func TestOrderUsecase_CreateCheckoutRegular(t *testing.T) {
	mockRepo := mRepo.NewRepository(t)
	mockClient := mGrpc.NewProductGrpcClient(t)
	mockKafka := mKafka.NewOrderProducer(t)
	mockId := mId.NewIDGenerator(t)
	mockTxManager := mTxManager.NewTxManager(t)

	fsID := ""
	orderID := "order-uuid-111"
	userID := "USR-112"
	idempotencyKey := "ifnx-reguler-1002"
	outboxID := "outbox-123"
	itemID := "item-uuid-222"

	orderInput := domain.Order{
		UserID:         userID,
		TotalPrice:     17500,
		IdempotencyKey: idempotencyKey,
		Items: []domain.OrderItem{
			{
				OrderID:     "ODR-112",
				ProductID:   "PRD-122",
				FlashSaleID: &fsID,
				Quantity:    5,
				UnitPrice:   3500,
				Subtotal:    17500,
			},
		},
	}

	prod := client.RegularProductResult{
		ProductCode: "112",
		ProductName: "Mie Ayam bawang",
		CategoryID:  "mie-instan-122",
		Price:       3500,
		Stock:       100,
	}

	newOrder := domain.Order{
		ID:             orderID,
		UserID:         userID,
		TotalPrice:     17500,
		OrderStatus:    "PENDING_PAYMENT",
		IdempotencyKey: idempotencyKey,
		Items: []domain.OrderItem{
			{
				ID:          itemID,
				OrderID:     orderID,
				ProductID:   "PRD-122",
				FlashSaleID: &fsID,
				Quantity:    5,
				UnitPrice:   3500,
				Subtotal:    17500,
			},
		},
		CreatedAt: time.Date(1, time.January, 1, 0, 0, 0, 0, time.UTC),
	}

	mockRepo.On("CheckByIdempotencyKey", mock.Anything, idempotencyKey).Return(false, nil).Once()
	mockId.On("GenerateUUID").Return(orderID, nil).Once()

	mockTxManager.On("WithTransaction", mock.Anything, mock.Anything).Return(func(ctx context.Context, fn func(ctx context.Context) error) error {
		return fn(ctx)
	}).Once()

	mockClient.On("FindRegularProduct", mock.Anything, mock.Anything).Return(prod, nil).Once()
	mockId.On("GenerateUUID").Return(outboxID, nil).Once()
	mockRepo.On("CreateOutbox", mock.Anything, outboxID, "DeductRegularStock", mock.Anything).Return(nil).Once()
	mockId.On("GenerateUUID").Return(itemID, nil).Once()

	mockRepo.On("Create", mock.Anything, mock.MatchedBy(func(o domain.Order) bool {
		return o.ID == orderID && o.TotalPrice == orderInput.TotalPrice && len(o.Items) == 1 && o.Items[0].ID == itemID
	})).Return(nil).Once()

	uc := usecase.NewUsecase(mockRepo, mockClient, mockTxManager, mockKafka, mockId)
	ctx := context.Background()

	res, err := uc.CreateCheckout(ctx, orderInput)

	assert.NoError(t, err)
	assert.Equal(t, newOrder, res)
	assert.Len(t, res.Items, 1)

	mockRepo.AssertExpectations(t)
	mockClient.AssertExpectations(t)
	mockKafka.AssertExpectations(t)
	mockId.AssertExpectations(t)
	mockTxManager.AssertExpectations(t)
}
func TestOrderUsecase_GetOrderDetails(t *testing.T) {
	mockRepo := mRepo.NewRepository(t)
	mockClient := mGrpc.NewProductGrpcClient(t)
	mockKafka := mKafka.NewOrderProducer(t)
	mockId := mId.NewIDGenerator(t)
	mockTxManager := mTxManager.NewTxManager(t)

	fsID := "FS-112"

	order := domain.Order{
		UserID:         "USR-112",
		TotalPrice:     600000,
		IdempotencyKey: "ifnx-reguler-1002",
		Items: []domain.OrderItem{
			domain.OrderItem{
				OrderID:     "ODR-112",
				ProductID:   "PRD-122",
				FlashSaleID: &fsID,
				Quantity:    5,
				UnitPrice:   100000,
				Subtotal:    500000,
			},
		},
	}

	mockRepo.On("GetByID", mock.Anything, mock.Anything).Return(order, nil)

	uc := usecase.NewUsecase(mockRepo, mockClient, mockTxManager, mockKafka, mockId)
	ctx := context.Background()

	res, err := uc.GetOrderDetails(ctx, "ODR-112")

	assert.NoError(t, err)
	assert.Len(t, res.Items, 1)
	assert.Equal(t, order.UserID, res.UserID)
	assert.Equal(t, order.TotalPrice, res.TotalPrice)

	mockRepo.AssertExpectations(t)
	mockClient.AssertExpectations(t)
	mockKafka.AssertExpectations(t)
	mockId.AssertExpectations(t)
	mockTxManager.AssertExpectations(t)
}

func TestOrderUsecase_UpdateStatus(t *testing.T) {
	mockRepo := mRepo.NewRepository(t)
	mockClient := mGrpc.NewProductGrpcClient(t)
	mockKafka := mKafka.NewOrderProducer(t)
	mockId := mId.NewIDGenerator(t)
	mockTxManager := mTxManager.NewTxManager(t)

	mockRepo.On("UpdateStatus", mock.Anything, mock.Anything, mock.Anything).Return(nil)

	uc := usecase.NewUsecase(mockRepo, mockClient, mockTxManager, mockKafka, mockId)
	ctx := context.Background()

	err := uc.UpdateOrderStatus(ctx, "ODR-112", "PENDING_PAYMENT")

	assert.NoError(t, err)

	mockRepo.AssertExpectations(t)
	mockClient.AssertExpectations(t)
	mockKafka.AssertExpectations(t)
	mockId.AssertExpectations(t)
	mockTxManager.AssertExpectations(t)
}
