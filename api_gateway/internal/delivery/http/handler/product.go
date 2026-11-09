package handler

import (
	"e-commerce/api_gateway/internal/delivery/helper"
	"e-commerce/api_gateway/internal/delivery/http/request"
	"e-commerce/api_gateway/internal/delivery/http/response"
	"e-commerce/api_gateway/pkg/validator"
	productpb "e-commerce/pb/product"
	"e-commerce/pkg/utils"
	"e-commerce/pkg/utils/apperror"
	"e-commerce/pkg/utils/errgrpc"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
)

type ProductHandlerImpl struct {
	client productpb.ProductServiceClient
}

func NewProductHandler(client productpb.ProductServiceClient) *ProductHandlerImpl {
	return &ProductHandlerImpl{client: client}
}

func (h *ProductHandlerImpl) CreateProductHandler(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 10<<20)

	ctx := r.Context()

	err := r.ParseMultipartForm(5 << 20)
	if err != nil {
		utils.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "Ukuran maksimal request 10MB"})
		return
	}

	jsonString := r.FormValue("data")
	if jsonString == "" {
		utils.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "Data JSON tidak boleh kosong"})
		return
	}

	var productReq request.ProductRequest
	if err := json.Unmarshal([]byte(jsonString), &productReq); err != nil {
		utils.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "Format JSON tidak valid"})
		return
	}

	files := r.MultipartForm.File["product_images"]
	if len(files) <= 0 {
		utils.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "Gambar wajib di unggah"})
		return
	}

	if len(files) > 10 {
		utils.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "Gambar yang di upload tidak boleh lebih dari 4"})
		return
	}

	resFiles, err := helper.ProcessImage(files)
	if err != nil {
		if errors.Is(err, apperror.ErrBadRequest) {
			utils.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		}
		utils.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	productReq.ImageURLs = resFiles

	if errs := validator.ValidateStruct(productReq); errs != nil {
		utils.WriteJSON(w, http.StatusBadRequest, map[string]any{
			"message": "validasi payload gagal",
			"error":   errs,
		})
		return
	}

	resGRPC, errGRPC := h.client.CreateProduct(ctx, &productpb.CreateProductRequest{ProductCode: productReq.ProductCode, ProductName: productReq.ProductName, Description: productReq.Description, Price: productReq.Price, InitialStock: productReq.Stock, ImageUrls: resFiles, CategoryId: productReq.CategoryID})
	if errGRPC != nil {
		Rollback(resFiles)
		errgrpc.WriteGRPCError(w, errGRPC)
		return
	}

	productRes := response.ProductResponse{
		ID:          resGRPC.GetProductId(),
		ProductCode: resGRPC.GetProductCode(),
		ProductName: resGRPC.GetProductName(),
		CategoryID:  resGRPC.GetCategoryId(),
		Price:       resGRPC.GetPrice(),
		Stock:       resGRPC.GetStock(),
		Description: resGRPC.GetDescription(),
		ImageURLs:   resFiles,
	}

	utils.WriteJSON(w, http.StatusCreated, map[string]any{
		"data": productRes,
	})
}

func Rollback(paths []string) {
	for _, path := range paths {
		_ = os.Remove(path)
	}
}

func (h *ProductHandlerImpl) FindProductHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	idReq := chi.URLParam(r, "id")
	if idReq == "" {
		utils.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "product ID wajib diisi"})
	}

	resGRPC, errGRPC := h.client.GetProduct(ctx, &productpb.GetProductRequest{ProductId: idReq})
	if errGRPC != nil {
		errgrpc.WriteGRPCError(w, errGRPC)
		return
	}

	productRes := response.ProductResponse{
		ID:          resGRPC.GetProductId(),
		ProductCode: resGRPC.GetProductCode(),
		ProductName: resGRPC.GetProductName(),
		CategoryID:  resGRPC.GetCategoryId(),
		Price:       resGRPC.GetPrice(),
		Stock:       resGRPC.GetStock(),
		Description: resGRPC.GetDescription(),
		ImageURLs:   resGRPC.ImageUrls,
		CreatedAt:   resGRPC.GetCreatedAt().AsTime(),
		UpdatedAt:   resGRPC.GetUpdatedAt().AsTime(),
	}

	utils.WriteJSON(w, http.StatusOK, map[string]any{
		"data": productRes,
	})
}

