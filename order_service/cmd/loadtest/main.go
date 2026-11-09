package main

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	orderpb "e-commerce/pb/order"
)

func main() {
	// 1. Hubungkan gRPC client ke Order Service (sesuaikan portnya, misal 50052)
	conn, err := grpc.NewClient("localhost:50052", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Gagal terhubung ke Order Service: %v", err)
	}
	defer conn.Close()

	client := orderpb.NewOrdersServiceClient(conn)

	totalRequests := 500
	concurrency := 50
	semaphore := make(chan struct{}, concurrency)
	var wg sync.WaitGroup

	fmt.Println("🚀 Memulai simulasi load test CreateOrder...")
	start := time.Now()

	for i := 0; i < totalRequests; i++ {
		wg.Add(1)
		semaphore <- struct{}{} // Ambil slot concurrency

		go func(idx int) {
			defer wg.Done()
			defer func() { <-semaphore }() // Lepas slot setelah selesai

			// Buat data unik agar tidak bentrok dengan idempotency key
			userID := uuid.New().String()
			idempotencyKey := fmt.Sprintf("loadtest-%d-%s", idx, uuid.New().String())

			// 2. INI YANG DIPANGGIL: gRPC CreateOrder
			req := &orderpb.CreateOrderRequest{
				UserId:         userID,
				IdempotencyKey: idempotencyKey,
				Items: []*orderpb.OrderItemRequest{
					{
						ProductId:   "01a17d38-d4cf-7bda-a3ed-54c829e37a17", // ID produk di DB Anda
						FlashSaleId: "01a17d79-e5d8-7224-b1e1-25ae504724bf", // Isi ID Flash Sale jika ingin tes worker pool
						Quantity:    1,
					},
				},
			}

			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			res, err := client.CreateOrder(ctx, req)
			if err != nil {
				fmt.Printf("❌ Request %d GAGAL: %v\n", idx, err)
				return
			}

			fmt.Printf("✅ Request %d SUKSES! Order ID: %s\n", idx, res.GetOrderId())
		}(i)
	}

	wg.Wait()
	fmt.Printf("🎉 Selesai! Semua %d request terkirim dalam waktu: %v\n", totalRequests, time.Since(start))
}
