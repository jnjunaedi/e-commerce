package handler

import (
	"context"
	"e-commerce/payment_service/internal/domain"
	paymentpb "e-commerce/pb/payment"
	"e-commerce/pkg/utils/errgrpc"
)

type gRPCHandlerImpl struct {
	paymentpb.UnimplementedPaymentsServiceServer
	uc domain.Usecase
}

func NewGRPCHandler(uc domain.Usecase) paymentpb.PaymentsServiceServer {
	return &gRPCHandlerImpl{uc: uc}
}

func (h *gRPCHandlerImpl) ProcessPayment(ctx context.Context, req *paymentpb.ProcessPaymentRequest) (*paymentpb.PaymentsResponse, error) {
	res, err := h.uc.Process(ctx, domain.Payment{
		OrderID:       req.GetOrderId(),
		PaymentMethod: req.GetPaymentMethod(),
		PaymentAmount: req.GetPaymentAmount(),
	})
	if err != nil {
		return nil, errgrpc.ReadInternalError(err)
	}

	return &paymentpb.PaymentsResponse{
		PaymentId:     res.ID,
		OrderId:       res.OrderID,
		PaymentStatus: res.PaymentStatus,
		PaymentMethod: res.PaymentMethod,
		PaymentAmount: res.PaymentAmount,
		Message:       "Registrasi transaksi Virtual Account berhasil diproses",
	}, nil

}

func (h *gRPCHandlerImpl) HandleWebhook(ctx context.Context, req *paymentpb.HandleWebhookRequest) (*paymentpb.HandleWebhookResponse, error) {
	err := h.uc.HandleWebhook(ctx, map[string]interface{}{
		"order_id":           req.GetOrderId(),
		"transaction_status": req.GetTransactionStatus(),
		"fraud_status":       req.GetFraudStatus(),
	})
	if err != nil {
		return nil, errgrpc.ReadInternalError(err)
	}

	return &paymentpb.HandleWebhookResponse{
		Success: true,
		Message: "Webhook sukses diproses via gRPC",
	}, nil
}
