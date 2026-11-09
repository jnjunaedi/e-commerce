package main

import (
	"context"
	"e-commerce/auth_service/config"
	"e-commerce/auth_service/infrastructure/client"
	"e-commerce/auth_service/internal/delivery/grpc/handler"
	"e-commerce/auth_service/internal/repository"
	"e-commerce/auth_service/internal/repository/worker"
	"e-commerce/auth_service/internal/usecase"
	"e-commerce/auth_service/pkg/security"
	authpb "e-commerce/pb/auth"
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

	userClient := client.InitUserClient("localhost:50056")
	userBridge := client.NewUserGrpcClient(userClient)

	password := security.NewPassword()
	jwtProvider := security.NewJwtProvider()
	idGenerator := utils.NewIDGenerator()

	authRepo := repository.NewRepository(db)
	txManager := repository.NewSQLTxManager(db)
	authUc := usecase.NewUsecase(authRepo, txManager, password, jwtProvider, idGenerator)
	authGRPCHandler := handler.NewGRPCHandler(authUc)

	outboxWorker := worker.NewOutboxWorker(db, userBridge)

	ctx, cancelWorker := context.WithCancel(context.Background())
	defer cancelWorker()

	go outboxWorker.Start(ctx)

	lis, err := net.Listen("tcp", ":50055")
	if err != nil {
		log.Fatalf("Gagal membuka port 50055: %v", err)
	}

	server := grpc.NewServer()
	authpb.RegisterAuthServiceServer(server, authGRPCHandler)

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
