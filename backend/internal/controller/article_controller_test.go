// backend/internal/controller/article_controller_test.go
// article controller 單元測試：以 mock service 驗證成功回應與錯誤轉交 middleware

package controller

import (
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
)

// mockArticleService 測試用 ArticleService
type mockArticleService struct {
	articles []model.Article
	err      error
}

func (m *mockArticleService) List() ([]model.Article, error) {
	return m.articles, m.err
}

// performArticles 掛載 article 路由（含 error handler）並發送請求
func performArticles(t *testing.T, svc *mockArticleService) *httptest.ResponseRecorder {
	t.Helper()
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(middleware.ErrorHandler(zap.NewNop()))
	r.GET("/api/v1/articles", NewArticleController(svc).List)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/v1/articles", nil))
	return rec
}

func TestArticleList_ReturnsEnvelopeWithArticles(t *testing.T) {
	// Arrange
	svc := &mockArticleService{articles: []model.Article{{Title: "hello"}}}

	// Act
	rec := performArticles(t, svc)

	// Assert
	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", rec.Code)
	}
	var body struct {
		Code string          `json:"code"`
		Data []model.Article `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("回應不是合法 JSON: %v", err)
	}
	if body.Code != response.CodeSuccess {
		t.Errorf("code = %q, want %q", body.Code, response.CodeSuccess)
	}
	if len(body.Data) != 1 || body.Data[0].Title != "hello" {
		t.Errorf("data 內容錯誤: %+v", body.Data)
	}
}

func TestArticleList_ServiceErrorHandledByMiddleware(t *testing.T) {
	// Arrange
	svc := &mockArticleService{err: apperror.Internal(nil, "取得文章列表失敗")}

	// Act
	rec := performArticles(t, svc)

	// Assert
	if rec.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want 500", rec.Code)
	}
	var body map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("回應不是合法 JSON: %v", err)
	}
	if body["code"] != apperror.CodeInternal {
		t.Errorf("code = %v, want %q", body["code"], apperror.CodeInternal)
	}
}
