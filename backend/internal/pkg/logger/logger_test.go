// backend/internal/pkg/logger/logger_test.go
// logger 套件單元測試：驗證 dev/prod 模式初始化與 LOG_LEVEL 過濾

package logger

import (
	"testing"

	"go.uber.org/zap/zapcore"
)

func TestNew_DevModeCreatesLogger(t *testing.T) {
	// Arrange
	ginMode := "debug"

	// Act
	log, err := New(ginMode, "info")

	// Assert
	if err != nil {
		t.Fatalf("New() 應成功，卻回傳錯誤: %v", err)
	}
	if log == nil {
		t.Fatal("New() 應回傳 logger 實例")
	}
}

func TestNew_ProdModeCreatesLogger(t *testing.T) {
	// Arrange
	ginMode := "release"

	// Act
	log, err := New(ginMode, "info")

	// Assert
	if err != nil {
		t.Fatalf("New() 應成功，卻回傳錯誤: %v", err)
	}
	if log == nil {
		t.Fatal("New() 應回傳 logger 實例")
	}
}

func TestNew_FiltersBelowConfiguredLevel(t *testing.T) {
	// Arrange
	logLevel := "warn"

	// Act
	log, err := New("release", logLevel)

	// Assert
	if err != nil {
		t.Fatalf("New() 應成功，卻回傳錯誤: %v", err)
	}
	if log.Core().Enabled(zapcore.InfoLevel) {
		t.Error("LOG_LEVEL=warn 時 info 等級不應啟用")
	}
	if !log.Core().Enabled(zapcore.WarnLevel) {
		t.Error("LOG_LEVEL=warn 時 warn 等級應啟用")
	}
}

func TestNew_InvalidLevelReturnsError(t *testing.T) {
	// Arrange
	logLevel := "not-a-level"

	// Act
	_, err := New("release", logLevel)

	// Assert
	if err == nil {
		t.Fatal("無效的 LOG_LEVEL 應回傳錯誤")
	}
}
