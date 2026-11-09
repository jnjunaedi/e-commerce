package client

import (
	authpb "e-commerce/pb/auth"
	"log"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func InitAuthClient(serverAddress string) authpb.AuthServiceClient {
	conn, err := grpc.NewClient(serverAddress, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("API Gateway gagal tersambung ke Auth Service gRPC: %v", err)
	}

	return authpb.NewAuthServiceClient(conn)
}
