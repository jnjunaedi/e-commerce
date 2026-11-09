package client

import (
	productpb "e-commerce/pb/product"
	"log"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func InitProductClient(serverAddress string) productpb.ProductServiceClient {
	conn, err := grpc.NewClient(serverAddress, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("API Gateway gagal tersambung ke Product Service gRPC: %v", err)
	}

	return productpb.NewProductServiceClient(conn)
}
