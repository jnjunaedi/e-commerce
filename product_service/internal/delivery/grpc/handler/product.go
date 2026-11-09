package handler

import (
	"context"
	productpb "e-commerce/pb/product"
	"e-commerce/pkg/utils/errgrpc"
	"e-commerce/product_service/internal/domain"
	"errors"

	"google.golang.org/protobuf/types/known/timestamppb"
)

type DeductStockJob struct {
	Ctx        context.Context
	Req        *productpb.UpdateStockRequest
	ResultChan chan DeductStockResult
}
type DeductStockResult struct {
	Err  error
	Resp *productpb.UpdateStockResponse
}

type gRPCHandlerImpl struct {
	productpb.UnimplementedProductServiceServer
	jobQueue chan DeductStockJob
	uc       domain.Usecase
}

func NewGRPCHandler(uc domain.Usecase, maxWorker int, maxQueue int) productpb.ProductServiceServer {
	handler := &gRPCHandlerImpl{uc: uc, jobQueue: make(chan DeductStockJob, maxQueue)}

	for i := 0; i < maxWorker; i++ {
		go handler.startWorker()
	}

	return handler
}

func (h *gRPCHandlerImpl) startWorker() {
	for job := range h.jobQueue {
		success, msg, err := h.uc.DeductFlashSaleStock(job.Ctx, job.Req.GetProductId(), job.Req.GetFlashSaleId(), job.Req.GetQuantity())

		var currentStock int32
		if err == nil && success {
			currentStock = 0
		}
		job.ResultChan <- DeductStockResult{
			Err:  err,
			Resp: &productpb.UpdateStockResponse{Success: success, Message: msg, CurrentStock: currentStock},
		}

	}
}

func (h *gRPCHandlerImpl) CreateProduct(ctx context.Context, req *productpb.CreateProductRequest) (*productpb.ProductResponse, error) {
	res, err := h.uc.RegisterProduct(ctx, domain.Product{ProductCode: req.GetProductCode(), CategoryID: req.GetCategoryId(), ProductName: req.GetProductName(), Price: req.GetPrice(), Description: req.GetDescription(), ImageURLs: req.GetImageUrls(), Stock: req.GetInitialStock()})
	if err != nil {
		return nil, errgrpc.ReadInternalError(err)
	}

	return &productpb.ProductResponse{ProductId: res.ID, ProductCode: res.ProductCode, ProductName: res.ProductName, Description: res.Description, Price: res.Price, Stock: res.Stock, ImageUrls: res.ImageURLs, CategoryId: res.CategoryID}, nil
}

func (h *gRPCHandlerImpl) GetProduct(ctx context.Context, req *productpb.GetProductRequest) (*productpb.ProductResponse, error) {
	res, err := h.uc.FindProduct(ctx, req.GetProductId())
	if err != nil {
		return nil, errgrpc.ReadInternalError(err)
	}
	return &productpb.ProductResponse{ProductId: res.ID, ProductCode: res.ProductCode, ProductName: res.ProductName, Description: res.Description, Price: res.Price, Stock: res.Stock, ImageUrls: res.ImageURLs, CategoryId: res.CategoryID, CreatedAt: timestamppb.New(res.CreatedAt), UpdatedAt: timestamppb.New(res.UpdatedAt)}, nil
}

func (h *gRPCHandlerImpl) CreateFlashSaleProduct(ctx context.Context, req *productpb.CreateFlashSaleRequest) (*productpb.FlashSaleResponse, error) {

	res, err := h.uc.StartFlashSaleEvent(ctx, domain.ProductFlashSale{ProductID: req.GetProductId(), FlashSalePrice: req.GetFlashSalePrice(), FlashSaleStock: req.GetFlashSaleStock(), MaxPurchasePerUser: req.GetMaxPurchasePerUser(), StartTime: req.GetStartTime(), EndTime: req.GetEndTime()})
	if err != nil {
		return nil, errgrpc.ReadInternalError(err)
	}

	return &productpb.FlashSaleResponse{FlashSaleId: res.ID, ProductId: res.ProductID, FlashSalePrice: res.FlashSalePrice, FlashSaleStock: res.FlashSaleStock, MaxPurchasePerUser: res.MaxPurchasePerUser, StartTime: res.StartTime, EndTime: res.EndTime, Version: res.Version}, nil
}

func (h *gRPCHandlerImpl) GetFlashSaleProduct(ctx context.Context, req *productpb.GetFlashSaleRequest) (*productpb.FlashSaleResponse, error) {
	res, err := h.uc.FindFlashSale(ctx, req.GetFlashSaleId())
	if err != nil {
		return nil, errgrpc.ReadInternalError(err)
	}
	return &productpb.FlashSaleResponse{FlashSaleId: res.ID, ProductId: res.ProductID, FlashSalePrice: res.FlashSalePrice, FlashSaleStock: res.FlashSaleStock, MaxPurchasePerUser: res.MaxPurchasePerUser, StartTime: res.StartTime, EndTime: res.EndTime, Version: res.Version}, nil
}

func (h *gRPCHandlerImpl) UpdateStock(ctx context.Context, req *productpb.UpdateStockRequest) (*productpb.UpdateStockResponse, error) {
	if req.GetFlashSaleId() != "" {
		resultChan := make(chan DeductStockResult, 1)

		job := DeductStockJob{
			Ctx:        ctx,
			Req:        req,
			ResultChan: resultChan,
		}

		select {
		case h.jobQueue <- job:
			select {
			case res := <-resultChan:
				return res.Resp, res.Err
			case <-ctx.Done():
				return nil, errors.New("request timeout saat diproses worker")
			}
		default:
			return &productpb.UpdateStockResponse{
				Success: false,
				Message: "Server sedang sibuk, silakan coba beberapa saat lagi",
			}, nil
		}
	} else {
		err := h.uc.DeductRegularStock(ctx, req.GetProductId(), req.GetQuantity())
		if err != nil {
			return nil, errgrpc.ReadInternalError(err)
		}
		return &productpb.UpdateStockResponse{
			Success: true,
			Message: "Update stok produk reguler berhasil",
		}, nil
	}

}

func (h *gRPCHandlerImpl) MonitorStock(req *productpb.MonitorStockRequest, stream productpb.ProductService_MonitorStockServer) error {
	ctx := stream.Context()
	fsID := req.GetFlashSaleId()

	stockChan, err := h.uc.StreamStock(ctx, fsID)
	if err != nil {
		return errgrpc.ReadInternalError(err)
	}

	fsInit, errInit := h.uc.FindFlashSale(ctx, fsID)
	if errInit == nil {
		_ = stream.Send(&productpb.MonitorStockResponse{FlashSaleId: fsID, CurrentStock: fsInit.FlashSaleStock})
	}

	for {
		select {
		case latestStock, ok := <-stockChan:
			if !ok {
				return nil
			}

			err := stream.Send(&productpb.MonitorStockResponse{FlashSaleId: fsID, CurrentStock: latestStock})

			if err != nil {
				return err
			}
		case <-ctx.Done():
			return ctx.Err()
		}

	}
}
