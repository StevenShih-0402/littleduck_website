// backend/internal/repository/setting_repository.go
// SiteSetting repository：單列設定的讀取與更新

package repository

import (
	"errors"

	"gorm.io/gorm"

	"littleduck-backend/internal/model"
)

// SettingRepository 網站設定資料存取介面
type SettingRepository interface {
	// Get 取得單列設定；查無回傳 ErrNotFound
	Get() (*model.SiteSetting, error)
	// Save 覆寫單列設定（固定 id=1）
	Save(s *model.SiteSetting) error
	// Exists 單列設定是否已存在，供 seed 判斷
	Exists() (bool, error)
}

// gormSettingRepository SettingRepository 的 GORM 實作
type gormSettingRepository struct {
	db *gorm.DB
}

// NewSettingRepository 建立 GORM 版 SettingRepository
func NewSettingRepository(db *gorm.DB) SettingRepository {
	return &gormSettingRepository{db: db}
}

// Get 取得單列設定
func (r *gormSettingRepository) Get() (*model.SiteSetting, error) {
	var s model.SiteSetting
	if err := r.db.First(&s, "id = ?", model.SiteSettingID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &s, nil
}

// Save 覆寫單列設定
func (r *gormSettingRepository) Save(s *model.SiteSetting) error {
	s.ID = model.SiteSettingID
	return r.db.Save(s).Error
}

// Exists 單列設定是否存在
func (r *gormSettingRepository) Exists() (bool, error) {
	var n int64
	if err := r.db.Model(&model.SiteSetting{}).Count(&n).Error; err != nil {
		return false, err
	}
	return n > 0, nil
}
