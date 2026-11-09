package usecase

import (
	"context"
	"e-commerce/payment_service/infrastructure/client"
	"e-commerce/payment_service/internal/domain"
	"e-commerce/pkg/utils"
	"e-commerce/pkg/utils/apperror"
	"encoding/json"
	"fmt"
	"time"

	"github.com/midtrans/midtrans-go"
	"github.com/midtrans/midtrans-go/coreapi"
)

type UsecaseImpl struct {
	repo           domain.Repository
	txManager      domain.TxManager
	orderClient    client.OrderGrpcClient
	midtransClient coreapi.Client
	idGenerator    utils.IDGenerator
}

func NewUsecase(repo domain.Repository, orderClient client.OrderGrpcClient, txManager domain.TxManager, midtransClient coreapi.Client, idGenerator utils.IDGenerator) domain.Usecase {
	return &UsecaseImpl{repo: repo, orderClient: orderClient, txManager: txManager, midtransClient: midtransClient, idGenerator: idGenerator}
}

func (u *UsecaseImpl) Process(ctx context.Context, p domain.Payment) (domain.Payment, error) {
	orderData, err := u.orderClient.GetOrderDetails(ctx, p.OrderID)
	if err != nil {
		return domain.Payment{}, err
	}

	if p.PaymentAmount < orderData.TotalPrice {
		return domain.Payment{}, apperror.New(apperror.ErrBadRequest, fmt.Sprintf("Nominal kurang, total tagihan asli: %d", orderData.TotalPrice))
	}

	paymentUUID, err := u.idGenerator.GenerateUUID()
	if err != nil {
		return domain.Payment{}, apperror.New(apperror.ErrInternalServer, "error internal")
	}
	p.ID = paymentUUID
	p.PaymentStatus = "PENDING"

	chargeReq := &coreapi.ChargeReq{
		PaymentType: coreapi.PaymentTypeBankTransfer,
		TransactionDetails: midtrans.TransactionDetails{
			OrderID:  p.OrderID,
			GrossAmt: p.PaymentAmount,
		},
		BankTransfer: &coreapi.BankTransferDetails{
			Bank: midtrans.BankBca,
		},
	}

	midtransRes, errMidtrans := u.midtransClient.ChargeTransaction(chargeReq)
	if errMidtrans != nil {
		return domain.Payment{}, apperror.New(apperror.ErrInternalServer, "Gagal berkomunikasi dengan Midtrans Gateway")
	}

	var virtualAccount string
	if len(midtransRes.VaNumbers) > 0 {
		virtualAccount = midtransRes.VaNumbers[0].VANumber
	}

	err = u.repo.Create(ctx, p)
	if err != nil {
		return domain.Payment{}, err
	}

	p.PaymentMethod = "BCA Virtual Account (" + virtualAccount + ")"

	return p, nil

}

func (u *UsecaseImpl) HandleWebhook(ctx context.Context, payload map[string]interface{}) error {
	orderID, okOrder := payload["order_id"].(string)

	if !okOrder || orderID == "" {
		return apperror.New(apperror.ErrBadRequest, "Payload notifikasi Midtrans tidak valid")
	}

	midtransStatusResp, errCheck := u.midtransClient.CheckTransaction(orderID)
	if errCheck != nil {
		return apperror.New(apperror.ErrInternalServer, "Gagal memverifikasi keaslian transaksi ke Midtrans Gateway")
	}

	txStatus := midtransStatusResp.TransactionStatus
	fraudStatus := midtransStatusResp.FraudStatus

	err := u.txManager.WithTransaction(ctx, func(ctx context.Context) error {
		if txStatus == "settlement" || txStatus == "capture" {
			if txStatus == "capture" && fraudStatus != "accept" {
				err := u.repo.UpdateStatus(ctx, orderID, "FAILED", 0)
				if err != nil {
					return err
				}

				payloadMap := map[string]string{
					"order_id": orderID,
					"status":   "CANCELLED",
				}

				payloadBytes, err := json.Marshal(payloadMap)
				if err != nil {
					return err
				}

				outboxID, err := u.idGenerator.GenerateUUID()
				if err != nil {
					return err
				}

				err = u.repo.CreateOutbox(ctx, outboxID, "UpdateOrderStatusCancelled", payloadBytes)
				if err != nil {
					return err
				}

				return apperror.New(apperror.ErrBadRequest, "Transaksi terindikasi fraud / ditolak sistem keamanan")
			}

			var paidAtUnix int64 = time.Now().Unix()

			err := u.repo.UpdateStatus(ctx, orderID, "SUCCESS", paidAtUnix)
			if err != nil {
				return err
			}

			payloadMap := map[string]string{
				"order_id": orderID,
				"status":   "PAID",
			}

			payloadBytes, err := json.Marshal(payloadMap)
			if err != nil {
				return err
			}

			outboxID, err := u.idGenerator.GenerateUUID()
			if err != nil {
				return err
			}

			err = u.repo.CreateOutbox(ctx, outboxID, "UpdateOrderStatusPaid", payloadBytes)
			if err != nil {
				return err
			}
		} else if txStatus == "deny" || txStatus == "expire" || txStatus == "cancel" {
			err := u.repo.UpdateStatus(ctx, orderID, "FAILED", 0)
			if err != nil {
				return err
			}
			payloadMap := map[string]string{
				"order_id": orderID,
				"status":   "CANCELLED",
			}

			payloadBytes, err := json.Marshal(payloadMap)
			if err != nil {
				return err
			}

			outboxID, err := u.idGenerator.GenerateUUID()
			if err != nil {
				return err
			}

			err = u.repo.CreateOutbox(ctx, outboxID, "UpdateOrderStatusCancelled", payloadBytes)
			if err != nil {
				return err
			}
		}
		return nil
	})

	if err != nil {
		return err
	}

	return nil

}
