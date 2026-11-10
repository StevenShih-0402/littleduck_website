// backend/internal/service/stat_service.go
// Stat service：瀏覽事件記錄（visitor_key = sha256(ip+ua) 截短）與每日統計聚合

package service

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"littleduck-backend/internal/model"
	"littleduck-backend/internal/pkg/apperror"
	"littleduck-backend/internal/repository"
)

// statsWindowDays 統計回看天數
const statsWindowDays = 30

// StatService 統計業務邏輯介面
type StatService interface {
	// RecordView 記錄一次瀏覽；clientIP 與 userAgent 不落地，僅用於去識別化雜湊
	RecordView(clientIP, userAgent, path string) error
	// Daily 回傳最近 statsWindowDays 天的每日 views／visitors（日期升冪）
	Daily() ([]repository.DailyStat, error)
}

// statService StatService 實作
type statService struct {
	repo repository.StatRepository
	now  func() time.Time
}

// NewStatService 建立 StatService
func NewStatService(repo repository.StatRepository) StatService {
	return &statService{repo: repo, now: time.Now}
}

// RecordView 以 sha256(ip+ua)[:16] 當去識別化訪客鍵，寫入一筆事件
func (s *statService) RecordView(clientIP, userAgent, path string) error {
	key := visitorKey(clientIP, userAgent)
	v := &model.PageView{
		Day:        s.now().Format("2006-01-02"),
		VisitorKey: key,
		Path:       path,
	}
	if err := s.repo.RecordView(v); err != nil {
		return apperror.Internal(err, "記錄瀏覽事件失敗")
	}
	return nil
}

// Daily 聚合最近 statsWindowDays 天的統計
func (s *statService) Daily() ([]repository.DailyStat, error) {
	since := s.now().AddDate(0, 0, -(statsWindowDays - 1)).Format("2006-01-02")
	rows, err := s.repo.AggregateSince(since)
	if err != nil {
		return nil, apperror.Internal(err, "取得統計失敗")
	}
	if rows == nil {
		rows = []repository.DailyStat{}
	}
	return rows, nil
}

// visitorKey = hex(sha256(ip + "|" + ua))[:16]，不含原始 IP
func visitorKey(clientIP, userAgent string) string {
	sum := sha256.Sum256([]byte(fmt.Sprintf("%s|%s", clientIP, userAgent)))
	return hex.EncodeToString(sum[:])[:16]
}
