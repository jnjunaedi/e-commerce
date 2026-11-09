package main

import (
	"context"
	"e-commerce/order_service/config"
	"e-commerce/order_service/infrastructure/client"
	"e-commerce/order_service/infrastructure/kafka"
	"e-commerce/order_service/internal/delivery/grpc/handler"
	"e-commerce/order_service/internal/repository"
	"e-commerce/order_service/internal/repository/worker"
	"e-commerce/order_service/internal/usecase"
	orderpb "e-commerce/pb/order"
	"e-commerce/pkg/utils"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/joho/godotenv"
	"google.golang.org/grpc"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Fatalf("Gagal memuat file .env : %v\n", err)
	}

	db, err := config.ConnectDB()
	if err != nil {
		log.Fatalf("Gagal terhubung ke MySQL : %v\n", err)
	}
	defer db.Close()

	productClient := client.InitProductClient("localhost:50051")
	productBridge := client.NewProductClient(productClient)

	kafkaBrokers := []string{"localhost:9092"}
	kafkaProducer, err := kafka.NewOrderProducer(kafkaBrokers)
	if err != nil {
		log.Fatalf("Gagal menginisialisasi Kafka Producer: %v", err)
	}

	idGenerator := utils.NewIDGenerator()

	orderRepo := repository.NewRepository(db, idGenerator)
	txManager := repository.NewSQLTxManager(db)
	orderUc := usecase.NewUsecase(orderRepo, productBridge, txManager, kafkaProducer, idGenerator)
	orderGRPCHandler := handler.NewGRPCHandler(orderUc)

	outboxWorker := worker.NewOutboxWorker(db, productBridge, idGenerator, kafkaProducer)

	ctx, cancelWorker := context.WithCancel(context.Background())
	defer cancelWorker()

	go outboxWorker.Start(ctx)

	lis, err := net.Listen("tcp", ":50052")
	if err != nil {
		log.Fatalf("Gagal membuka port 50052: %v", err)
	}

	server := grpc.NewServer()
	orderpb.RegisterOrdersServiceServer(server, orderGRPCHandler)

	go func() {
		if err := server.Serve(lis); err != nil && grpc.ErrServerStopped != nil {
			log.Fatalf("Gagal menjalankan server gRPC: %v", err)
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	<-quit

	server.GracefulStop()
}
