// backend/internal/model/donation.go
// Donation model：贊助／付款紀錄，供後台唯讀清單；交易寫入屬後續 change（SDP-104）

package model

import (
	"time"

	"gorm.io/gorm"
)

// Donation 贊助紀錄 model
//
// id 沿用前端字串 id（d1…）作主鍵。
type Donation struct {
	ID        string         `gorm:"primaryKey;size:64" json:"id"` // 贊助紀錄 id（如 d1）
	Date      string         `gorm:"size:64" json:"date"`          // 顯示用日期字串
	Amt       int            `json:"amt"`                          // 金額
	Name      string         `gorm:"size:255" json:"name"`         // 贊助者名稱
	Note      string         `gorm:"type:text" json:"note"`        // 備註
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

func init() {
	Register(&Donation{})
}
