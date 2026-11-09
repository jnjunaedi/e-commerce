package client

import (
	"context"
	productpb "e-commerce/pb/product"
	"errors"
)

type FlashSaleProductResult struct {
	FlashSaleID        string
	ProductID          string
	FlashSalePrice     int64
	FlashSaleStock     int32
	MaxPurchasePerUser int32
}

type RegularProductResult struct {
	ID          string
	ProductCode string
	ProductName string
	CategoryID  string
	Price       int64
	Stock       int32
	Description string
	ImageURLs   []string
}

type ProductGrpcClient interface {
	FindFlashSale(ctx context.Context, fsID string) (FlashSaleProductResult, error)
	DeductFlashSaleStock(ctx context.Context, productID, flashSaleID string, qty int32) error
	FindRegularProduct(ctx context.Context, prodRegulerID string) (RegularProductResult, error)
	DeductRegularStock(ctx context.Context, productID string, qty int32) error
}

type productClient struct {
	productGRPC productpb.ProductServiceClient
}

func NewProductClient(productGRPC productpb.ProductServiceClient) ProductGrpcClient {
	return &productClient{productGRPC: productGRPC}
}

func (c *productClient) FindRegularProduct(ctx context.Context, id string) (RegularProductResult, error) {
	res, err := c.productGRPC.GetProduct(ctx, &productpb.GetProductRequest{ProductId: id})
	if err != nil {
		return RegularProductResult{}, err
	}

	return RegularProductResult{ID: res.GetProductId(), ProductCode: res.GetProductCode(), ProductName: res.GetProductName(), CategoryID: res.GetCategoryId(), Price: res.GetPrice(), Stock: res.GetStock(), Description: res.GetDescription(), ImageURLs: res.GetImageUrls()}, nil
}

func (c *productClient) DeductRegularStock(ctx context.Context, productID string, qty int32) error {
	_, err := c.productGRPC.UpdateStock(ctx, &productpb.UpdateStockRequest{ProductId: productID, Quantity: qty})
	if err != nil {
		return err
	}
	return nil
}

func (c *productClient) FindFlashSale(ctx context.Context, id string) (FlashSaleProductResult, error) {
	res, err := c.productGRPC.GetFlashSaleProduct(ctx, &productpb.GetFlashSaleRequest{FlashSaleId: id})
	if err != nil {
		return FlashSaleProductResult{}, err
	}

	return FlashSaleProductResult{FlashSaleID: res.GetFlashSaleId(), ProductID: res.GetProductId(), FlashSalePrice: res.GetFlashSalePrice(), FlashSaleStock: res.GetFlashSaleStock(), MaxPurchasePerUser: res.GetMaxPurchasePerUser()}, nil
}

func (c *productClient) DeductFlashSaleStock(ctx context.Context, productID string, flashSaleID string, qty int32) error {
	res, err := c.productGRPC.UpdateStock(ctx, &productpb.UpdateStockRequest{ProductId: productID, Quantity: qty, FlashSaleId: flashSaleID})
	if err != nil {
		return err
	}

	if !res.GetSuccess() {
		return errors.New(res.GetMessage())
	}

	return nil
}
