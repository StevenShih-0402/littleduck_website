// backend/internal/service/payment/ecpay_test.go
// ECPay service 單元測試：驗證設定載入與骨架方法回傳未實作

package payment

import (
	"net/http"
	"testing"

	"littleduck-backend/internal/config"
	"littleduck-backend/internal/pkg/apperror"
)

func TestNewECPayService_LoadsConfigFromEnvironment(t *testing.T) {
	// Arrange：模擬 config 套件由 ECPAY_* 環境變數載入的設定
	cfg := config.ECPayConfig{
		MerchantID: "2000132",
		HashKey:    "test-key",
		HashIV:     "test-iv",
		Env:        "stage",
	}

	// Act
	svc := NewECPayService(cfg)

	// Assert
	if svc.MerchantID() != "2000132" {
		t.Errorf("MerchantID = %q, want %q", svc.MerchantID(), "2000132")
	}
}

func TestECPayCreatePayment_ReturnsNotImplemented(t *testing.T) {
	// Arrange
	svc := NewECPayService(config.ECPayConfig{})

	// Act
	result, err := svc.CreatePayment(CreatePaymentInput{OrderID: "A001", Amount: 100})

	// Assert
	if result != nil {
		t.Errorf("骨架階段不應回傳結果: %+v", result)
	}
	appErr, ok := apperror.AsAppError(err)
	if !ok || appErr.Status != http.StatusNotImplemented {
		t.Errorf("應回傳 501 AppError，得到 %v", err)
	}
}

func TestECPayVerifyCallback_ReturnsNotImplemented(t *testing.T) {
	// Arrange
	svc := NewECPayService(config.ECPayConfig{})

	// Act
	err := svc.VerifyCallback(map[string]string{"CheckMacValue": "xxx"})

	// Assert
	appErr, ok := apperror.AsAppError(err)
	if !ok || appErr.Status != http.StatusNotImplemented {
		t.Errorf("應回傳 501 AppError，得到 %v", err)
	}
}
