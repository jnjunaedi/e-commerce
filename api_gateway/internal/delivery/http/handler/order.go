package handler

import (
	"e-commerce/api_gateway/internal/delivery/http/request"
	"e-commerce/api_gateway/internal/delivery/http/response"
	"e-commerce/api_gateway/pkg/validator"
	orderpb "e-commerce/pb/order"
	"e-commerce/pkg/utils"
	"e-commerce/pkg/utils/errgrpc"
	"net/http"

	"github.com/go-chi/chi/v5"
)

type OrderHandlerImpl struct {
	client orderpb.OrdersServiceClient
}

func NewOrderHandler(client orderpb.OrdersServiceClient) *OrderHandlerImpl {
	return &OrderHandlerImpl{client: client}
}

func (h *OrderHandlerImpl) CreateOrderHandler(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	ctx := r.Context()

	var orderReq request.OrderRequest

	if err := utils.ReadJSON(r, &orderReq); err != nil {
		utils.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "Format JSON tidak valid"})
		return
	}

	if errs := validator.ValidateStruct(orderReq); errs != nil {
		utils.WriteJSON(w, http.StatusBadRequest, map[string]any{
			"message": "validasi payload gagal",
			"error":   errs,
		})
		return
	}

	orderItemsGRPCResult := make([]*orderpb.OrderItemRequest, 0, len(orderReq.Items))
	for _, item := range orderReq.Items {
		orderItemsGRPCResult = append(orderItemsGRPCResult, &orderpb.OrderItemRequest{
			ProductId:   item.ProductID,
			FlashSaleId: item.FlashSaleID,
			Quantity:    item.Quantity,
		})
	}

	grpcResp, errGRPC := h.client.CreateOrder(ctx, &orderpb.CreateOrderRequest{UserId: orderReq.UserID, IdempotencyKey: orderReq.IdempotencyKey, Items: orderItemsGRPCResult})
	if errGRPC != nil {
		errgrpc.WriteGRPCError(w, errGRPC)
		return
	}

	orderItemsResult := make([]response.OrderItem, 0, len(orderReq.Items))
	for _, item := range grpcResp.GetItems() {
		orderItemsResult = append(orderItemsResult, response.OrderItem{
			ID:          item.GetItemId(),
			ProductID:   item.GetProductId(),
			FlashSaleID: item.GetFlashSaleId(),
			Quantity:    item.GetQuantity(),
			UnitPrice:   item.GetUnitPrice(),
		})
	}

	orderResp := response.OrderResponse{
		ID:             grpcResp.GetOrderId(),
		UserID:         grpcResp.GetUserId(),
		TotalPrice:     grpcResp.GetTotalPrice(),
		OrderStatus:    grpcResp.GetOrderStatus(),
		IdempotencyKey: grpcResp.GetIdempotencyKey(),
		Items:          orderItemsResult,
	}

	utils.WriteJSON(w, http.StatusOK, map[string]any{
		"data": orderResp,
	})
}

func (h *OrderHandlerImpl) GetOrderHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	idReq := chi.URLParam(r, "id")

	resGRPC, errGRPC := h.client.GetOrder(ctx, &orderpb.GetOrderRequest{OrderId: idReq})
	if errGRPC != nil {
		errgrpc.WriteGRPCError(w, errGRPC)
		return
	}

	orderItemsResult := make([]response.OrderItem, 0, len(resGRPC.Items))
	for _, item := range resGRPC.GetItems() {
		orderItemsResult = append(orderItemsResult, response.OrderItem{
			ID:          item.GetItemId(),
			ProductID:   item.GetProductId(),
			FlashSaleID: item.GetFlashSaleId(),
			Quantity:    item.GetQuantity(),
			UnitPrice:   item.GetUnitPrice(),
		})
	}

	orderResp := response.OrderResponse{
		ID:             resGRPC.GetOrderId(),
		UserID:         resGRPC.GetUserId(),
		TotalPrice:     resGRPC.GetTotalPrice(),
		OrderStatus:    resGRPC.GetOrderStatus(),
		IdempotencyKey: resGRPC.GetIdempotencyKey(),
		Items:          orderItemsResult,
	}

	utils.WriteJSON(w, http.StatusCreated, map[string]any{
		"data": orderResp,
	})

}
