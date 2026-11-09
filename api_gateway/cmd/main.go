package main

import (
	"context"
	"e-commerce/api_gateway/internal/client"
	"e-commerce/api_gateway/internal/delivery/http/handler"
	authMiddlware "e-commerce/api_gateway/internal/delivery/http/middleware"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Fatalf("Gagal memuat file .env : %v\n", err)
	}

	productAddr := getEnv("PRODUCT_SERVICE_ADDR", "localhost:50051")
	orderAddr := getEnv("ORDER_SERVICE_ADDR", "localhost:50052")
	payAddr := getEnv("PAYMENT_SERVICE_ADDR", "localhost:50053")
	authAddr := getEnv("AUTH_SERVICE_ADDR", "localhost:50055")
	userAddr := getEnv("USER_SERVICE_ADDR", "localhost:50056")

	productClient := client.InitProductClient(productAddr)
	productHandler := handler.NewProductHandler(productClient)

	orderClient := client.InitOrderClient(orderAddr)
	orderHandler := handler.NewOrderHandler(orderClient)

	payClient := client.InitPaymentClient(payAddr)
	payHandler := handler.NewPaymentHandler(payClient)

	authClient := client.InitAuthClient(authAddr)
	authHandler := handler.NewAuthHandler(authClient)

	userClient := client.InitUserClient(userAddr)
	userHandler := handler.NewUserHandler(userClient)

	r := chi.NewRouter()

	id, _ := uuid.NewV7()
	log.Println(id.String())

	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(60 * time.Second))

	authMiddlware := authMiddlware.NewAuthMiddleware(authClient)

	r.Route("/api/v1", func(r chi.Router) {
		r.Post("/auth/register", authHandler.RegisterHandler)
		r.Post("/auth/login", authHandler.LoginHandler)
		r.Get("/products/{id}", productHandler.FindProductHandler)
		r.Get("/products/flash-sale/{id}", productHandler.FindFlashSale)
		r.Get("/products/flash-sale/{id}/stream", productHandler.MonitorStockSSEHandler)
		r.Post("/transaction/payment/webhook", payHandler.HandleMidtransWebhookHandler)

		r.Group(func(r chi.Router) {
			r.Use(authMiddlware.RequireAuth)

			r.Put("/users/profiles", userHandler.UpdateUserProfileHandler)
			r.Put("/users/addresses", userHandler.UpdateUserAddressHandler)
			r.Post("/products", productHandler.CreateProductHandler)
			r.Post("/products/flash-sales", productHandler.StartFlashSaleEventHandler)
			r.Post("/orders/checkout", orderHandler.CreateOrderHandler)
			r.Post("/transaction/payments", payHandler.ProcessPaymentHandler)
		})
	})

	addr := os.Getenv("API_PORT")
	if addr == "" {
		addr = ":8080"
	}

	srv := &http.Server{
		Addr:    addr,
		Handler: r,
	}
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Gagal menyalakan API Gateway: %v", err)
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	<-quit

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("Eror saat melakukan shutdown API Gateway: %v\n", err)
	}
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
