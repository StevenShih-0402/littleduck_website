// backend/internal/controller/stat_controller_test.go
// stat controller 單元測試：瀏覽回報回 204（即使記錄失敗）、每日統計回 envelope（mock service）

package controller

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"littleduck-backend/internal/middleware"
	"littleduck-backend/internal/pkg/apperror"
	"littleduck-backend/internal/repository"
	"littleduck-backend/internal/service"
)

// mockStatService 測試用 StatService
type mockStatService struct {
	daily     []repository.DailyStat
	recordErr error
}

func (m *mockStatService) RecordView(string, string, string) error { return m.recordErr }
func (m *mockStatService) Daily() ([]repository.DailyStat, error)  { return m.daily, nil }

func newStatRouter(svc service.StatService) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(middleware.ErrorHandler(zap.NewNop()))
	ctrl := NewStatController(svc)
	r.POST("/api/v1/stats/pageview", ctrl.RecordPageview)
	r.GET("/api/v1/stats", ctrl.Daily)
	return r
}

func TestPageview_Returns204(t *testing.T) {
	// Arrange
	svc := &mockStatService{}
	r := newStatRouter(svc)

	// Act
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/stats/pageview", bytes.NewBufferString(`{"path":"/"}`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(rec, req)

	// Assert
	if rec.Code != http.StatusNoContent {
		t.Errorf("status = %d, want 204", rec.Code)
	}
}

func TestPageview_NonBlockingEvenWhenRecordFails(t *testing.T) {
	// Arrange：記錄失敗仍不阻擋前台（回 204）
	svc := &mockStatService{recordErr: apperror.Internal(nil, "db down")}
	r := newStatRouter(svc)

	// Act
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/stats/pageview", bytes.NewBufferString(`{}`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(rec, req)

	// Assert
	if rec.Code != http.StatusNoContent {
		t.Errorf("記錄失敗時仍應回 204，得到 %d", rec.Code)
	}
}

func TestStatDaily_ReturnsEnvelope(t *testing.T) {
	// Arrange
	svc := &mockStatService{daily: []repository.DailyStat{{Day: "2026-07-02", Views: 10, Visitors: 4}}}
	r := newStatRouter(svc)

	// Act
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/v1/stats", nil))

	// Assert
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	var body struct {
		Data []repository.DailyStat `json:"data"`
	}
	_ = json.Unmarshal(rec.Body.Bytes(), &body)
	if len(body.Data) != 1 || body.Data[0].Views != 10 {
		t.Errorf("統計內容錯誤: %+v", body.Data)
	}
}
