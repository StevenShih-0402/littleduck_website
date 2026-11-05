// backend/internal/service/article_service_test.go
// article service 單元測試：以 mock repository 驗證列表邏輯與錯誤包裝

package service

import (
	"errors"
	"net/http"
	"testing"

	"littleduck-backend/internal/model"
	"littleduck-backend/internal/pkg/apperror"
)

// mockArticleRepo 測試用 ArticleRepository
type mockArticleRepo struct {
	articles []model.Article
	err      error
}

func (m *mockArticleRepo) FindAll() ([]model.Article, error) {
	return m.articles, m.err
}

func TestArticleServiceList_ReturnsArticlesFromRepository(t *testing.T) {
	// Arrange
	repo := &mockArticleRepo{articles: []model.Article{
		{Title: "第一篇"},
		{Title: "第二篇"},
	}}
	svc := NewArticleService(repo)

	// Act
	articles, err := svc.List()

	// Assert
	if err != nil {
		t.Fatalf("List() 應成功，卻回傳錯誤: %v", err)
	}
	if len(articles) != 2 || articles[0].Title != "第一篇" {
		t.Errorf("回傳內容錯誤: %+v", articles)
	}
}

func TestArticleServiceList_WrapsRepositoryErrorAsAppError(t *testing.T) {
	// Arrange
	repo := &mockArticleRepo{err: errors.New("db down")}
	svc := NewArticleService(repo)

	// Act
	_, err := svc.List()

	// Assert
	if err == nil {
		t.Fatal("repository 失敗時 List() 應回傳錯誤")
	}
	appErr, ok := apperror.AsAppError(err)
	if !ok {
		t.Fatalf("錯誤應為 AppError，得到 %T", err)
	}
	if appErr.Status != http.StatusInternalServerError {
		t.Errorf("Status = %d, want 500", appErr.Status)
	}
}
