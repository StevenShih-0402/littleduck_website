// backend/internal/middleware/auth_test.go
// auth middleware 單元測試：有效權杖放行、缺失／格式錯誤／無效（過期）回 401

package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"littleduck-backend/internal/pkg/apperror"
	"littleduck-backend/internal/service/auth"
)

// stubAuthService 測試用 auth.Service：以固定有效權杖字串判斷
type stubAuthService struct {
	valid   string
	subject string
}

func (s *stubAuthService) Login(string, string) (*auth.Token, error) { return nil, nil }
func (s *stubAuthService) Parse(token string) (string, error) {
	if token == s.valid {
		return s.subject, nil
	}
	return "", apperror.Unauthorized("權杖無效或已過期")
}

// performAuth 掛上 auth middleware 與 error handler 後發送帶指定 header 的請求
func performAuth(svc auth.Service, header string) *httptest.ResponseRecorder {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(ErrorHandler(zap.NewNop()))
	r.GET("/protected", Auth(svc), func(c *gin.Context) {
		subject, _ := c.Get(ContextKeyAdmin)
		c.JSON(http.StatusOK, gin.H{"subject": subject})
	})
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	if header != "" {
		req.Header.Set("Authorization", header)
	}
	r.ServeHTTP(rec, req)
	return rec
}

func TestAuth_ValidTokenPassesThrough(t *testing.T) {
	// Arrange
	svc := &stubAuthService{valid: "good-token", subject: "admin"}

	// Act
	rec := performAuth(svc, "Bearer good-token")

	// Assert
	if rec.Code != http.StatusOK {
		t.Errorf("有效權杖應放行，status = %d", rec.Code)
	}
}

func TestAuth_MissingHeaderReturns401(t *testing.T) {
	// Arrange
	svc := &stubAuthService{valid: "good-token"}

	// Act
	rec := performAuth(svc, "")

	// Assert
	if rec.Code != http.StatusUnauthorized {
		t.Errorf("缺少 header 應回 401，status = %d", rec.Code)
	}
}

func TestAuth_MalformedHeaderReturns401(t *testing.T) {
	// Arrange
	svc := &stubAuthService{valid: "good-token"}

	// Act：缺 Bearer 前綴
	rec := performAuth(svc, "good-token")

	// Assert
	if rec.Code != http.StatusUnauthorized {
		t.Errorf("格式錯誤 header 應回 401，status = %d", rec.Code)
	}
}

func TestAuth_InvalidOrExpiredTokenReturns401(t *testing.T) {
	// Arrange
	svc := &stubAuthService{valid: "good-token"}

	// Act
	rec := performAuth(svc, "Bearer stale-token")

	// Assert
	if rec.Code != http.StatusUnauthorized {
		t.Errorf("無效／過期權杖應回 401，status = %d", rec.Code)
	}
}
