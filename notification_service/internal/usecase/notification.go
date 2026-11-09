package usecase

import (
	"context"
	"e-commerce/notification_service/infrastructure/email"
	"e-commerce/notification_service/internal/domain"
)

type UsecaseImpl struct {
	repo        domain.Repository
	emailSender email.EmailSender
}

func NewUsecase(repo domain.Repository, emailSender email.EmailSender) domain.Usecase {
	return &UsecaseImpl{repo: repo, emailSender: emailSender}
}

func (uc *UsecaseImpl) SendAndSaveNotification(ctx context.Context, n domain.Notification) error {
	err := uc.emailSender.SendEmail(n.Email, n.Title, n.Message)
	if err != nil {
		n.Status = "FAILED"
	} else {
		n.Status = "SENT"
	}

	if saveErr := uc.repo.Create(ctx, n); saveErr != nil {
		return saveErr
	}
	return nil
}
