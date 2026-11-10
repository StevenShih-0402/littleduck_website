// backend/internal/controller/setting_controller_test.go
// setting controller 單元測試：讀取 envelope、更新成功、空標題 400（mock service）

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
	"littleduck-backend/internal/model"
	"littleduck-backend/internal/pkg/apperror"
	"littleduck-backend/internal/service"
)

// mockSettingService 測試用 SettingService
type mockSettingService struct {
	current   *model.SiteSetting
	updateErr error
}

func (m *mockSettingService) Get() (*model.SiteSetting, error) { return m.current, nil }
func (m *mockSettingService) Update(in model.SiteSetting) (*model.SiteSetting, error) {
	if m.updateErr != nil {
		return nil, m.updateErr
	}
	m.current = &in
	return &in, nil
}

func newSettingRouter(svc service.SettingService) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(middleware.ErrorHandler(zap.NewNop()))
	ctrl := NewSettingController(svc)
	r.GET("/api/v1/settings", ctrl.Get)
	r.PUT("/api/v1/settings", ctrl.Update)
	return r
}

func TestSettingGet_ReturnsEnvelope(t *testing.T) {
	// Arrange
	svc := &mockSettingService{current: &model.SiteSetting{Title: "小鴨"}}
	r := newSettingRouter(svc)

	// Act
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/v1/settings", nil))

	// Assert
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	var body struct {
		Data model.SiteSetting `json:"data"`
	}
	_ = json.Unmarshal(rec.Body.Bytes(), &body)
	if body.Data.Title != "小鴨" {
		t.Errorf("讀取內容錯誤: %+v", body.Data)
	}
}

func TestSettingUpdate_Success(t *testing.T) {
	// Arrange
	svc := &mockSettingService{current: &model.SiteSetting{Title: "舊"}}
	r := newSettingRouter(svc)

	// Act
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPut, "/api/v1/settings", bytes.NewBufferString(`{"title":"新"}`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(rec, req)

	// Assert
	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", rec.Code)
	}
}

func TestSettingUpdate_EmptyTitleReturns400(t *testing.T) {
	// Arrange
	svc := &mockSettingService{updateErr: apperror.BadRequest("網站標題不可為空")}
	r := newSettingRouter(svc)

	// Act
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPut, "/api/v1/settings", bytes.NewBufferString(`{"title":""}`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(rec, req)

	// Assert
	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", rec.Code)
	}
}
