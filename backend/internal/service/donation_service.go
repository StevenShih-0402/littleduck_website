// backend/internal/service/donation_service.go
// Donation service：贊助紀錄唯讀清單

package service

import (
	"littleduck-backend/internal/model"
	"littleduck-backend/internal/pkg/apperror"
	"littleduck-backend/internal/repository"
)

// DonationService 贊助紀錄業務邏輯介面
type DonationService interface {
	// List 取得全部贊助紀錄
	List() ([]model.Donation, error)
}

// donationService DonationService 實作
type donationService struct {
	repo repository.DonationRepository
}

// NewDonationService 建立 DonationService
func NewDonationService(repo repository.DonationRepository) DonationService {
	return &donationService{repo: repo}
}

// List 取得贊助紀錄
func (s *donationService) List() ([]model.Donation, error) {
	donations, err := s.repo.FindAll()
	if err != nil {
		return nil, apperror.Internal(err, "取得贊助紀錄失敗")
	}
	return donations, nil
}
