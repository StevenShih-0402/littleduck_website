// backend/internal/model/feature_flag.go
// FeatureFlag model：功能開關資料表，一列一開關；名稱與說明放程式碼登記表，不入庫

package model

import "time"

// FeatureFlag 功能開關 model
//
// enabled 刻意不加 default tag：GORM 會把 false 當零值略過而套用預設值，建立時必須明確寫值。
type FeatureFlag struct {
	Key       string    `gorm:"primaryKey;size:64" json:"key"`
	Enabled   bool      `gorm:"not null" json:"enabled"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func init() {
	Register(&FeatureFlag{})
}
