// backend/internal/router/router_test.go
// router 單元測試：驗證路由掛載（health、api/v1、swagger）與未初始化 DB 時的健康檢查

package router

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"go.uber.org/zap"

	"littleduck-backend/internal/config"
)

// newTestRouter 以測試組態建立 router（db 為 nil，模擬未初始化）
func newTestRouter() http.Handler {
	cfg := &config.Config{}
	cfg.Server.GinMode = "test"
	return New(cfg, zap.NewNop(), nil)
}

func TestNew_MountsHealthRoute(t *testing.T) {
	// Arrange
	r := newTestRouter()
	rec := httptest.NewRecorder()

	// Act：db 未初始化，健康檢查應回報 DB 異常
	r.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/health", nil))

	// Assert
	if rec.Code != http.StatusServiceUnavailable {
		t.Errorf("/health（無 DB）status = %d, want 503", rec.Code)
	}
}

func TestNew_MountsSwaggerRoute(t *testing.T) {
	// Arrange
	r := newTestRouter()
	rec := httptest.NewRecorder()

	// Act
	r.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/swagger/index.html", nil))

	// Assert
	if rec.Code != http.StatusOK {
		t.Errorf("/swagger/index.html status = %d, want 200", rec.Code)
	}
}

func TestNew_MountsAPIV1Routes(t *testing.T) {
	// Arrange
	r := newTestRouter()

	// Act & Assert：路由存在（非 404）即代表已掛載
	for _, tc := range []struct{ method, path string }{
		{http.MethodGet, "/api/v1/articles"},
		{http.MethodPost, "/api/v1/payments"},
		{http.MethodPost, "/api/v1/payments/ecpay/callback"},
	} {
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, httptest.NewRequest(tc.method, tc.path, nil))
		if rec.Code == http.StatusNotFound {
			t.Errorf("%s %s 未掛載（404）", tc.method, tc.path)
		}
	}
}
