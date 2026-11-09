package main

import (
	"context"
	"e-commerce/notification_service/config"
	"e-commerce/notification_service/infrastructure/email"
	"e-commerce/notification_service/internal/delivery/grpc/handler"
	"e-commerce/notification_service/internal/delivery/kafka"
	"e-commerce/notification_service/internal/repository"
	"e-commerce/notification_service/internal/usecase"
	notificationpb "e-commerce/pb/notification"
	"log"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/joho/godotenv"
	"google.golang.org/grpc"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := godotenv.Load(); err != nil {
		log.Fatalf("Gagal memuat file .env : %v\n", err)
	}

	db, err := config.ConnectDB()
	if err != nil {
		log.Fatalf("Gagal terhubung ke MySQL : %v\n", err)
	}
	defer db.Close()

	lis, err := net.Listen("tcp", ":50054")
	if err != nil {
		log.Fatalf("Gagal membuka port 50054: %v", err)
	}

	emailSender := email.NewSMTPSender(
		"sandbox.smtp.mailtrap.io",
		os.Getenv("SMTP_PORT"),
		os.Getenv("SMTP_USERNAME"),
		os.Getenv("SMTP_PASSWORD"),
		"no-reply@ecommerce.com",
	)

	notifRepo := repository.NewRepository(db)
	notifUc := usecase.NewUsecase(notifRepo, emailSender)
	notifGRPCHandler := handler.NewNotificationGRPCHandler(notifUc)

	brokers := []string{"localhost:9092"}
	topics := []string{"notification-events"}
	groupID := "notification-service-group"

	consumer, err := kafka.NewConsumer(brokers, groupID, topics, notifUc)
	if err != nil {
		log.Fatalf("Failed to create kafka consumer: %v", err)
	}
	defer consumer.Close()

	errChan := make(chan error, 1)
	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		consumer.Start(ctx, errChan)
	}()

	go func() {
		for errC := range errChan {
			log.Printf("Error notif consumer : %v\n", errC)
		}
	}()

	server := grpc.NewServer()
	notificationpb.RegisterNotificationsServiceServer(server, notifGRPCHandler)

	go func() {
		if err := server.Serve(lis); err != nil && err != grpc.ErrServerStopped {
			cancel()
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	<-quit

	cancel()

	server.GracefulStop()

	wg.Wait()

}
