// backend/internal/model/site_setting.go
// SiteSetting model：網站設定單列（id=1），巢狀內容（社群連結、贊助級距）以 JSON 欄位存

package model

import (
	"time"

	"gorm.io/gorm"
)

// SiteSettingID 設定固定為單一列，主鍵恆為 1
const SiteSettingID uint = 1

// SocialLink 社群連結
type SocialLink struct {
	Label string `json:"label"`
	URL   string `json:"url"`
}

// SponsorTier 贊助級距
type SponsorTier struct {
	Amt   int    `json:"amt"`
	Label string `json:"label"`
}

// Sponsor 贊助區塊設定
type Sponsor struct {
	Title    string        `json:"title"`
	Subtitle string        `json:"subtitle"`
	Tiers    []SponsorTier `json:"tiers"`
}

// SiteSetting 網站設定 model（單列）
type SiteSetting struct {
	ID           uint           `gorm:"primaryKey" json:"-"`
	Title        string         `gorm:"size:255" json:"title"`
	TitleEn      string         `gorm:"size:255" json:"titleEn"`
	Owner        string         `gorm:"size:255" json:"owner"`
	Aka          string         `gorm:"size:255" json:"aka"`
	Tagline      string         `gorm:"size:255" json:"tagline"`
	Subtitle     string         `gorm:"size:255" json:"subtitle"`
	Motto        string         `gorm:"size:255" json:"motto"`
	Announcement string         `gorm:"type:text" json:"announcement"`
	Social       []SocialLink   `gorm:"serializer:json" json:"social"`
	CatsBlog     []string       `gorm:"serializer:json" json:"catsBlog"`
	CatsWorks    []string       `gorm:"serializer:json" json:"catsWorks"`
	Sponsor      Sponsor        `gorm:"serializer:json" json:"sponsor"`
	CreatedAt    time.Time      `json:"-"`
	UpdatedAt    time.Time      `json:"-"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`
}

func init() {
	Register(&SiteSetting{})
}
