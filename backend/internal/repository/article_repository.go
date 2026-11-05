// backend/internal/repository/article_repository.go
// Article repository：資料存取介面與 GORM 實作

package repository

import (
	"gorm.io/gorm"

	"littleduck-backend/internal/model"
)

// ArticleRepository 文章資料存取介面（service 依賴此介面，測試以 mock 替換）
type ArticleRepository interface {
	FindAll() ([]model.Article, error)
}

// gormArticleRepository ArticleRepository 的 GORM 實作
type gormArticleRepository struct {
	db *gorm.DB
}

// NewArticleRepository 建立 GORM 版 ArticleRepository
func NewArticleRepository(db *gorm.DB) ArticleRepository {
	return &gormArticleRepository{db: db}
}

// FindAll 取得全部文章（依建立時間新到舊）
func (r *gormArticleRepository) FindAll() ([]model.Article, error) {
	var articles []model.Article
	if err := r.db.Order("created_at DESC").Find(&articles).Error; err != nil {
		return nil, err
	}
	return articles, nil
}
