// backend/internal/repository/donation_repository.go
// Donation repository：贊助紀錄清單、建立、查詢與付款狀態的條件更新（SDP-104）

package repository

import (
	"errors"
	"time"

	"gorm.io/gorm"

	"littleduck-backend/internal/model"
)

// ErrDuplicate repository 層的「主鍵重複」標記錯誤（如交易編號碰撞）
var ErrDuplicate = errors.New("資料重複")

// DonationRepository 贊助紀錄資料存取介面
type DonationRepository interface {
	// FindAll 取得全部贊助紀錄（依建立時間新到舊）
	FindAll() ([]model.Donation, error)
	Count() (int64, error)
	// Create 新增紀錄；主鍵重複回傳 ErrDuplicate
	Create(d *model.Donation) error
	// FindByID 以 id 取得單筆；查無回傳 ErrNotFound
	FindByID(id string) (*model.Donation, error)
	// MarkPaid 條件更新為 paid（WHERE id=? AND status<>'paid'）；回傳是否有列被更新
	MarkPaid(id string, paidAt time.Time, ecpayTradeNo, paymentType string) (bool, error)
	// MarkFailed 條件更新為 failed（WHERE id=? AND status='pending'）；回傳是否有列被更新
	MarkFailed(id string) (bool, error)
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

// Create 新增贊助紀錄；PostgreSQL unique violation（SQLSTATE 23505）轉為 ErrDuplicate
func (r *gormDonationRepository) Create(d *model.Donation) error {
	if err := r.db.Create(d).Error; err != nil {
		if isUniqueViolation(err) {
			return ErrDuplicate
		}
		return err
	}
	return nil
}

// FindByID 以 id 取得單筆贊助紀錄
func (r *gormDonationRepository) FindByID(id string) (*model.Donation, error) {
	var d model.Donation
	if err := r.db.First(&d, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &d, nil
}

// MarkPaid 條件更新為 paid：已是 paid 的列不會被更新（冪等，回傳 false）
func (r *gormDonationRepository) MarkPaid(id string, paidAt time.Time, ecpayTradeNo, paymentType string) (bool, error) {
	res := r.db.Model(&model.Donation{}).
		Where("id = ? AND status <> ?", id, model.DonationStatusPaid).
		Updates(map[string]any{
			"status":         model.DonationStatusPaid,
			"paid_at":        paidAt,
			"ecpay_trade_no": ecpayTradeNo,
			"payment_type":   paymentType,
		})
	if res.Error != nil {
		return false, res.Error
	}
	return res.RowsAffected > 0, nil
}

// MarkFailed 條件更新為 failed：只有 pending 的列會被更新，paid 不會被改回
func (r *gormDonationRepository) MarkFailed(id string) (bool, error) {
	res := r.db.Model(&model.Donation{}).
		Where("id = ? AND status = ?", id, model.DonationStatusPending).
		Update("status", model.DonationStatusFailed)
	if res.Error != nil {
		return false, res.Error
	}
	return res.RowsAffected > 0, nil
}

// isUniqueViolation 判斷是否為主鍵／唯一鍵衝突（不直接相依 pgconn，以 SQLState 介面判斷）
func isUniqueViolation(err error) bool {
	if errors.Is(err, gorm.ErrDuplicatedKey) {
		return true
	}
	var st interface{ SQLState() string }
	return errors.As(err, &st) && st.SQLState() == "23505"
}
