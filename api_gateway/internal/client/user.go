package client

import (
	userpb "e-commerce/pb/user"
	"log"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func InitUserClient(serverAddress string) userpb.UserServiceClient {
	conn, err := grpc.NewClient(serverAddress, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("API Gateway gagal tersambung ke User Service gRPC: %v", err)
	}

	return userpb.NewUserServiceClient(conn)
}
