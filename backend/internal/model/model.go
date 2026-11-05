// backend/internal/model/model.go
// model 套件基礎：共用欄位 Base 與 AutoMigrate 註冊機制

package model

import (
	"time"

	"gorm.io/gorm"
)

// Base 各 model 共用的基礎欄位
type Base struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// registry 已註冊待 AutoMigrate 的 model 清單
var registry []any

// Register 將 model 加入 AutoMigrate 清單（於各 model 檔案的 init() 呼叫）
func Register(models ...any) {
	registry = append(registry, models...)
}

// All 回傳全部已註冊的 model，供 db.AutoMigrate 使用
func All() []any {
	return registry
}
