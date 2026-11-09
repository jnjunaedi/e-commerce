package client

import (
	orderpb "e-commerce/pb/order"
	"log"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func InitOrderClient(serverAddress string) orderpb.OrdersServiceClient {
	conn, err := grpc.NewClient(serverAddress, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("API Gateway gagal tersambung ke Order Service gRPC: %v", err)
	}

	return orderpb.NewOrdersServiceClient(conn)
}
