// backend/internal/config/config_test.go
// config 套件單元測試：驗證環境變數載入、預設值與必要變數 fail-fast

package config

import (
	"strings"
	"testing"
	"time"
)

// setRequiredEnv 設定全部必要環境變數，作為各測試的共用 Arrange 基礎
func setRequiredEnv(t *testing.T) {
	t.Helper()
	t.Setenv("DB_HOST", "localhost")
	t.Setenv("DB_USER", "duck")
	t.Setenv("DB_PASSWORD", "secret")
	t.Setenv("DB_NAME", "littleduck")
	t.Setenv("ADMIN_USERNAME", "admin")
	t.Setenv("ADMIN_PASSWORD", "admin-pass")
	t.Setenv("JWT_SECRET", "test-secret")
	t.Setenv("JWT_TTL", "24h")
}

func TestLoad_ReadsAllEnvironmentVariables(t *testing.T) {
	// Arrange
	setRequiredEnv(t)
	t.Setenv("PORT", "9000")
	t.Setenv("GIN_MODE", "release")
	t.Setenv("DB_PORT", "5433")
	t.Setenv("DB_SSLMODE", "require")
	t.Setenv("LOG_LEVEL", "warn")
	t.Setenv("ECPAY_MERCHANT_ID", "2000132")
	t.Setenv("ECPAY_HASH_KEY", "key")
	t.Setenv("ECPAY_HASH_IV", "iv")
	t.Setenv("ECPAY_ENV", "production")

	// Act
	cfg, err := Load()

	// Assert
	if err != nil {
		t.Fatalf("Load() 應成功，卻回傳錯誤: %v", err)
	}
	if cfg.Server.Port != "9000" || cfg.Server.GinMode != "release" {
		t.Errorf("Server 組態錯誤: %+v", cfg.Server)
	}
	if cfg.DB.Host != "localhost" || cfg.DB.Port != "5433" || cfg.DB.User != "duck" ||
		cfg.DB.Password != "secret" || cfg.DB.Name != "littleduck" || cfg.DB.SSLMode != "require" {
		t.Errorf("DB 組態錯誤: %+v", cfg.DB)
	}
	if cfg.Log.Level != "warn" {
		t.Errorf("Log 組態錯誤: %+v", cfg.Log)
	}
	if cfg.ECPay.MerchantID != "2000132" || cfg.ECPay.HashKey != "key" ||
		cfg.ECPay.HashIV != "iv" || cfg.ECPay.Env != "production" {
		t.Errorf("ECPay 組態錯誤: %+v", cfg.ECPay)
	}
	if cfg.Auth.AdminUsername != "admin" || cfg.Auth.AdminPassword != "admin-pass" ||
		cfg.Auth.JWTSecret != "test-secret" || cfg.Auth.JWTTTL != 24*time.Hour {
		t.Errorf("Auth 組態錯誤: %+v", cfg.Auth)
	}
}

func TestLoad_ParsesJWTTTLAsSecondsWhenBareInteger(t *testing.T) {
	// Arrange
	setRequiredEnv(t)
	t.Setenv("JWT_TTL", "3600")

	// Act
	cfg, err := Load()

	// Assert
	if err != nil {
		t.Fatalf("Load() 應成功，卻回傳錯誤: %v", err)
	}
	if cfg.Auth.JWTTTL != time.Hour {
		t.Errorf("JWT_TTL=3600 應解析為 1h，得到 %v", cfg.Auth.JWTTTL)
	}
}

func TestLoad_FailsWhenJWTTTLInvalid(t *testing.T) {
	// Arrange
	setRequiredEnv(t)
	t.Setenv("JWT_TTL", "not-a-duration")

	// Act
	_, err := Load()

	// Assert
	if err == nil {
		t.Fatal("JWT_TTL 格式錯誤時 Load() 應回傳錯誤")
	}
	if !strings.Contains(err.Error(), "JWT_TTL") {
		t.Errorf("錯誤訊息應提及 JWT_TTL，得到 %q", err.Error())
	}
}

