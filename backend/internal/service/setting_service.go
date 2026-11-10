// backend/internal/service/setting_service.go
// Setting service：網站設定讀取與更新

package service

import (
	"errors"

	"littleduck-backend/internal/model"
	"littleduck-backend/internal/pkg/apperror"
	"littleduck-backend/internal/repository"
)

// SettingService 網站設定業務邏輯介面
type SettingService interface {
	// Get 讀取目前設定
	Get() (*model.SiteSetting, error)
	// Update 更新設定（基本驗證：標題不可為空）
	Update(in model.SiteSetting) (*model.SiteSetting, error)
}

// settingService SettingService 實作
type settingService struct {
	repo repository.SettingRepository
}

// NewSettingService 建立 SettingService
func NewSettingService(repo repository.SettingRepository) SettingService {
	return &settingService{repo: repo}
}

// Get 讀取設定；查無回 404
func (s *settingService) Get() (*model.SiteSetting, error) {
	setting, err := s.repo.Get()
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, apperror.NotFound("網站設定尚未初始化")
		}
		return nil, apperror.Internal(err, "取得網站設定失敗")
	}
	return setting, nil
}

// Update 更新設定
func (s *settingService) Update(in model.SiteSetting) (*model.SiteSetting, error) {
	if in.Title == "" {
		return nil, apperror.BadRequest("網站標題不可為空")
	}
	if err := s.repo.Save(&in); err != nil {
		return nil, apperror.Internal(err, "更新網站設定失敗")
	}
	return &in, nil
}
