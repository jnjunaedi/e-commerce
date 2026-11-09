package usecase_test

import (
	"bytes"
	"context"
	"e-commerce/payment_service/infrastructure/client"
	mGrpc "e-commerce/payment_service/infrastructure/client/mocks"
	"e-commerce/payment_service/internal/domain"
	mRepo "e-commerce/payment_service/internal/repository/mocks"
	mTxManager "e-commerce/payment_service/internal/repository/mocks"
	"e-commerce/payment_service/internal/usecase"
	mId "e-commerce/pkg/utils/mocks"
	"encoding/json"
	"io"
	"net/http"
	"testing"

	"github.com/midtrans/midtrans-go"
	"github.com/midtrans/midtrans-go/coreapi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockTransport struct {
	response *http.Response
	err      error
}

func (m *mockTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	return m.response, m.err
}

type midtransHttpClientMock struct {
	mockClient *http.Client
}

func (m *midtransHttpClientMock) Call(method string, url string, secret *string, options *midtrans.ConfigOptions, body io.Reader, v interface{}) *midtrans.Error {
	req, _ := http.NewRequest(method, url, body)

	resp, err := m.mockClient.Do(req)
	if err != nil {
		return &midtrans.Error{Message: err.Error()}
	}
	defer resp.Body.Close()

	if v != nil && resp.StatusCode >= 200 && resp.StatusCode < 300 {
		_ = json.NewDecoder(resp.Body).Decode(v)
	}

	return nil
}

func TestPaymentUsecase_PaymentProcess(t *testing.T) {
	mockRepo := mRepo.NewRepository(t)
	mockClient := mGrpc.NewOrderGrpcClient(t)
	mockId := mId.NewIDGenerator(t)
	mockTxManager := mTxManager.NewTxManager(t)

	pay := domain.Payment{
		OrderID:       "ODR-117",
		PaymentMethod: "bank_transfer",
		PaymentAmount: 10000000,
	}

	order := client.OrderResult{
		UserID:         "USR-113",
		TotalPrice:     45000,
		OrderStatus:    "PENDING_PAYMENT",
		IdempotencyKey: "lpt-reguler-prd-132",
	}

	jsonResponse := `{
		"status_code": "201",
		"status_message": "Success, Bank Transfer transaction is created",
		"transaction_id": "tx-uuid-12345",
		"order_id": "ODR-117",
		"gross_amount": "45000.00",
		"payment_type": "bank_transfer",
		"transaction_status": "pending",
		"va_numbers": [
			{
				"bank": "bca",
				"va_number": "1234567890"
			}
		]
	}`

	fakeHttpClient := &http.Client{
		Transport: &mockTransport{
			response: &http.Response{
				StatusCode: 201,
				Body:       io.NopCloser(bytes.NewBufferString(jsonResponse)),
				Header:     make(http.Header),
			},
		},
	}

	midClient := coreapi.Client{}
	midClient.New("palsu-server-key", midtrans.Sandbox)
	midClient.HttpClient = &midtransHttpClientMock{mockClient: fakeHttpClient}

	idDummy := "gen-pay-112"

	mockId.On("GenerateUUID").Return(idDummy, nil)

	mockClient.On("GetOrderDetails", mock.Anything, mock.Anything).Return(order, nil)

	mockRepo.On("Create", mock.Anything, mock.MatchedBy(func(p domain.Payment) bool {
		return p.PaymentStatus == "PENDING" && p.OrderID == "ODR-117"
	})).Return(nil)

	uc := usecase.NewUsecase(mockRepo, mockClient, mockTxManager, midClient, mockId)
	ctx := context.Background()

	res, err := uc.Process(ctx, pay)

	assert.NoError(t, err)
	assert.Equal(t, pay.PaymentAmount, res.PaymentAmount)

	mockRepo.AssertExpectations(t)
	mockClient.AssertExpectations(t)
	mockId.AssertExpectations(t)
	mockTxManager.AssertExpectations(t)
}

func TestPaymentUsecase_HandlerWebhook(t *testing.T) {
	mockRepo := mRepo.NewRepository(t)
	mockClient := mGrpc.NewOrderGrpcClient(t)
	mockId := mId.NewIDGenerator(t)
	mockTxManager := mTxManager.NewTxManager(t)

	jsonResponse := `{
		 "transaction_status": "settlement",
    	 "fraud_status": "accept",
    	 "order_id": "ODR-117"
	}`

	fakeHttpClient := &http.Client{
		Transport: &mockTransport{
			response: &http.Response{
				StatusCode: 201,
				Body:       io.NopCloser(bytes.NewBufferString(jsonResponse)),
				Header:     make(http.Header),
			},
		},
	}

	outboxID := "outbox-123"

	midClient := coreapi.Client{}
	midClient.New("palsu-server-key", midtrans.Sandbox)
	midClient.HttpClient = &midtransHttpClientMock{mockClient: fakeHttpClient}

	mockTxManager.On("WithTransaction", mock.Anything, mock.Anything).Return(func(ctx context.Context, fn func(ctx context.Context) error) error {
		return fn(ctx)
	}).Once()

	mockRepo.On("UpdateStatus", mock.Anything, mock.Anything, mock.Anything, mock.AnythingOfType("int64")).Return(nil)

	mockId.On("GenerateUUID").Return(outboxID, nil)

	mockRepo.On("CreateOutbox", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)

	uc := usecase.NewUsecase(mockRepo, mockClient, mockTxManager, midClient, mockId)
	ctx := context.Background()

	err := uc.HandleWebhook(ctx, map[string]interface{}{
		"order_id": "ODR-117",
	})

	assert.NoError(t, err)

	mockRepo.AssertExpectations(t)
	mockClient.AssertExpectations(t)
	mockId.AssertExpectations(t)
	mockTxManager.AssertExpectations(t)

}
