package handler

import (
	"context"
	"e-commerce/notification_service/internal/domain"
	notificationpb "e-commerce/pb/notification"
	"e-commerce/pkg/utils/errgrpc"
)

type gRPCHandlerImpl struct {
	notificationpb.UnimplementedNotificationsServiceServer
	uc domain.Usecase
}

func NewNotificationGRPCHandler(uc domain.Usecase) notificationpb.NotificationsServiceServer {
	return &gRPCHandlerImpl{uc: uc}
}

func (h *gRPCHandlerImpl) SendNotification(ctx context.Context, req *notificationpb.SendNotificationRequest) (*notificationpb.NotificationsResponse, error) {
	notification := domain.Notification{
		UserID:           req.GetUserId(),
		OrderID:          req.GetOrderId(),
		NotificationType: req.GetNotificationType(),
		Title:            req.GetTitle(),
		Message:          req.GetMessage(),
	}
	err := h.uc.SendAndSaveNotification(ctx, notification)
	if err != nil {
		return nil, errgrpc.ReadInternalError(err)
	}

	return &notificationpb.NotificationsResponse{Success: true, Message: "Notifikasi berhasil terkirim"}, nil
}
