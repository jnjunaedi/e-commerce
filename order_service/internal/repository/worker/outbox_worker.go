package worker

import (
	"context"
	"database/sql"
	"e-commerce/order_service/infrastructure/client"
	"e-commerce/order_service/infrastructure/kafka"
	"e-commerce/pkg/utils"
	"encoding/json"
	"fmt"
	"log"
	"time"
)

type OutboxWorker struct {
	db            *sql.DB
	grpcClient    client.ProductGrpcClient
	idGenerator   utils.IDGenerator
	kafkaProducer kafka.OrderProducer
}

func NewOutboxWorker(db *sql.DB, grpcClient client.ProductGrpcClient, idGenerator utils.IDGenerator, kafkaProducer kafka.OrderProducer) *OutboxWorker {
	return &OutboxWorker{db: db, grpcClient: grpcClient, idGenerator: idGenerator, kafkaProducer: kafkaProducer}
}

func (w *OutboxWorker) Start(ctx context.Context) {
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			w.processPendingOutbox(ctx)
		}
	}
}

func (w *OutboxWorker) processPendingOutbox(ctx context.Context) {
	query := `SELECT id, event_type, payload FROM outbox WHERE status = 'PENDING' LIMIT 5`
	rows, err := w.db.QueryContext(ctx, query)
	if err != nil {
		return
	}
	defer rows.Close()

	type OutboxData struct {
		ID        string
		EventType string
		Payload   string
	}

	var pendingEvents []OutboxData
	for rows.Next() {
		var d OutboxData

		if err := rows.Scan(&d.ID, &d.EventType, &d.Payload); err == nil {
			pendingEvents = append(pendingEvents, d)
		}
	}

	for _, event := range pendingEvents {
		var payloadMap map[string]any
		_ = json.Unmarshal([]byte(event.Payload), &payloadMap)

		orderID := payloadMap["order_id"].(string)
		userID := payloadMap["user_id"].(string)
		productID := payloadMap["product_id"].(string)
		qty := payloadMap["quantity"].(float64)
		qtyResult := int32(qty)

		if event.EventType == "DeductFlashSaleStock" {

			flashSaleID := payloadMap["flash_sale_id"].(string)

			success := w.sendWithRetryDeductFlashSaleStock(ctx, productID, flashSaleID, qtyResult)

			if success {
				_, _ = w.db.ExecContext(ctx, `UPDATE orders SET order_status = 'PENDING_PAYMENT' WHERE id = ?`, orderID)
				_, _ = w.db.ExecContext(ctx, `UPDATE outbox SET status = 'PROCESSED' WHERE id = ?`, event.ID)
				log.Printf("[Worker] FlashSale Order %s SUKSES alokasi stok.", orderID)

				notifID, _ := w.idGenerator.GenerateUUID()
				notificationEvent := kafka.OrderCreatedEvent{
					ID:               notifID,
					UserID:           userID,
					Email:            "customer@example.com",
					OrderID:          orderID,
					NotificationType: "EMAIL",
					Title:            "Konfirmasi Pesanan Flash Sale Berhasil",
					Message:          fmt.Sprintf("Halo, pesanan flash sale Anda dengan ID %s berhasil dibuat dan siap dibayar.", orderID),
					Status:           "PENDING",
				}
				_ = w.kafkaProducer.PublishOrderCreated(ctx, notificationEvent)
			} else {
				_, _ = w.db.ExecContext(ctx, `UPDATE orders SET order_status = 'FAILED_STOCK_OUT' WHERE id = ?`, orderID)
				_, _ = w.db.ExecContext(ctx, `UPDATE outbox SET status = 'FAILED' WHERE id = ?`, event.ID)
				log.Printf("[Worker] FlashSale Order %s GAGAL (Stok Habis).", orderID)
			}
		} else if event.EventType == "DeductRegularStock" {
			success := w.sendWithRetryDeductRegularStock(ctx, productID, qtyResult)

			if success {
				_, _ = w.db.ExecContext(ctx, `UPDATE outbox SET status = 'PROCESSED' WHERE id = ?`, event.ID)
				log.Printf("[Worker] Regular Order %s Sukses potong stok di gRPC.", orderID)
				notifID, _ := w.idGenerator.GenerateUUID()
				notificationEvent := kafka.OrderCreatedEvent{
					ID:               notifID,
					UserID:           userID,
					Email:            "customer@example.com",
					OrderID:          orderID,
					NotificationType: "EMAIL",
					Title:            "Konfirmasi Pesanan Reguler Berhasil",
					Message:          fmt.Sprintf("Halo, pesanan reguler Anda dengan ID %s siap dibayar.", orderID),
					Status:           "PENDING",
				}
				_ = w.kafkaProducer.PublishOrderCreated(ctx, notificationEvent)
			} else {
				_, _ = w.db.ExecContext(ctx, `UPDATE outbox SET status = 'PENDING' WHERE id = ?`, event.ID)
			}
		}
	}

}

func (w *OutboxWorker) sendWithRetryDeductFlashSaleStock(ctx context.Context, productID, flashSaleID string, qty int32) bool {
	maxRetries := 3
	backoff := 1 * time.Second

	for i := 1; i <= maxRetries; i++ {
		err := w.grpcClient.DeductFlashSaleStock(ctx, productID, flashSaleID, qty)
		if err == nil {
			return true
		}

		if err.Error() == "stok produk flash sale sudah habis" {
			return false
		}

		log.Printf("[Outbox Worker] Gagal kirim gRPC (Percobaan %d): %v. Mencoba lagi dalam %v...", i, err, backoff)
		time.Sleep(backoff)
		backoff *= 2
	}

	return false
}

func (w *OutboxWorker) sendWithRetryDeductRegularStock(ctx context.Context, productID string, qty int32) bool {
	maxRetries := 3
	backoff := 1 * time.Second

	for i := 1; i <= maxRetries; i++ {
		err := w.grpcClient.DeductRegularStock(ctx, productID, qty)
		if err == nil {
			return true
		}

		log.Printf("[Outbox Worker] Gagal kirim gRPC (Percobaan %d): %v. Mencoba lagi dalam %v...", i, err, backoff)
		time.Sleep(backoff)
		backoff *= 2
	}

	return false
}
