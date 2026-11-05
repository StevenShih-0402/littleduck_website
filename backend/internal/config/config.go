// backend/internal/config/config.go
// 以 os.Getenv() 集中讀取全部環境變數組態，缺少必要變數時回傳錯誤以達成啟動 fail-fast

package config

import (
	"fmt"
	"os"
	"strings"
)

// ServerConfig HTTP 伺服器相關組態
type ServerConfig struct {
	Port    string // PORT，預設 8080
	GinMode string // GIN_MODE：debug / release / test，預設 debug
}

// DBConfig PostgreSQL 連線組態
type DBConfig struct {
	Host     string // DB_HOST（必要）
	Port     string // DB_PORT，預設 5432
	User     string // DB_USER（必要）
	Password string // DB_PASSWORD（必要）
	Name     string // DB_NAME（必要）
	SSLMode  string // DB_SSLMODE，預設 disable
}

// DSN 組成 GORM postgres driver 使用的連線字串
func (d DBConfig) DSN() string {
	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		d.Host, d.Port, d.User, d.Password, d.Name, d.SSLMode,
	)
}

// LogConfig 日誌組態
type LogConfig struct {
	Level string // LOG_LEVEL：debug / info / warn / error，預設 info
}

// ECPayConfig 綠界金流組態（骨架階段允許留空）
type ECPayConfig struct {
	MerchantID string // ECPAY_MERCHANT_ID
	HashKey    string // ECPAY_HASH_KEY
	HashIV     string // ECPAY_HASH_IV
	Env        string // ECPAY_ENV：stage / production，預設 stage
}

// Config 全部組態的彙總 struct，於啟動時載入一次後注入各層
type Config struct {
	Server ServerConfig
	DB     DBConfig
	Log    LogConfig
	ECPay  ECPayConfig
}

// requiredVars 啟動時必須存在的環境變數
var requiredVars = []string{"DB_HOST", "DB_USER", "DB_PASSWORD", "DB_NAME"}

// Load 讀取環境變數組成 Config；必要變數缺漏時回傳列出全部缺漏項的錯誤
func Load() (*Config, error) {
	if err := validateRequired(); err != nil {
		return nil, err
	}

	return &Config{
		Server: ServerConfig{
			Port:    getEnv("PORT", "8080"),
			GinMode: getEnv("GIN_MODE", "debug"),
		},
		DB: DBConfig{
			Host:     os.Getenv("DB_HOST"),
			Port:     getEnv("DB_PORT", "5432"),
			User:     os.Getenv("DB_USER"),
			Password: os.Getenv("DB_PASSWORD"),
			Name:     os.Getenv("DB_NAME"),
			SSLMode:  getEnv("DB_SSLMODE", "disable"),
		},
		Log: LogConfig{
			Level: getEnv("LOG_LEVEL", "info"),
		},
		ECPay: ECPayConfig{
			MerchantID: os.Getenv("ECPAY_MERCHANT_ID"),
			HashKey:    os.Getenv("ECPAY_HASH_KEY"),
			HashIV:     os.Getenv("ECPAY_HASH_IV"),
			Env:        getEnv("ECPAY_ENV", "stage"),
		},
	}, nil
}

// validateRequired 檢查必要變數，一次回報全部缺漏項
func validateRequired() error {
	var missing []string
	for _, name := range requiredVars {
		if os.Getenv(name) == "" {
			missing = append(missing, name)
		}
	}
	if len(missing) > 0 {
		return fmt.Errorf("缺少必要環境變數: %s", strings.Join(missing, ", "))
	}
	return nil
}

// getEnv 讀取環境變數，空值時回傳預設值
func getEnv(name, fallback string) string {
	if v := os.Getenv(name); v != "" {
		return v
	}
	return fallback
}
