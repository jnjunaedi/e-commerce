package main

import (
	userpb "e-commerce/pb/user"
	"e-commerce/pkg/utils"
	"e-commerce/user_service/config"
	"e-commerce/user_service/internal/delivery/grpc/handler"
	"e-commerce/user_service/internal/repository"
	"e-commerce/user_service/internal/usecase"
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

	idGenerator := utils.NewIDGenerator()

	userRepo := repository.NewRepository(db)
	userUc := usecase.NewUsecase(userRepo, idGenerator)
	userGRPCHandler := handler.NewGRPCHandler(userUc)

	lis, err := net.Listen("tcp", ":50056")
	if err != nil {
		log.Fatalf("Gagal membuka port 50056: %v", err)
	}

	server := grpc.NewServer()
	userpb.RegisterUserServiceServer(server, userGRPCHandler)

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
