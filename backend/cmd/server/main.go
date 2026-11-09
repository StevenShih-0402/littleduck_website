// backend/cmd/server/main.go
// 服務進入點：載入 config → 初始化 logger → 連線資料庫（AutoMigrate）→ 組裝 router → 啟動 HTTP 伺服器

package main

import (
	"log"

	"go.uber.org/zap"

	"littleduck-backend/internal/config"
	"littleduck-backend/internal/database"
	"littleduck-backend/internal/pkg/logger"
	"littleduck-backend/internal/router"
)

// @title			LittleDuck API
// @version		0.1.0
// @description	LittleDuck 個人網站後端 RESTful API
// @BasePath		/
func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("組態載入失敗: %v", err)
	}

	zapLog, err := logger.New(cfg.Server.GinMode, cfg.Log.Level)
	if err != nil {
		log.Fatalf("logger 初始化失敗: %v", err)
	}
	defer func() { _ = zapLog.Sync() }()

	db, err := database.Connect(cfg.DB, zapLog)
	if err != nil {
		zapLog.Fatal("資料庫連線失敗", zap.Error(err))
	}
	if err := database.AutoMigrate(db, zapLog); err != nil {
		zapLog.Fatal("資料庫遷移失敗", zap.Error(err))
	}

	r := router.New(cfg, zapLog, db)

	zapLog.Info("伺服器啟動", zap.String("port", cfg.Server.Port), zap.String("mode", cfg.Server.GinMode))
	if err := r.Run(":" + cfg.Server.Port); err != nil {
		zapLog.Fatal("伺服器異常終止", zap.Error(err))
	}
}
