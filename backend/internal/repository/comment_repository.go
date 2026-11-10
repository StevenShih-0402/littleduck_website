// backend/internal/repository/comment_repository.go
// Comment repository：資料存取介面與 GORM 實作

package repository

import (
	"errors"

	"gorm.io/gorm"

	"littleduck-backend/internal/model"
)

// CommentRepository 留言資料存取介面
type CommentRepository interface {
	// FindVisible 公開讀取：僅 visible；articleID 非空時依文章過濾
	FindVisible(articleID string) ([]model.Comment, error)
	// FindAll 後台讀取：全部狀態
	FindAll() ([]model.Comment, error)
	FindByID(id string) (*model.Comment, error)
	Create(c *model.Comment) error
	Update(c *model.Comment) error
	Count() (int64, error)
}

// gormCommentRepository CommentRepository 的 GORM 實作
type gormCommentRepository struct {
	db *gorm.DB
}

// NewCommentRepository 建立 GORM 版 CommentRepository
func NewCommentRepository(db *gorm.DB) CommentRepository {
	return &gormCommentRepository{db: db}
}

// FindVisible 取得可見留言
func (r *gormCommentRepository) FindVisible(articleID string) ([]model.Comment, error) {
	var comments []model.Comment
	q := r.db.Order("created_at DESC").Where("status = ?", model.CommentStatusVisible)
	if articleID != "" {
		q = q.Where("article_id = ?", articleID)
	}
	if err := q.Find(&comments).Error; err != nil {
		return nil, err
	}
	return comments, nil
}

// FindAll 取得全部留言
func (r *gormCommentRepository) FindAll() ([]model.Comment, error) {
	var comments []model.Comment
	if err := r.db.Order("created_at DESC").Find(&comments).Error; err != nil {
		return nil, err
	}
	return comments, nil
}

// FindByID 以 id 取得單則
func (r *gormCommentRepository) FindByID(id string) (*model.Comment, error) {
	var c model.Comment
	if err := r.db.First(&c, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &c, nil
}

// Create 新增留言
func (r *gormCommentRepository) Create(c *model.Comment) error {
	return r.db.Create(c).Error
}

// Update 覆寫留言
func (r *gormCommentRepository) Update(c *model.Comment) error {
	return r.db.Save(c).Error
}

// Count 回傳留言總數
func (r *gormCommentRepository) Count() (int64, error) {
	var n int64
	if err := r.db.Model(&model.Comment{}).Count(&n).Error; err != nil {
		return 0, err
	}
	return n, nil
}
