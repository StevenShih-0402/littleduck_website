// backend/internal/controller/feature_controller_test.go
// feature controller 單元測試：公開對照表、後台清單、切換成功與 400／404 錯誤（mock service）

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
	"littleduck-backend/internal/service"
)

// mockFeatureService 測試用 FeatureService
type mockFeatureService struct {
	public    map[string]bool
	list      []service.FeatureView
	updateErr error
	gotKey    string
	gotValue  *bool
}

func (m *mockFeatureService) IsEnabled(string) (bool, error)       { return true, nil }
func (m *mockFeatureService) PublicMap() (map[string]bool, error)  { return m.public, nil }
func (m *mockFeatureService) List() ([]service.FeatureView, error) { return m.list, nil }
func (m *mockFeatureService) EnsureDefaults() error                { return nil }
func (m *mockFeatureService) Update(key string, enabled *bool) (*service.FeatureView, error) {
	m.gotKey, m.gotValue = key, enabled
	if m.updateErr != nil {
		return nil, m.updateErr
	}
	return &service.FeatureView{Key: key, Enabled: *enabled}, nil
}

func newFeatureRouter(svc service.FeatureService) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(middleware.ErrorHandler(zap.NewNop()))
	ctrl := NewFeatureController(svc)
	r.GET("/api/v1/features", ctrl.Public)
	r.GET("/api/v1/admin/features", ctrl.AdminList)
	r.PUT("/api/v1/features/:key", ctrl.Update)
	return r
}

func putFeature(r *gin.Engine, key, body string) *httptest.ResponseRecorder {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPut, "/api/v1/features/"+key, bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(rec, req)
	return rec
}

func TestFeaturePublic_ReturnsBooleanMap(t *testing.T) {
	// Arrange
	r := newFeatureRouter(&mockFeatureService{public: map[string]bool{"guestbook": false, "sponsor": true}})

	// Act
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/v1/features", nil))

	// Assert
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	var body struct {
		Data map[string]bool `json:"data"`
	}
	_ = json.Unmarshal(rec.Body.Bytes(), &body)
	if v, ok := body.Data["guestbook"]; !ok || v || !body.Data["sponsor"] {
		t.Errorf("對照表錯誤: %+v", body.Data)
	}
}

func TestFeatureAdminList_ReturnsViews(t *testing.T) {
	// Arrange
	r := newFeatureRouter(&mockFeatureService{list: []service.FeatureView{{Key: "sponsor", Name: "贊助按鈕", Enabled: true}}})

	// Act
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/v1/admin/features", nil))

	// Assert
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	var body struct {
		Data []service.FeatureView `json:"data"`
	}
	_ = json.Unmarshal(rec.Body.Bytes(), &body)
	if len(body.Data) != 1 || body.Data[0].Name != "贊助按鈕" {
		t.Errorf("清單錯誤: %+v", body.Data)
	}
}

func TestFeatureUpdate_FalseIsPassedThrough(t *testing.T) {
	// Arrange
	svc := &mockFeatureService{}
	r := newFeatureRouter(svc)

	// Act
	rec := putFeature(r, "guestbook", `{"enabled":false}`)

	// Assert
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	if svc.gotKey != "guestbook" || svc.gotValue == nil || *svc.gotValue {
		t.Errorf("應以 key=guestbook、enabled=false 呼叫 service: %q %v", svc.gotKey, svc.gotValue)
	}
}

func TestFeatureUpdate_MissingFieldPassesNilToService(t *testing.T) {
	// Arrange：由 service 判斷缺欄位回 400
	svc := &mockFeatureService{updateErr: apperror.BadRequest("請提供 enabled 布林值")}
	r := newFeatureRouter(svc)

	// Act
	rec := putFeature(r, "guestbook", `{}`)

	// Assert
	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", rec.Code)
	}
	if svc.gotValue != nil {
		t.Error("缺欄位時應傳 nil")
	}
}

func TestFeatureUpdate_WrongTypeReturns400(t *testing.T) {
	// Arrange
	r := newFeatureRouter(&mockFeatureService{})

	// Act
	rec := putFeature(r, "guestbook", `{"enabled":"yes"}`)

	// Assert
	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", rec.Code)
	}
}

func TestFeatureUpdate_UnknownKeyReturns404(t *testing.T) {
	// Arrange
	r := newFeatureRouter(&mockFeatureService{updateErr: apperror.NotFound("功能開關不存在")})

	// Act
	rec := putFeature(r, "nope", `{"enabled":true}`)

	// Assert
	if rec.Code != http.StatusNotFound {
		t.Errorf("status = %d, want 404", rec.Code)
	}
}
