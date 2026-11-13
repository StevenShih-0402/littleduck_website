// backend/internal/repository/feature_repository.go
// FeatureFlag repository：功能開關的讀取、更新與缺漏補齊

package repository

import (
	"errors"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"littleduck-backend/internal/model"
)

// FeatureRepository 功能開關資料存取介面
type FeatureRepository interface {
	// FindAll 取得全部開關列
	FindAll() ([]model.FeatureFlag, error)
	// Find 取得單一開關；查無回傳 ErrNotFound
	Find(key string) (*model.FeatureFlag, error)
	// Save 寫入（建立或覆寫）開關
	Save(f *model.FeatureFlag) error
	// CreateIfMissing 只在 key 不存在時建立，已存在則不覆寫
	CreateIfMissing(f *model.FeatureFlag) error
}

// gormFeatureRepository FeatureRepository 的 GORM 實作
type gormFeatureRepository struct {
	db *gorm.DB
}

// NewFeatureRepository 建立 GORM 版 FeatureRepository
func NewFeatureRepository(db *gorm.DB) FeatureRepository {
	return &gormFeatureRepository{db: db}
}

// FindAll 取得全部開關
func (r *gormFeatureRepository) FindAll() ([]model.FeatureFlag, error) {
	var flags []model.FeatureFlag
	if err := r.db.Order(clause.OrderByColumn{Column: clause.Column{Name: "key"}}).Find(&flags).Error; err != nil {
		return nil, err
	}
	return flags, nil
}

// Find 取得單一開關
func (r *gormFeatureRepository) Find(key string) (*model.FeatureFlag, error) {
	var f model.FeatureFlag
	if err := r.db.First(&f, clause.Eq{Column: clause.Column{Name: "key"}, Value: key}).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &f, nil
}

// Save 建立或覆寫開關
func (r *gormFeatureRepository) Save(f *model.FeatureFlag) error {
	return r.db.Save(f).Error
}

// CreateIfMissing 只補缺：衝突時不做任何事
func (r *gormFeatureRepository) CreateIfMissing(f *model.FeatureFlag) error {
	return r.db.Clauses(clause.OnConflict{DoNothing: true}).Create(f).Error
}
