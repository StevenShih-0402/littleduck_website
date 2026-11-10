// backend/internal/controller/article_controller_test.go
// article controller 單元測試：公開列表 envelope、單篇 404、建立 201（mock service）

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
	"littleduck-backend/internal/pkg/response"
	"littleduck-backend/internal/service"
)

// mockArticleService 測試用 ArticleService
type mockArticleService struct {
	published []model.Article
	all       []model.Article
	one       *model.Article
	getErr    error
	created   *model.Article
	createErr error
}

func (m *mockArticleService) ListPublished() ([]model.Article, error) { return m.published, nil }
func (m *mockArticleService) ListAll() ([]model.Article, error)       { return m.all, nil }
func (m *mockArticleService) Get(string) (*model.Article, error)      { return m.one, m.getErr }
func (m *mockArticleService) Create(service.ArticleInput) (*model.Article, error) {
	return m.created, m.createErr
}
func (m *mockArticleService) Update(string, service.ArticleInput) (*model.Article, error) {
	return m.one, m.getErr
}
func (m *mockArticleService) Delete(string) error { return nil }

func newArticleRouter(svc service.ArticleService) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(middleware.ErrorHandler(zap.NewNop()))
	ctrl := NewArticleController(svc)
	r.GET("/api/v1/articles", ctrl.List)
	r.GET("/api/v1/articles/:id", ctrl.Get)
	r.POST("/api/v1/articles", ctrl.Create)
	return r
}

func TestArticleList_ReturnsPublishedEnvelope(t *testing.T) {
	// Arrange
	svc := &mockArticleService{published: []model.Article{{ID: "a1", Title: "hello"}}}
	r := newArticleRouter(svc)

	// Act
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/v1/articles", nil))

	// Assert
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	var body struct {
		Code string          `json:"code"`
		Data []model.Article `json:"data"`
	}
	_ = json.Unmarshal(rec.Body.Bytes(), &body)
	if body.Code != response.CodeSuccess || len(body.Data) != 1 || body.Data[0].ID != "a1" {
		t.Errorf("回應內容錯誤: %+v", body)
	}
}

func TestArticleGet_NotFoundReturns404(t *testing.T) {
	// Arrange
	svc := &mockArticleService{getErr: apperror.NotFound("文章不存在")}
	r := newArticleRouter(svc)

	// Act
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/v1/articles/missing", nil))

	// Assert
	if rec.Code != http.StatusNotFound {
		t.Errorf("status = %d, want 404", rec.Code)
	}
}

func TestArticleCreate_Returns201(t *testing.T) {
	// Arrange
	svc := &mockArticleService{created: &model.Article{ID: "n1", Title: "新文章"}}
	r := newArticleRouter(svc)

	// Act
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/articles", bytes.NewBufferString(`{"title":"新文章"}`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(rec, req)

	// Assert
	if rec.Code != http.StatusCreated {
		t.Errorf("status = %d, want 201", rec.Code)
	}
}
