package main

import (
	"context"
	"e-commerce/payment_service/config"
	"e-commerce/payment_service/infrastructure/client"
	"e-commerce/payment_service/internal/delivery/grpc/handler"
	"e-commerce/payment_service/internal/repository"
	"e-commerce/payment_service/internal/repository/worker"
	"e-commerce/payment_service/internal/usecase"
	paymentpb "e-commerce/pb/payment"
	"e-commerce/pkg/utils"
	"log"
	"net"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/joho/godotenv"
	"github.com/midtrans/midtrans-go"
	"github.com/midtrans/midtrans-go/coreapi"
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

	orderClient := client.InitOrderClient("localhost:50052")
	orderBridge := client.NewOrderClient(orderClient)

	serverKey := strings.TrimSpace(os.Getenv("MIDTRANS_SERVER_KEY"))
	if serverKey == "" {
		log.Fatal("MIDTRANS_SERVER_KEY belum di-set")
	}

	midtrans.ServerKey = serverKey
	midtrans.Environment = midtrans.Sandbox

	myNgrokURL := "https://vowed-ongoing-dawn.ngrok-free.dev"

	c := coreapi.Client{}
	c.New(serverKey, midtrans.Sandbox)
	c.Options.SetPaymentOverrideNotification(
		myNgrokURL + "/api/v1/transaction/payment/webhook",
	)

	idGenerator := utils.NewIDGenerator()

	payRepo := repository.NewRepository(db)
	txManager := repository.NewSQLTxManager(db)
	payUc := usecase.NewUsecase(payRepo, orderBridge, txManager, c, idGenerator)
	payGRPCHandler := handler.NewGRPCHandler(payUc)

	outboxWorker := worker.NewOutboxWorker(db, orderBridge)

	ctx, outboxCancel := context.WithCancel(context.Background())
	defer outboxCancel()

	go outboxWorker.Start(ctx)

	lis, err := net.Listen("tcp", ":50053")
	if err != nil {
		log.Fatalf("Gagal membuka port 50053: %v", err)
	}

	server := grpc.NewServer()
	paymentpb.RegisterPaymentsServiceServer(server, payGRPCHandler)

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
