// backend/internal/service/feature_service.go
// Feature service：功能開關業務邏輯——登記表、公開對照表、後台清單與切換、啟動補齊

package service

import (
	"errors"
	"time"

	"littleduck-backend/internal/model"
	"littleduck-backend/internal/pkg/apperror"
	"littleduck-backend/internal/repository"
)

// 功能開關 key 常數
const (
	FeatureSponsor         = "sponsor"
	FeatureGuestbook       = "guestbook"
	FeatureArticleComments = "article_comments"
)

// FeatureDef 開關登記表的一列；名稱與說明放程式碼，不入庫
type FeatureDef struct {
	Key         string
	Name        string
	Description string
	Default     bool
}

// featureRegistry 有序的開關登記表；新增功能只要在此加一列
var featureRegistry = []FeatureDef{
	{Key: FeatureSponsor, Name: "贊助按鈕", Description: "前台右下角的贊助浮動按鈕與贊助視窗", Default: true},
	{Key: FeatureGuestbook, Name: "留言板投書", Description: "留言板頁面的投書表單（已公開的留言照常顯示）", Default: true},
	{Key: FeatureArticleComments, Name: "文章留言", Description: "文章頁底部的留言表單（已公開的留言照常顯示）", Default: true},
}

// FeatureView 後台清單的單筆開關（登記表資訊 + 目前狀態）
type FeatureView struct {
	Key         string    `json:"key"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Enabled     bool      `json:"enabled"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// FeatureChecker 只暴露「某功能是否啟用」的查詢，供其他 service 注入
type FeatureChecker interface {
	// IsEnabled 查詢開關；登記表沒有的 key 視為停用
	IsEnabled(key string) (bool, error)
}

// FeatureService 功能開關業務邏輯介面
type FeatureService interface {
	FeatureChecker
	// PublicMap 公開對照表 {key: enabled}，只含登記表內的 key
	PublicMap() (map[string]bool, error)
	// List 後台清單，依登記表順序
	List() ([]FeatureView, error)
	// Update 切換開關；enabled 為 nil（缺欄位）回 400，未知 key 回 404
	Update(key string, enabled *bool) (*FeatureView, error)
	// EnsureDefaults 補齊資料庫缺少的 key（預設值），不覆寫既有值
	EnsureDefaults() error
}

// featureService FeatureService 實作
type featureService struct {
	repo repository.FeatureRepository
}

// NewFeatureService 建立 FeatureService
func NewFeatureService(repo repository.FeatureRepository) FeatureService {
	return &featureService{repo: repo}
}

// lookupDef 以 key 查登記表
func lookupDef(key string) (FeatureDef, bool) {
	for _, d := range featureRegistry {
		if d.Key == key {
			return d, true
		}
	}
	return FeatureDef{}, false
}

// states 讀出資料庫現況並以登記表為準：多餘 key 忽略、缺漏 key 回預設值
func (s *featureService) states() ([]FeatureView, error) {
	rows, err := s.repo.FindAll()
	if err != nil {
		return nil, apperror.Internal(err, "取得功能開關失敗")
	}
	byKey := make(map[string]model.FeatureFlag, len(rows))
	for _, r := range rows {
		byKey[r.Key] = r
	}
	out := make([]FeatureView, 0, len(featureRegistry))
	for _, d := range featureRegistry {
		v := FeatureView{Key: d.Key, Name: d.Name, Description: d.Description, Enabled: d.Default}
		if r, ok := byKey[d.Key]; ok {
			v.Enabled = r.Enabled
			v.UpdatedAt = r.UpdatedAt
		}
		out = append(out, v)
	}
	return out, nil
}

// PublicMap 公開對照表
func (s *featureService) PublicMap() (map[string]bool, error) {
	views, err := s.states()
	if err != nil {
		return nil, err
	}
	m := make(map[string]bool, len(views))
	for _, v := range views {
		m[v.Key] = v.Enabled
	}
	return m, nil
}

// List 後台清單
func (s *featureService) List() ([]FeatureView, error) {
	return s.states()
}

// IsEnabled 查詢單一開關
func (s *featureService) IsEnabled(key string) (bool, error) {
	def, ok := lookupDef(key)
	if !ok {
		return false, nil
	}
	row, err := s.repo.Find(key)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return def.Default, nil
		}
		return false, apperror.Internal(err, "取得功能開關失敗")
	}
	return row.Enabled, nil
}

// Update 切換開關
func (s *featureService) Update(key string, enabled *bool) (*FeatureView, error) {
	def, ok := lookupDef(key)
	if !ok {
		return nil, apperror.NotFound("功能開關不存在")
	}
	if enabled == nil {
		return nil, apperror.BadRequest("請提供 enabled 布林值")
	}
	row, err := s.repo.Find(key)
	if err != nil {
		if !errors.Is(err, repository.ErrNotFound) {
			return nil, apperror.Internal(err, "取得功能開關失敗")
		}
		row = &model.FeatureFlag{Key: key}
	}
	row.Enabled = *enabled
	if err := s.repo.Save(row); err != nil {
		return nil, apperror.Internal(err, "更新功能開關失敗")
	}
	return &FeatureView{
		Key: def.Key, Name: def.Name, Description: def.Description,
		Enabled: row.Enabled, UpdatedAt: row.UpdatedAt,
	}, nil
}

// EnsureDefaults 補齊缺少的 key，已存在的列不覆寫
func (s *featureService) EnsureDefaults() error {
	for _, d := range featureRegistry {
		if err := s.repo.CreateIfMissing(&model.FeatureFlag{Key: d.Key, Enabled: d.Default}); err != nil {
			return apperror.Internal(err, "補齊功能開關失敗")
		}
	}
	return nil
}
