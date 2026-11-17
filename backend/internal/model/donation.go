// backend/internal/model/donation.go
// Donation model：贊助／付款紀錄；綠界付款建立時為 pending，由付款結果通知轉 paid／failed（SDP-104）

package model

import (
	"time"

	"gorm.io/gorm"
)

// 贊助付款狀態常數
const (
	DonationStatusPending = "pending" // 已建立、等待綠界通知付款結果
	DonationStatusPaid    = "paid"    // 已付款（升級前既有紀錄亦為此值）
	DonationStatusFailed  = "failed"  // 綠界通知付款失敗
)

// Donation 贊助紀錄 model
//
// id 沿用前端字串 id（d1…）作主鍵；綠界付款建立的紀錄以 MerchantTradeNo 為 id。
// Status 的資料庫預設為 paid，讓 AutoMigrate 為既有列補值；新建付款時明確寫 pending。
type Donation struct {
	ID           string         `gorm:"primaryKey;size:64" json:"id"`                // 贊助紀錄 id（如 d1 或綠界 MerchantTradeNo）
	Date         string         `gorm:"size:64" json:"date"`                         // 顯示用日期字串
	Amt          int            `json:"amt"`                                         // 金額
	Name         string         `gorm:"size:255" json:"name"`                        // 贊助者名稱
	Note         string         `gorm:"type:text" json:"note"`                       // 備註
	Status       string         `gorm:"size:16;not null;default:paid" json:"status"` // pending / paid / failed
	PaidAt       *time.Time     `json:"paid_at"`                                     // 付款完成時間
	EcpayTradeNo string         `gorm:"size:32" json:"ecpay_trade_no"`               // 綠界交易編號（TradeNo）
	PaymentType  string         `gorm:"size:32" json:"payment_type"`                 // 綠界付款方式（PaymentType）
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`
}

func init() {
	Register(&Donation{})
}
