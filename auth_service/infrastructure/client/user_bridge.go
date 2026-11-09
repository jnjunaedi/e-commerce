package client

import (
	"context"
	userpb "e-commerce/pb/user"
)

type CreateUserProfileResponse struct {
	Message string
	Success bool
}

type UserGrpcClient interface {
	CreateUserProfile(ctx context.Context, userID, username string) (CreateUserProfileResponse, error)
}

type userGrpcClient struct {
	client userpb.UserServiceClient
}

func NewUserGrpcClient(client userpb.UserServiceClient) UserGrpcClient {
	return &userGrpcClient{client: client}
}

func (c *userGrpcClient) CreateUserProfile(ctx context.Context, userID string, username string) (CreateUserProfileResponse, error) {
	res, err := c.client.CreateUserProfile(ctx, &userpb.CreateUserProfileRequest{UserId: userID, Username: username})
	if err != nil {
		return CreateUserProfileResponse{}, err
	}
	return CreateUserProfileResponse{Message: res.GetMessage(), Success: res.GetSuccess()}, nil
}
