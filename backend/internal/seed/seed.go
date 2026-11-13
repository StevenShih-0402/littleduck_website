// backend/internal/seed/seed.go
// 啟動 seed：各資料域為空時，以前端 store/lake.js 的既有種子內容寫入後端，
// 確保切換資料層後前台內容與原型一致。各表獨立判斷，已有資料則略過。

package seed

import (
	"fmt"
	"strconv"
	"strings"

	"go.uber.org/zap"
	"gorm.io/gorm"

	"littleduck-backend/internal/repository"
)

// FeatureEnsurer 補齊功能開關預設值（由 service.FeatureService 實作）
type FeatureEnsurer interface {
	EnsureDefaults() error
}

// Repositories seed 所需的各域 repository（以 Count/Exists 判斷資料表是否已有資料）
type Repositories struct {
	Article  repository.ArticleRepository
	Comment  repository.CommentRepository
	Setting  repository.SettingRepository
	Donation repository.DonationRepository
	// Features 功能開關補齊；與其他域不同，每次啟動都執行（只補缺、不覆寫），讓後續新增的 key 能進既有資料庫
	Features FeatureEnsurer
}

// Run 對每個為空的資料域寫入種子資料；任一步驟失敗即回傳錯誤
func Run(db *gorm.DB, repos Repositories, log *zap.Logger) error {
	if err := seedArticles(db, repos.Article, log); err != nil {
		return err
	}
	if err := seedComments(db, repos.Comment, log); err != nil {
		return err
	}
	if err := seedSetting(db, repos.Setting, log); err != nil {
		return err
	}
	if err := seedDonations(db, repos.Donation, log); err != nil {
		return err
	}
	return seedFeatures(repos.Features, log)
}

// seedFeatures 補齊功能開關；未提供 ensurer 時略過
func seedFeatures(f FeatureEnsurer, log *zap.Logger) error {
	if f == nil {
		return nil
	}
	if err := f.EnsureDefaults(); err != nil {
		return fmt.Errorf("補齊功能開關失敗: %w", err)
	}
	log.Info("功能開關補齊完成")
	return nil
}

func seedArticles(db *gorm.DB, repo repository.ArticleRepository, log *zap.Logger) error {
	n, err := repo.Count()
	if err != nil {
		return fmt.Errorf("檢查文章資料表失敗: %w", err)
	}
	if n > 0 {
		return nil
	}
	articles := buildArticles()
	if err := db.Create(&articles).Error; err != nil {
		return fmt.Errorf("寫入文章種子失敗: %w", err)
	}
	log.Info("seed 文章完成", zap.Int("count", len(articles)))
	return nil
}

func seedComments(db *gorm.DB, repo repository.CommentRepository, log *zap.Logger) error {
	n, err := repo.Count()
	if err != nil {
		return fmt.Errorf("檢查留言資料表失敗: %w", err)
	}
	if n > 0 {
		return nil
	}
	comments := buildComments(buildArticles())
	if err := db.Create(&comments).Error; err != nil {
		return fmt.Errorf("寫入留言種子失敗: %w", err)
	}
	log.Info("seed 留言完成", zap.Int("count", len(comments)))
	return nil
}

func seedSetting(db *gorm.DB, repo repository.SettingRepository, log *zap.Logger) error {
	exists, err := repo.Exists()
	if err != nil {
		return fmt.Errorf("檢查設定資料表失敗: %w", err)
	}
	if exists {
		return nil
	}
	setting := buildSetting()
	if err := db.Create(&setting).Error; err != nil {
		return fmt.Errorf("寫入設定種子失敗: %w", err)
	}
	log.Info("seed 網站設定完成")
	return nil
}

func seedDonations(db *gorm.DB, repo repository.DonationRepository, log *zap.Logger) error {
	n, err := repo.Count()
	if err != nil {
		return fmt.Errorf("檢查贊助資料表失敗: %w", err)
	}
	if n > 0 {
		return nil
	}
	donations := buildDonations()
	if err := db.Create(&donations).Error; err != nil {
		return fmt.Errorf("寫入贊助種子失敗: %w", err)
	}
	log.Info("seed 贊助紀錄完成", zap.Int("count", len(donations)))
	return nil
}

// itoa 簡短整數轉字串
func itoa(n int) string { return strconv.Itoa(n) }

// padZero 對齊 lake.js 的 String(n).padStart(2,"0")
func padZero(n int) string { return fmt.Sprintf("%02d", n) }

// contains 判斷標題是否含子字串（對齊 lake.js title.includes）
func contains(title, sub string) bool { return strings.Contains(title, sub) }
