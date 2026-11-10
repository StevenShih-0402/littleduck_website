// backend/internal/model/comment.go
// Comment model：留言資料表定義，對齊前端 lake.js 的留言欄位（含審核狀態與管理員回覆）

package model

import (
	"time"

	"gorm.io/gorm"
)

// 留言狀態常數
const (
	CommentStatusPending = "pending" // 待審核（預設）
	CommentStatusVisible = "visible" // 公開可見
	CommentStatusHidden  = "hidden"  // 已隱藏
)

// Comment 留言 model
//
// id 沿用前端字串 id（c1…）作主鍵；ArticleID 可空字串＝留言板（非掛在特定文章）。
type Comment struct {
	ID        string         `gorm:"primaryKey;size:64" json:"id"`         // 留言 id（如 c1）
	Name      string         `gorm:"size:255;not null" json:"name"`        // 留言者名稱
	Date      string         `gorm:"size:64" json:"date"`                  // 顯示用日期字串
	Text      string         `gorm:"type:text;not null" json:"text"`       // 留言內容
	Status    string         `gorm:"size:16;index;not null" json:"status"` // pending / visible / hidden
	Reply     string         `gorm:"type:text" json:"reply"`               // 管理員回覆
	ArticleID string         `gorm:"size:64;index" json:"articleId"`       // 所屬文章 id；空字串＝留言板
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

func init() {
	Register(&Comment{})
}