func TestLoad_FailsFastWhenAuthVariableMissing(t *testing.T) {
	// Arrange
	setRequiredEnv(t)
	t.Setenv("JWT_SECRET", "")

	// Act
	_, err := Load()

	// Assert
	if err == nil {
		t.Fatal("缺少 JWT_SECRET 時 Load() 應回傳錯誤")
	}
	if !strings.Contains(err.Error(), "JWT_SECRET") {
		t.Errorf("錯誤訊息應指出缺少 JWT_SECRET，得到 %q", err.Error())
	}
}

func TestLoad_AppliesDefaultsForOptionalVariables(t *testing.T) {
	// Arrange：只設必要變數，選填變數清空以觸發預設值
	setRequiredEnv(t)
	t.Setenv("PORT", "")
	t.Setenv("GIN_MODE", "")
	t.Setenv("DB_PORT", "")
	t.Setenv("DB_SSLMODE", "")
	t.Setenv("LOG_LEVEL", "")
	t.Setenv("ECPAY_ENV", "")

	// Act
	cfg, err := Load()

	// Assert
	if err != nil {
		t.Fatalf("Load() 應成功，卻回傳錯誤: %v", err)
	}
	if cfg.Server.Port != "8080" {
		t.Errorf("PORT 預設值應為 8080，得到 %q", cfg.Server.Port)
	}
	if cfg.Server.GinMode != "debug" {
		t.Errorf("GIN_MODE 預設值應為 debug，得到 %q", cfg.Server.GinMode)
	}
	if cfg.DB.Port != "5432" {
		t.Errorf("DB_PORT 預設值應為 5432，得到 %q", cfg.DB.Port)
	}
	if cfg.DB.SSLMode != "disable" {
		t.Errorf("DB_SSLMODE 預設值應為 disable，得到 %q", cfg.DB.SSLMode)
	}
	if cfg.Log.Level != "info" {
		t.Errorf("LOG_LEVEL 預設值應為 info，得到 %q", cfg.Log.Level)
	}
	if cfg.ECPay.Env != "stage" {
		t.Errorf("ECPAY_ENV 預設值應為 stage，得到 %q", cfg.ECPay.Env)
	}
}

func TestLoad_FailsFastWhenRequiredVariableMissing(t *testing.T) {
	// Arrange：設好必要變數後清空 DB_HOST 模擬缺漏
	setRequiredEnv(t)
	t.Setenv("DB_HOST", "")

	// Act
	cfg, err := Load()

	// Assert
	if err == nil {
		t.Fatal("缺少 DB_HOST 時 Load() 應回傳錯誤")
	}
	if cfg != nil {
		t.Errorf("失敗時不應回傳 config，得到 %+v", cfg)
	}
	if !strings.Contains(err.Error(), "DB_HOST") {
		t.Errorf("錯誤訊息應指出缺少的變數名稱，得到 %q", err.Error())
	}
}

func TestLoad_ReportsAllMissingRequiredVariables(t *testing.T) {
	// Arrange：全部必要變數皆缺漏
	t.Setenv("DB_HOST", "")
	t.Setenv("DB_USER", "")
	t.Setenv("DB_PASSWORD", "")
	t.Setenv("DB_NAME", "")

	// Act
	_, err := Load()

	// Assert
	if err == nil {
		t.Fatal("缺少全部必要變數時 Load() 應回傳錯誤")
	}
	for _, name := range []string{"DB_HOST", "DB_USER", "DB_PASSWORD", "DB_NAME"} {
		if !strings.Contains(err.Error(), name) {
			t.Errorf("錯誤訊息應包含 %s，得到 %q", name, err.Error())
		}
	}
}

func TestDBConfig_DSNComposesConnectionString(t *testing.T) {
	// Arrange
	db := DBConfig{
		Host: "db", Port: "5432", User: "duck",
		Password: "secret", Name: "littleduck", SSLMode: "disable",
	}

	// Act
	dsn := db.DSN()

	// Assert
	want := "host=db port=5432 user=duck password=secret dbname=littleduck sslmode=disable"
	if dsn != want {
		t.Errorf("DSN() = %q, want %q", dsn, want)
	}
}
