// backend/internal/model/pageview.go
// PageView model：瀏覽事件原始紀錄；stat 端點以 GROUP BY day 聚合 views、distinct visitor_key 聚合 visitors

package model

// PageView 單筆瀏覽事件
//
// 僅保留去識別化的 VisitorKey（sha256(ip+ua) 截短），不存原始 IP（隱私）。
type PageView struct {
	Base
	Day        string `gorm:"size:10;index;not null" json:"day"`         // 事件日期（YYYY-MM-DD）
	VisitorKey string `gorm:"size:32;index;not null" json:"visitor_key"` // 去識別化訪客鍵
	Path       string `gorm:"size:255" json:"path"`                      // 瀏覽路徑
}

func init() {
	Register(&PageView{})
}
