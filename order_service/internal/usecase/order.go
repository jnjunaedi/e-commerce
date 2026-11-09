package usecase

import (
	"context"
	"e-commerce/order_service/infrastructure/client"
	"e-commerce/order_service/infrastructure/kafka"
	"e-commerce/order_service/internal/domain"
	"e-commerce/pkg/utils"
	"e-commerce/pkg/utils/apperror"
	"encoding/json"
)

type OrderUsecaseImpl struct {
	repo          domain.Repository
	txManager     domain.TxManager
	client        client.ProductGrpcClient
	kafkaProcedur kafka.OrderProducer
	idGenerator   utils.IDGenerator
}

func NewUsecase(repo domain.Repository, client client.ProductGrpcClient, txManager domain.TxManager, kafkaProcedur kafka.OrderProducer, idGenerator utils.IDGenerator) domain.Usecase {
	return &OrderUsecaseImpl{repo: repo, client: client, txManager: txManager, kafkaProcedur: kafkaProcedur, idGenerator: idGenerator}
}

func (u *OrderUsecaseImpl) CreateCheckout(ctx context.Context, order domain.Order) (domain.Order, error) {
	isExist, err := u.repo.CheckByIdempotencyKey(ctx, order.IdempotencyKey)
	if err != nil {
		return domain.Order{}, err
	}

	if isExist {
		return domain.Order{}, apperror.New(apperror.ErrBadRequest, err.Error())
	}

	orderID, err := u.idGenerator.GenerateUUID()
	if err != nil {
		return domain.Order{}, err
	}

	var resultOrder domain.Order

	err = u.txManager.WithTransaction(ctx, func(ctx context.Context) error {
		var totalPrice int64
		finalizedItems := make([]domain.OrderItem, 0, len(order.Items))

		for _, item := range order.Items {
			var unitPrice int64
			var subTotal int64

			if item.FlashSaleID != nil && *item.FlashSaleID != "" {
				fsID := ""
				if item.FlashSaleID != nil {
					fsID = *item.FlashSaleID
				}

				rowsAffected, err := u.repo.DeductLocalFlashSaleStock(ctx, fsID, item.Quantity)
				if err != nil {
					return err
				}

				if rowsAffected == 0 {
					return apperror.New(apperror.ErrBadRequest, "stok produk flash sale sudah habis")
				}

				res, err := u.client.FindFlashSale(ctx, fsID)
				if err != nil {
					return err
				}

				if item.Quantity > res.MaxPurchasePerUser {
					return apperror.New(apperror.ErrBadRequest, "pembelian melebihi batas maksimal")
				}

				orderData := map[string]any{
					"order_id":      orderID,
					"user_id":       order.UserID,
					"product_id":    item.ProductID,
					"flash_sale_id": item.FlashSaleID,
					"quantity":      item.Quantity,
				}

				payloadBytes, err := json.Marshal(orderData)
				if err != nil {
					return err
				}

				orderboxID, err := u.idGenerator.GenerateUUID()
				if err != nil {
					return err
				}

				err = u.repo.CreateOutbox(ctx, orderboxID, "DeductFlashSaleStock", payloadBytes)
				if err != nil {
					return err
				}

				unitPrice = res.FlashSalePrice
				subTotal = res.FlashSalePrice * int64(item.Quantity)
			} else {
				res, err := u.client.FindRegularProduct(ctx, item.ProductID)
				if err != nil {
					return err
				}
				if res.Stock < item.Quantity {
					return apperror.New(apperror.ErrBadRequest, "Stok produk sudah habis")
				}

				orderData := map[string]any{
					"order_id":   orderID,
					"user_id":    order.UserID,
					"product_id": item.ProductID,
					"quantity":   item.Quantity,
				}

				payloadBytes, err := json.Marshal(orderData)
				if err != nil {
					return err
				}

				orderboxID, err := u.idGenerator.GenerateUUID()
				if err != nil {
					return err
				}

				err = u.repo.CreateOutbox(ctx, orderboxID, "DeductRegularStock", payloadBytes)
				if err != nil {
					return err
				}

				unitPrice = res.Price
				subTotal = res.Price * int64(item.Quantity)
			}

			itemID, err := u.idGenerator.GenerateUUID()
			if err != nil {
				return err
			}

			finalizedItems = append(finalizedItems, domain.OrderItem{
				ID:          itemID,
				OrderID:     orderID,
				ProductID:   item.ProductID,
				FlashSaleID: item.FlashSaleID,
				Quantity:    item.Quantity,
				UnitPrice:   unitPrice,
				Subtotal:    subTotal,
			})

			totalPrice += subTotal
		}

		newOrder := domain.Order{
			ID:             orderID,
			UserID:         order.UserID,
			TotalPrice:     totalPrice,
			OrderStatus:    "PENDING_PAYMENT",
			IdempotencyKey: order.IdempotencyKey,
			Items:          finalizedItems,
		}

		err = u.repo.Create(ctx, newOrder)
		if err != nil {
			return err
		}

		resultOrder = newOrder
		return nil
	})
	if err != nil {
		return domain.Order{}, err
	}

	return resultOrder, nil
}

func (u *OrderUsecaseImpl) GetOrderDetails(ctx context.Context, orderID string) (domain.Order, error) {
	return u.repo.GetByID(ctx, orderID)
}

func (u *OrderUsecaseImpl) UpdateOrderStatus(ctx context.Context, orderID string, status string) error {
	return u.repo.UpdateStatus(ctx, orderID, status)
}
