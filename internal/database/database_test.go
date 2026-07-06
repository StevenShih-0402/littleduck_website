// backend/internal/database/database_test.go
// database 套件單元測試：驗證連線重試策略（注入假 open 函式，不需真實資料庫）

package database

import (
	"errors"
	"testing"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"

	"littleduck-backend/internal/config"
)

func TestConnectWith_RetriesUntilSuccess(t *testing.T) {
	// Arrange：前兩次失敗、第三次成功的假 open
	cfg := config.DBConfig{Host: "db", Port: "5432", User: "u", Password: "p", Name: "n", SSLMode: "disable"}
	attempts := 0
	open := func(dsn string) (*gorm.DB, error) {
		attempts++
		if attempts < 3 {
			return nil, errors.New("connection refused")
		}
		return &gorm.DB{}, nil
	}

	// Act
	db, err := connectWith(cfg, zap.NewNop(), open, time.Millisecond)

	// Assert
	if err != nil {
		t.Fatalf("重試後應成功，卻回傳錯誤: %v", err)
	}
	if db == nil {
		t.Fatal("應回傳 db 實例")
	}
	if attempts != 3 {
		t.Errorf("open 應被呼叫 3 次，實際 %d 次", attempts)
	}
}

func TestConnectWith_FailsAfterRetriesExhausted(t *testing.T) {
	// Arrange：永遠失敗的假 open
	cfg := config.DBConfig{Host: "db", Port: "5432", User: "u", Password: "p", Name: "n", SSLMode: "disable"}
	attempts := 0
	open := func(dsn string) (*gorm.DB, error) {
		attempts++
		return nil, errors.New("connection refused")
	}

	// Act
	_, err := connectWith(cfg, zap.NewNop(), open, time.Millisecond)

	// Assert
	if err == nil {
		t.Fatal("重試耗盡後應回傳錯誤")
	}
	if attempts != maxRetries {
		t.Errorf("open 應被呼叫 %d 次，實際 %d 次", maxRetries, attempts)
	}
}

func TestConnectWith_PassesDSNToOpen(t *testing.T) {
	// Arrange
	cfg := config.DBConfig{Host: "db", Port: "5432", User: "u", Password: "p", Name: "n", SSLMode: "disable"}
	var gotDSN string
	open := func(dsn string) (*gorm.DB, error) {
		gotDSN = dsn
		return &gorm.DB{}, nil
	}

	// Act
	_, err := connectWith(cfg, zap.NewNop(), open, time.Millisecond)

	// Assert
	if err != nil {
		t.Fatalf("應成功，卻回傳錯誤: %v", err)
	}
	if gotDSN != cfg.DSN() {
		t.Errorf("open 收到的 DSN = %q, want %q", gotDSN, cfg.DSN())
	}
}
