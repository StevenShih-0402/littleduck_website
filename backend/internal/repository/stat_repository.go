// backend/internal/repository/stat_repository.go
// Stat repository：瀏覽事件寫入與每日聚合查詢

package repository

import (
	"gorm.io/gorm"

	"littleduck-backend/internal/model"
)

// DailyStat 單日聚合結果
type DailyStat struct {
	Day      string `json:"day"`
	Views    int    `json:"views"`
	Visitors int    `json:"visitors"`
}

// StatRepository 統計資料存取介面
type StatRepository interface {
	// RecordView 寫入一筆瀏覽事件
	RecordView(v *model.PageView) error
	// AggregateSince 聚合 since（含）之後的每日 views 與 distinct visitors，依日期升冪
	AggregateSince(since string) ([]DailyStat, error)
}

// gormStatRepository StatRepository 的 GORM 實作
type gormStatRepository struct {
	db *gorm.DB
}

// NewStatRepository 建立 GORM 版 StatRepository
func NewStatRepository(db *gorm.DB) StatRepository {
	return &gormStatRepository{db: db}
}

// RecordView 寫入瀏覽事件
func (r *gormStatRepository) RecordView(v *model.PageView) error {
	return r.db.Create(v).Error
}

// AggregateSince 以 GROUP BY day 聚合
func (r *gormStatRepository) AggregateSince(since string) ([]DailyStat, error) {
	var rows []DailyStat
	err := r.db.Model(&model.PageView{}).
		Select("day, COUNT(*) AS views, COUNT(DISTINCT visitor_key) AS visitors").
		Where("day >= ?", since).
		Group("day").
		Order("day ASC").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	return rows, nil
}
