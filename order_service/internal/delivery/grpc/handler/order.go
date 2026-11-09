package handler

import (
	"context"
	"e-commerce/order_service/internal/domain"
	orderpb "e-commerce/pb/order"
	"e-commerce/pkg/utils/errgrpc"

	"google.golang.org/protobuf/types/known/timestamppb"
)

type gRPCHandlerImpl struct {
	orderpb.UnimplementedOrdersServiceServer
	uc domain.Usecase
}

func NewGRPCHandler(uc domain.Usecase) orderpb.OrdersServiceServer {
	return &gRPCHandlerImpl{uc: uc}
}

func (h *gRPCHandlerImpl) CreateOrder(ctx context.Context, req *orderpb.CreateOrderRequest) (*orderpb.OrdersResponse, error) {
	itemsUc := make([]domain.OrderItem, 0, len(req.GetItems()))

	for _, item := range req.GetItems() {
		fsID := item.GetFlashSaleId()
		itemsUc = append(itemsUc, domain.OrderItem{
			ProductID:   item.GetProductId(),
			FlashSaleID: &fsID,
			Quantity:    item.GetQuantity(),
		})
	}

	res, err := h.uc.CreateCheckout(ctx, domain.Order{UserID: req.GetUserId(), IdempotencyKey: req.GetIdempotencyKey(), Items: itemsUc})
	if err != nil {
		return nil, errgrpc.ReadInternalError(err)
	}

	itemsGRPC := make([]*orderpb.OrderItemResponse, 0, len(req.GetItems()))
	for _, item := range res.Items {
		fsResult := ""
		if item.FlashSaleID != nil {
			fsResult = *item.FlashSaleID
		}

		itemsGRPC = append(itemsGRPC, &orderpb.OrderItemResponse{
			ItemId:      item.ID,
			ProductId:   item.ProductID,
			FlashSaleId: fsResult,
			Quantity:    item.Quantity,
			UnitPrice:   item.UnitPrice,
		})
	}

	return &orderpb.OrdersResponse{OrderId: res.ID, UserId: res.UserID, TotalPrice: res.TotalPrice, OrderStatus: res.OrderStatus, IdempotencyKey: res.IdempotencyKey, Items: itemsGRPC, CreatedAt: timestamppb.New(res.CreatedAt)}, nil
}

func (h *gRPCHandlerImpl) GetOrder(ctx context.Context, req *orderpb.GetOrderRequest) (*orderpb.OrdersResponse, error) {
	res, err := h.uc.GetOrderDetails(ctx, req.GetOrderId())
	if err != nil {
		return nil, errgrpc.ReadInternalError(err)
	}

	itemsGRPC := make([]*orderpb.OrderItemResponse, 0, len(res.Items))
	for _, item := range res.Items {

		itemsGRPC = append(itemsGRPC, &orderpb.OrderItemResponse{
			ItemId:      item.ID,
			ProductId:   item.ProductID,
			FlashSaleId: *item.FlashSaleID,
			Quantity:    item.Quantity,
			UnitPrice:   item.UnitPrice,
		})
	}

	return &orderpb.OrdersResponse{OrderId: res.ID, UserId: res.UserID, TotalPrice: res.TotalPrice, OrderStatus: res.OrderStatus, IdempotencyKey: res.IdempotencyKey, Items: itemsGRPC, CreatedAt: timestamppb.New(res.CreatedAt)}, nil
}

func (h *gRPCHandlerImpl) UpdateOrderStatus(ctx context.Context, req *orderpb.UpdateOrderStatusRequest) (*orderpb.UpdateOrderStatusResponse, error) {
	err := h.uc.UpdateOrderStatus(ctx, req.GetOrderId(), req.GetStatus())
	if err != nil {
		return nil, errgrpc.ReadInternalError(err)
	}

	return &orderpb.UpdateOrderStatusResponse{Success: true,
		Message: "Status order sukses diperbarui"}, nil
}
