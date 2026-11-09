package usecase_test

import (
	"context"
	mEmail "e-commerce/notification_service/infrastructure/email/mocks"
	"e-commerce/notification_service/internal/domain"
	mRepo "e-commerce/notification_service/internal/repository/mocks"
	"e-commerce/notification_service/internal/usecase"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestNotificationUsecase_SendAndSaveNotification(t *testing.T) {
	mockRepo := mRepo.NewRepository(t)
	mockEmail := mEmail.NewEmailSender(t)

	inputNotification := domain.Notification{
		ID:               "notif-uuid-112",
		UserID:           "USR-999",
		Email:            "customer@gmail.com",
		OrderID:          "ODR-777",
		NotificationType: "EMAIL",
		Title:            "Pesanan Berhasil",
		Message:          "Halo, pesanan Anda berhasil diproses.",
		Status:           "PENDING",
		CreatedAt:        time.Now(),
	}

	mockRepo.On("Create", mock.Anything, mock.MatchedBy(func(n domain.Notification) bool {
		return n.ID == "notif-uuid-112" && n.Status == "SENT"
	})).Return(nil).Once()

	mockEmail.On("SendEmail", "customer@gmail.com", "Pesanan Berhasil", "Halo, pesanan Anda berhasil diproses.").
		Return(nil).Once()

	uc := usecase.NewUsecase(mockRepo, mockEmail)
	ctx := context.Background()

	err := uc.SendAndSaveNotification(ctx, inputNotification)

	assert.NoError(t, err)

	mockRepo.AssertExpectations(t)
	mockEmail.AssertExpectations(t)

}
