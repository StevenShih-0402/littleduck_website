// backend/internal/controller/auth_controller_test.go
// auth controller 單元測試：登入成功回 token、失敗回 401、缺欄位回 400（mock service）

package controller

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"littleduck-backend/internal/middleware"
	"littleduck-backend/internal/pkg/apperror"
	"littleduck-backend/internal/service/auth"
)

// mockAuthService 測試用 auth.Service
type mockAuthService struct {
	token    *auth.Token
	loginErr error
}

func (m *mockAuthService) Login(string, string) (*auth.Token, error) {
	return m.token, m.loginErr
}
func (m *mockAuthService) Parse(string) (string, error) { return "admin", nil }

func performLogin(svc auth.Service, body string) *httptest.ResponseRecorder {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(middleware.ErrorHandler(zap.NewNop()))
	r.POST("/api/v1/auth/login", NewAuthController(svc).Login)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(rec, req)
	return rec
}

func TestLogin_SuccessReturnsToken(t *testing.T) {
	// Arrange
	svc := &mockAuthService{token: &auth.Token{Token: "jwt-abc", ExpiresAt: time.Now().Add(time.Hour)}}

	// Act
	rec := performLogin(svc, `{"username":"admin","password":"s3cret"}`)

	// Assert
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	var body struct {
		Data auth.Token `json:"data"`
	}
	_ = json.Unmarshal(rec.Body.Bytes(), &body)
	if body.Data.Token != "jwt-abc" {
		t.Errorf("回應應含 token，得到 %+v", body.Data)
	}
}

func TestLogin_BadCredentialsReturns401(t *testing.T) {
	// Arrange
	svc := &mockAuthService{loginErr: apperror.Unauthorized("帳號或密碼錯誤")}

	// Act
	rec := performLogin(svc, `{"username":"admin","password":"wrong"}`)

	// Assert
	if rec.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401", rec.Code)
	}
}

func TestLogin_MissingFieldsReturns400(t *testing.T) {
	// Arrange
	svc := &mockAuthService{}

	// Act
	rec := performLogin(svc, `{"username":"admin"}`)

	// Assert
	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", rec.Code)
	}
}
