package handler

import (
	"context"
	"e-commerce/api_gateway/internal/delivery/http/request"
	"e-commerce/api_gateway/internal/delivery/http/response"
	"e-commerce/api_gateway/pkg/validator"
	paymentpb "e-commerce/pb/payment"
	"e-commerce/pkg/utils"
	"e-commerce/pkg/utils/errgrpc"
	"net/http"
	"time"
)

type PaymentHandlerImpl struct {
	client paymentpb.PaymentsServiceClient
}

func NewPaymentHandler(client paymentpb.PaymentsServiceClient) *PaymentHandlerImpl {
	return &PaymentHandlerImpl{client: client}
}

func (h *PaymentHandlerImpl) ProcessPaymentHandler(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	ctx := r.Context()

	var payReq request.PaymentRequest

	if err := utils.ReadJSON(r, &payReq); err != nil {
		utils.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "Format JSON tidak valid"})
		return
	}

	if errs := validator.ValidateStruct(payReq); errs != nil {
		utils.WriteJSON(w, http.StatusBadRequest, map[string]any{
			"message": "validasi payload gagal",
			"error":   errs,
		})
		return
	}

	grpcResp, errGRPC := h.client.ProcessPayment(ctx, &paymentpb.ProcessPaymentRequest{OrderId: payReq.OrderID, PaymentMethod: payReq.PaymentMethod, PaymentAmount: payReq.PaymentAmount})
	if errGRPC != nil {
		errgrpc.WriteGRPCError(w, errGRPC)
		return
	}

	respClient := response.PaymentResponse{
		PaymentID:     grpcResp.GetPaymentId(),
		OrderID:       grpcResp.GetOrderId(),
		PaymentStatus: grpcResp.GetPaymentStatus(),
		PaymentMethod: grpcResp.GetPaymentMethod(),
		PaymentAmount: grpcResp.GetPaymentAmount(),
		Message:       grpcResp.GetMessage(),
	}

	utils.WriteJSON(w, http.StatusOK, map[string]any{
		"data": respClient,
	})
}

func (h *PaymentHandlerImpl) HandleMidtransWebhookHandler(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	ctx := r.Context()

	var payload map[string]any

	if err := utils.ReadJSON(r, &payload); err != nil {
		utils.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "Format JSON tidak valid"})
		return
	}

	orderID, okOrder := payload["order_id"].(string)
	txStatus, okStatus := payload["transaction_status"].(string)
	fraudStatus, okFraud := payload["fraud_status"].(string)

	if !okOrder || !okStatus {
		utils.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "Payload Midtrans cacat atau tidak lengkap"})
		return
	}

	if !okFraud {
		fraudStatus = ""
	}

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	grpcResp, errGRPC := h.client.HandleWebhook(ctx, &paymentpb.HandleWebhookRequest{
		OrderId:           orderID,
		TransactionStatus: txStatus,
		FraudStatus:       fraudStatus,
	})

	if errGRPC != nil {
		errgrpc.WriteGRPCError(w, errGRPC)
		return
	}

	utils.WriteJSON(w, http.StatusOK, map[string]any{
		"message": grpcResp.GetMessage(),
	})

}
