// backend/internal/database/database.go
// 資料庫初始化：GORM + PostgreSQL 連線（含重試策略）與已註冊 model 的 AutoMigrate

package database

import (
	"fmt"
	"time"

	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"littleduck-backend/internal/config"
	"littleduck-backend/internal/model"
)

// 連線重試策略（docker-compose 下等待 postgres 就緒）
const (
	maxRetries    = 10
	retryInterval = 3 * time.Second
)

// openFunc 建立 GORM 連線的函式型別，測試時可替換
type openFunc func(dsn string) (*gorm.DB, error)

// defaultOpen 以 postgres driver 開啟連線
func defaultOpen(dsn string) (*gorm.DB, error) {
	return gorm.Open(postgres.Open(dsn), &gorm.Config{})
}

// Connect 依組態連線 PostgreSQL：失敗時依重試策略重試，重試耗盡回傳錯誤
func Connect(cfg config.DBConfig, log *zap.Logger) (*gorm.DB, error) {
	return connectWith(cfg, log, defaultOpen, retryInterval)
}

// connectWith 可注入 open 函式與重試間隔的實作，供單元測試使用
func connectWith(cfg config.DBConfig, log *zap.Logger, open openFunc, interval time.Duration) (*gorm.DB, error) {
	var lastErr error
	for attempt := 1; attempt <= maxRetries; attempt++ {
		db, err := open(cfg.DSN())
		if err == nil {
			log.Info("資料庫連線成功", zap.String("host", cfg.Host), zap.String("db", cfg.Name))
			return db, nil
		}
		lastErr = err
		log.Warn("資料庫連線失敗，稍後重試",
			zap.Int("attempt", attempt),
			zap.Int("max_retries", maxRetries),
			zap.Error(err),
		)
		if attempt < maxRetries {
			time.Sleep(interval)
		}
	}
	return nil, fmt.Errorf("資料庫連線重試 %d 次後仍失敗: %w", maxRetries, lastErr)
}

// AutoMigrate 對全部已註冊的 model 執行自動遷移
func AutoMigrate(db *gorm.DB, log *zap.Logger) error {
	models := model.All()
	if err := db.AutoMigrate(models...); err != nil {
		return fmt.Errorf("AutoMigrate 失敗: %w", err)
	}
	log.Info("AutoMigrate 完成", zap.Int("models", len(models)))
	return nil
}
