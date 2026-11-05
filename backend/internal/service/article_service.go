// backend/internal/service/article_service.go
// Article service：文章業務邏輯介面與實作，依賴 repository 介面

package service

import (
	"littleduck-backend/internal/model"
	"littleduck-backend/internal/pkg/apperror"
	"littleduck-backend/internal/repository"
)

// ArticleService 文章業務邏輯介面（controller 依賴此介面，測試以 mock 替換）
type ArticleService interface {
	List() ([]model.Article, error)
}

// articleService ArticleService 實作
type articleService struct {
	repo repository.ArticleRepository
}

// NewArticleService 建立 ArticleService
func NewArticleService(repo repository.ArticleRepository) ArticleService {
	return &articleService{repo: repo}
}

// List 取得文章列表；repository 錯誤包裝為 AppError
func (s *articleService) List() ([]model.Article, error) {
	articles, err := s.repo.FindAll()
	if err != nil {
		return nil, apperror.Internal(err, "取得文章列表失敗")
	}
	return articles, nil
}
