package client

import (
	paymentpb "e-commerce/pb/payment"
	"log"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func InitPaymentClient(serverAddress string) paymentpb.PaymentsServiceClient {
	conn, err := grpc.NewClient(serverAddress, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("API Gateway gagal tersambung ke Payment Service gRPC: %v", err)
	}

	return paymentpb.NewPaymentsServiceClient(conn)
}
