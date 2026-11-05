// backend/internal/controller/payment_controller_test.go
// payment controller 單元測試：驗證預留端點回傳 501 與統一錯誤格式

package controller

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"littleduck-backend/internal/config"
	"littleduck-backend/internal/middleware"
	"littleduck-backend/internal/pkg/apperror"
	"littleduck-backend/internal/service/payment"
)

// performPayment 掛載 payment 路由（含 error handler）並發送請求
func performPayment(t *testing.T, path string) *httptest.ResponseRecorder {
	t.Helper()
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(middleware.ErrorHandler(zap.NewNop()))
	ctrl := NewPaymentController(payment.NewECPayService(config.ECPayConfig{}))
	r.POST("/api/v1/payments", ctrl.Create)
	r.POST("/api/v1/payments/ecpay/callback", ctrl.ECPayCallback)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, path, nil))
	return rec
}

// assertNotImplemented 檢查回應為 501 與統一 envelope 格式
func assertNotImplemented(t *testing.T, rec *httptest.ResponseRecorder) {
	t.Helper()
	if rec.Code != http.StatusNotImplemented {
		t.Errorf("status = %d, want 501", rec.Code)
	}
	var body map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("回應不是合法 JSON: %v", err)
	}
	if body["code"] != apperror.CodeNotImplemented {
		t.Errorf("code = %v, want %q", body["code"], apperror.CodeNotImplemented)
	}
}

func TestCreatePayment_Returns501(t *testing.T) {
	// Arrange & Act
	rec := performPayment(t, "/api/v1/payments")

	// Assert
	assertNotImplemented(t, rec)
}

func TestECPayCallback_Returns501(t *testing.T) {
	// Arrange & Act
	rec := performPayment(t, "/api/v1/payments/ecpay/callback")

	// Assert
	assertNotImplemented(t, rec)
}
