package worker

import (
	"context"
	"database/sql"
	"e-commerce/payment_service/infrastructure/client"
	"encoding/json"
	"log"
	"time"
)

type OutboxWorker struct {
	db          *sql.DB
	orderClient client.OrderGrpcClient
}

func NewOutboxWorker(db *sql.DB, orderClient client.OrderGrpcClient) *OutboxWorker {
	return &OutboxWorker{db: db, orderClient: orderClient}
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
		if event.EventType == "UpdateOrderStatusCancelled" || event.EventType == "UpdateOrderStatusPaid" {

			var payloadMap map[string]string
			_ = json.Unmarshal([]byte(event.Payload), &payloadMap)

			success := w.sendWithRetry(ctx, payloadMap["order_id"], payloadMap["status"])

			if success {
				_, _ = w.db.ExecContext(ctx, `UPDATE outbox SET status = 'PROCESSED' WHERE id = ?`, event.ID)
				log.Printf("[Outbox Worker] Sukses memproses Perbaruan status menjadi %s untuk Order ID: %s", payloadMap["status"], payloadMap["order_id"])
			}

		}
	}
}

func (w *OutboxWorker) sendWithRetry(ctx context.Context, orderID, status string) bool {
	maxRetries := 3
	backoff := 1 * time.Second

	for i := 1; i <= maxRetries; i++ {
		err := w.orderClient.UpdateOrderStatus(ctx, orderID, status)
		if err == nil {
			return true
		}

		log.Printf("[Outbox Worker] Gagal kirim gRPC (Percobaan %d): %v. Mencoba lagi dalam %v...", i, err, backoff)
		time.Sleep(backoff)
		backoff *= 2
	}

	return false
}