func (h *ProductHandlerImpl) StartFlashSaleEventHandler(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	ctx := r.Context()

	var fsRequest request.FlashSaleRequest

	if err := utils.ReadJSON(r, &fsRequest); err != nil {
		utils.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "Format JSON tidak valid"})
		return
	}

	if errs := validator.ValidateStruct(fsRequest); errs != nil {
		utils.WriteJSON(w, http.StatusBadRequest, map[string]any{
			"message": "validasi payload gagal",
			"error":   errs,
		})
		return
	}

	resGRPC, errGRPC := h.client.CreateFlashSaleProduct(ctx, &productpb.CreateFlashSaleRequest{ProductId: fsRequest.ProductID, FlashSalePrice: fsRequest.FlashSalePrice, FlashSaleStock: fsRequest.FlashSaleStock, MaxPurchasePerUser: fsRequest.MaxPurchasePerUser, StartTime: fsRequest.StartTime, EndTime: fsRequest.EndTime})
	if errGRPC != nil {
		errgrpc.WriteGRPCError(w, errGRPC)
		return
	}

	fsResponse := response.FlashSaleResponse{
		ProductID:          resGRPC.GetProductId(),
		FlashSalePrice:     resGRPC.GetFlashSalePrice(),
		FlashSaleStock:     resGRPC.GetFlashSaleStock(),
		MaxPurchasePerUser: resGRPC.GetMaxPurchasePerUser(),
		StartTime:          resGRPC.GetStartTime(),
		EndTime:            resGRPC.GetEndTime(),
		Version:            resGRPC.GetVersion(),
	}

	utils.WriteJSON(w, http.StatusCreated, map[string]any{
		"data": fsResponse,
	})
}

func (h *ProductHandlerImpl) FindFlashSale(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	flashSaleID := chi.URLParam(r, "id")
	if flashSaleID == "" {
		utils.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "flash sale ID wajib diisi"})
	}

	resGRPC, errGRPC := h.client.GetFlashSaleProduct(ctx, &productpb.GetFlashSaleRequest{FlashSaleId: flashSaleID})
	if errGRPC != nil {
		errgrpc.WriteGRPCError(w, errGRPC)
		return
	}

	fsResponse := response.FlashSaleResponse{
		ProductID:          resGRPC.GetProductId(),
		FlashSalePrice:     resGRPC.GetFlashSalePrice(),
		FlashSaleStock:     resGRPC.GetFlashSaleStock(),
		MaxPurchasePerUser: resGRPC.GetMaxPurchasePerUser(),
		StartTime:          resGRPC.GetStartTime(),
		EndTime:            resGRPC.GetEndTime(),
		Version:            resGRPC.GetVersion(),
	}

	utils.WriteJSON(w, http.StatusCreated, map[string]any{
		"data": fsResponse,
	})
}

func (h *ProductHandlerImpl) MonitorStockSSEHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Contect-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	ctx := r.Context()
	flashSaleID := chi.URLParam(r, "id")
	if flashSaleID == "" {
		utils.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "flash sale ID wajib diisi"})
	}

	grpcStream, err := h.client.MonitorStock(ctx, &productpb.MonitorStockRequest{FlashSaleId: flashSaleID})
	if err != nil {
		errgrpc.WriteGRPCError(w, err)
		return
	}

	for {
		select {
		case <-ctx.Done():
			return
		default:
			res, err := grpcStream.Recv()
			if err != nil {
				return
			}

			_, _ = fmt.Fprintf(w, "data: {\"flash_sale_id\": \"%s\", \"current_stock\": %d}\n\n", res.GetFlashSaleId(), res.GetCurrentStock())
			w.(http.Flusher).Flush()
		}
	}
}
