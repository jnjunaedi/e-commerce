package worker

import (
	"context"
	"database/sql"
	"e-commerce/auth_service/infrastructure/client"
	"encoding/json"
	"log"
	"time"
)

type OutboxWorker struct {
	db         *sql.DB
	grpcClient client.UserGrpcClient
}

func NewOutboxWorker(db *sql.DB, grpcClient client.UserGrpcClient) *OutboxWorker {
	return &OutboxWorker{db: db, grpcClient: grpcClient}
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
		if event.EventType == "UserRegistered" {

			var payloadMap map[string]string
			_ = json.Unmarshal([]byte(event.Payload), &payloadMap)

			success := w.sendWithRetry(ctx, payloadMap["user_id"], payloadMap["username"])

			if success {
				_, _ = w.db.ExecContext(ctx, `UPDATE outbox SET status = 'PROCESSED' WHERE id = ?`, event.ID)
				log.Printf("[Outbox Worker] Sukses memproses Profil untuk User ID: %s", payloadMap["user_id"])
			}

		}
	}
}

func (w *OutboxWorker) sendWithRetry(ctx context.Context, userID, username string) bool {
	maxRetries := 3
	backoff := 1 * time.Second

	for i := 1; i <= maxRetries; i++ {
		_, err := w.grpcClient.CreateUserProfile(ctx, userID, username)
		if err == nil {
			return true
		}

		log.Printf("[Outbox Worker] Gagal kirim gRPC (Percobaan %d): %v. Mencoba lagi dalam %v...", i, err, backoff)
		time.Sleep(backoff)
		backoff *= 2
	}

	return false
}
