// backend/internal/repository/donation_repository.go
// Donation repository：贊助紀錄唯讀清單；交易寫入屬後續 change（SDP-104）

package repository

import (
	"gorm.io/gorm"

	"littleduck-backend/internal/model"
)

// DonationRepository 贊助紀錄資料存取介面
type DonationRepository interface {
	// FindAll 取得全部贊助紀錄（依建立時間新到舊）
	FindAll() ([]model.Donation, error)
	Count() (int64, error)
}

// gormDonationRepository DonationRepository 的 GORM 實作
type gormDonationRepository struct {
	db *gorm.DB
}

// NewDonationRepository 建立 GORM 版 DonationRepository
func NewDonationRepository(db *gorm.DB) DonationRepository {
	return &gormDonationRepository{db: db}
}

// FindAll 取得全部贊助紀錄
func (r *gormDonationRepository) FindAll() ([]model.Donation, error) {
	var donations []model.Donation
	if err := r.db.Order("created_at DESC").Find(&donations).Error; err != nil {
		return nil, err
	}
	return donations, nil
}

// Count 回傳贊助紀錄總數
func (r *gormDonationRepository) Count() (int64, error) {
	var n int64
	if err := r.db.Model(&model.Donation{}).Count(&n).Error; err != nil {
		return 0, err
	}
	return n, nil
}
