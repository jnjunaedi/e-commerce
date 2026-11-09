package main

import (
	productpb "e-commerce/pb/product"
	"e-commerce/pkg/utils"
	"e-commerce/product_service/config"
	"e-commerce/product_service/internal/delivery/grpc/handler"
	"e-commerce/product_service/internal/repository"
	"e-commerce/product_service/internal/usecase"
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

	productRepo := repository.NewRepository(db, idGenerator)
	productUc := usecase.NewUsecase(productRepo, idGenerator)
	productGRPCHandler := handler.NewGRPCHandler(productUc, 25, 20)

	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("Gagal membuka port 50051: %v", err)
	}

	server := grpc.NewServer()
	productpb.RegisterProductServiceServer(server, productGRPCHandler)

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
