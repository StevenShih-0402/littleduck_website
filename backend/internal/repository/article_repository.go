// backend/internal/repository/article_repository.go
// Article repository：資料存取介面與 GORM 實作

package repository

import (
	"errors"

	"gorm.io/gorm"

	"littleduck-backend/internal/model"
)

// ErrNotFound repository 層的「查無資料」標記錯誤，供 service 轉為 404
var ErrNotFound = errors.New("查無資料")

// ArticleRepository 文章資料存取介面（service 依賴此介面，測試以 mock 替換）
type ArticleRepository interface {
	// FindAll 取得文章；onlyPublished 為 true 時僅回傳已發布
	FindAll(onlyPublished bool) ([]model.Article, error)
	// FindByID 以 id 取得單篇；查無回傳 ErrNotFound
	FindByID(id string) (*model.Article, error)
	Create(a *model.Article) error
	Update(a *model.Article) error
	Delete(id string) error
	// Count 回傳文章總數，供 seed 判斷資料表是否為空
	Count() (int64, error)
}

// gormArticleRepository ArticleRepository 的 GORM 實作
type gormArticleRepository struct {
	db *gorm.DB
}

// NewArticleRepository 建立 GORM 版 ArticleRepository
func NewArticleRepository(db *gorm.DB) ArticleRepository {
	return &gormArticleRepository{db: db}
}

// FindAll 取得文章（依建立時間新到舊）；onlyPublished 時只回已發布
func (r *gormArticleRepository) FindAll(onlyPublished bool) ([]model.Article, error) {
	var articles []model.Article
	q := r.db.Order("created_at DESC")
	if onlyPublished {
		q = q.Where("status = ?", model.ArticleStatusPublished)
	}
	if err := q.Find(&articles).Error; err != nil {
		return nil, err
	}
	return articles, nil
}

// FindByID 以 id 取得單篇
func (r *gormArticleRepository) FindByID(id string) (*model.Article, error) {
	var a model.Article
	if err := r.db.First(&a, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &a, nil
}

// Create 新增文章
func (r *gormArticleRepository) Create(a *model.Article) error {
	return r.db.Create(a).Error
}

// Update 以主鍵覆寫全部欄位
func (r *gormArticleRepository) Update(a *model.Article) error {
	return r.db.Save(a).Error
}

// Delete 以 id 刪除；影響 0 列回 ErrNotFound
func (r *gormArticleRepository) Delete(id string) error {
	res := r.db.Delete(&model.Article{}, "id = ?", id)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

// Count 回傳文章總數
func (r *gormArticleRepository) Count() (int64, error) {
	var n int64
	if err := r.db.Model(&model.Article{}).Count(&n).Error; err != nil {
		return 0, err
	}
	return n, nil
}
