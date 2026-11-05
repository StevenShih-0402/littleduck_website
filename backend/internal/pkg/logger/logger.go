// backend/internal/pkg/logger/logger.go
// 結構化 logger 初始化：dev 模式輸出彩色 console、release 模式輸出 JSON（相容 GCP Cloud Logging），等級由 LOG_LEVEL 控制

package logger

import (
	"fmt"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// New 依 Gin 模式與日誌等級建立 zap logger：
// release 模式輸出 production JSON，其餘模式輸出開發用彩色 console
func New(ginMode, level string) (*zap.Logger, error) {
	zapLevel, err := zapcore.ParseLevel(level)
	if err != nil {
		return nil, fmt.Errorf("無效的 LOG_LEVEL %q: %w", level, err)
	}

	var cfg zap.Config
	if ginMode == "release" {
		cfg = zap.NewProductionConfig()
	} else {
		cfg = zap.NewDevelopmentConfig()
		cfg.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	}
	cfg.Level = zap.NewAtomicLevelAt(zapLevel)

	return cfg.Build()
}
