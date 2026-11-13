// backend/internal/controller/comment_controller_test.go
// comment controller 單元測試：公開讀取可見留言、張貼 201 pending、空內容 400（mock service）

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

// mockCommentService 測試用 CommentService
type mockCommentService struct {
	visible   []model.Comment
	all       []model.Comment
	created   *model.Comment
	createErr error
	gotAdmin  *bool
}

func (m *mockCommentService) ListVisible(string) ([]model.Comment, error) { return m.visible, nil }
func (m *mockCommentService) ListAll() ([]model.Comment, error)           { return m.all, nil }
func (m *mockCommentService) Create(_ service.CommentInput, asAdmin bool) (*model.Comment, error) {
	m.gotAdmin = &asAdmin
	return m.created, m.createErr
}
func (m *mockCommentService) SetStatus(string, string) (*model.Comment, error) { return nil, nil }
func (m *mockCommentService) Reply(string, string) (*model.Comment, error)     { return nil, nil }

func newCommentRouter(svc service.CommentService) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(middleware.ErrorHandler(zap.NewNop()))
	ctrl := NewCommentController(svc)
	r.GET("/api/v1/comments", ctrl.List)
	r.POST("/api/v1/comments", ctrl.Create)
	return r
}

func TestCommentList_ReturnsVisibleOnly(t *testing.T) {
	// Arrange：service 僅回 visible
	svc := &mockCommentService{visible: []model.Comment{{ID: "c1", Status: model.CommentStatusVisible}}}
	r := newCommentRouter(svc)

	// Act
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/v1/comments", nil))

	// Assert
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	var body struct {
		Data []model.Comment `json:"data"`
	}
	_ = json.Unmarshal(rec.Body.Bytes(), &body)
	if len(body.Data) != 1 || body.Data[0].Status != model.CommentStatusVisible {
		t.Errorf("公開結果錯誤: %+v", body.Data)
	}
}

func TestCommentCreate_Returns201(t *testing.T) {
	// Arrange
	svc := &mockCommentService{created: &model.Comment{ID: "n1", Status: model.CommentStatusPending}}
	r := newCommentRouter(svc)

	// Act
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/comments", bytes.NewBufferString(`{"name":"訪客","text":"你好"}`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(rec, req)

	// Assert
	if rec.Code != http.StatusCreated {
		t.Errorf("status = %d, want 201", rec.Code)
	}
}

func TestCommentCreate_EmptyTextReturns400(t *testing.T) {
	// Arrange
	svc := &mockCommentService{createErr: apperror.BadRequest("留言名稱與內容不可為空")}
	r := newCommentRouter(svc)

	// Act
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/comments", bytes.NewBufferString(`{"name":"訪客","text":""}`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(rec, req)

	// Assert
	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", rec.Code)
	}
}

// postCommentWithAdminCtx 以指定的 ContextKeyAdmin 設定（nil＝不設）送出張貼請求
func postCommentWithAdminCtx(svc service.CommentService, subject *string) *httptest.ResponseRecorder {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(middleware.ErrorHandler(zap.NewNop()))
	r.POST("/api/v1/comments", func(c *gin.Context) {
		if subject != nil {
			c.Set(middleware.ContextKeyAdmin, *subject)
		}
	}, NewCommentController(svc).Create)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/comments", bytes.NewBufferString(`{"name":"訪客","text":"你好"}`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(rec, req)
	return rec
}

func TestCommentCreate_AdminContextPassesAsAdminTrue(t *testing.T) {
	// Arrange
	svc := &mockCommentService{created: &model.Comment{ID: "n1"}}
	admin := "admin"

	// Act
	rec := postCommentWithAdminCtx(svc, &admin)

	// Assert
	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, want 201", rec.Code)
	}
	if svc.gotAdmin == nil || !*svc.gotAdmin {
		t.Error("有管理員身分時 asAdmin 應為 true")
	}
}

func TestCommentCreate_VisitorPassesAsAdminFalse(t *testing.T) {
	// Arrange
	svc := &mockCommentService{created: &model.Comment{ID: "n1"}}

	// Act
	rec := postCommentWithAdminCtx(svc, nil)

	// Assert
	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, want 201", rec.Code)
	}
	if svc.gotAdmin == nil || *svc.gotAdmin {
		t.Error("訪客 asAdmin 應為 false")
	}
}

func TestCommentCreate_FeatureDisabledReturns403(t *testing.T) {
	// Arrange
	svc := &mockCommentService{createErr: apperror.FeatureDisabled("留言板投書功能目前已停用")}

	// Act
	rec := postCommentWithAdminCtx(svc, nil)

	// Assert
	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want 403", rec.Code)
	}
	var body struct {
		Code string `json:"code"`
	}
	_ = json.Unmarshal(rec.Body.Bytes(), &body)
	if body.Code != "FEATURE_DISABLED" {
		t.Errorf("code = %q, want FEATURE_DISABLED", body.Code)
	}
}
