// backend/internal/controller/health_controller_test.go
// health controller 單元測試：以 mock Pinger 驗證 DB 正常與異常時的回應

package controller

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

// stubPinger 測試用 Pinger，回傳預先設定的錯誤
type stubPinger struct {
	err error
}

func (s stubPinger) Ping() error { return s.err }

// performHealth 掛載 health 路由並發送請求
func performHealth(t *testing.T, pinger Pinger) *httptest.ResponseRecorder {
	t.Helper()
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/health", NewHealthController(pinger).Check)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/health", nil))
	return rec
}

func TestHealthCheck_DatabaseUpReturns200(t *testing.T) {
	// Arrange
	pinger := stubPinger{err: nil}

	// Act
	rec := performHealth(t, pinger)

	// Assert
	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", rec.Code)
	}
	var body map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("回應不是合法 JSON: %v", err)
	}
	if body["status"] != "ok" || body["database"] != "up" {
		t.Errorf("回應內容錯誤: %+v", body)
	}
}

func TestHealthCheck_DatabaseDownReturnsNon200(t *testing.T) {
	// Arrange
	pinger := stubPinger{err: errors.New("connection refused")}

	// Act
	rec := performHealth(t, pinger)

	// Assert
	if rec.Code != http.StatusServiceUnavailable {
		t.Errorf("status = %d, want 503", rec.Code)
	}
	var body map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("回應不是合法 JSON: %v", err)
	}
	if body["database"] != "down" {
		t.Errorf("DB 異常時應標示 database=down: %+v", body)
	}
}
