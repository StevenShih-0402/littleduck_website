// backend/internal/model/article.go
// Article model：文章資料表定義，對齊前端 lake.js 的文章欄位（含發布狀態與前台渲染欄位）

package model

import (
	"time"

	"gorm.io/gorm"
)

// 文章狀態常數
const (
	ArticleStatusDraft     = "draft"
	ArticleStatusPublished = "published"
)

// Article 文章 model
//
// 文章主鍵沿用前端的字串 id（a0、a1…、p1…），故不內嵌 model.Base（其 ID 為自增整數），
// 改以字串 ID 當主鍵，並自行帶時間戳與軟刪除欄位，維持與其他 model 一致的生命週期欄位。
type Article struct {
	ID         string         `gorm:"primaryKey;size:64" json:"id"`         // 文章 id（如 a0、p1）
	Title      string         `gorm:"size:255;not null" json:"title"`       // 標題
	Cat        string         `gorm:"size:64;index" json:"cat"`             // 部落格分類（小說／隨筆／技術）
	Date       string         `gorm:"size:64" json:"date"`                  // 顯示用日期字串
	Y          int            `json:"y"`                                    // 年份
	Status     string         `gorm:"size:16;index;not null" json:"status"` // draft / published
	Views      int            `json:"views"`                                // 瀏覽數
	Blurb      string         `gorm:"type:text" json:"blurb"`               // 摘要
	Award      string         `gorm:"size:255" json:"award"`                // 得獎資訊或技術棧
	Banner     string         `gorm:"size:255" json:"banner"`               // 橫幅圖
	BannerFade bool           `json:"bannerFade"`                           // 橫幅是否淡出
	Showcase   bool           `json:"showcase"`                             // 是否顯示於作品展示
	WorkCat    string         `gorm:"size:64" json:"workCat"`               // 作品展示分類
	Gallery    []string       `gorm:"serializer:json" json:"gallery"`       // 展示用封面圖
	Body       string         `gorm:"type:text" json:"body"`                // 內文（Markdown）
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
	DeletedAt  gorm.DeletedAt `gorm:"index" json:"-"`
}

func init() {
	Register(&Article{})
}
