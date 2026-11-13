// backend/internal/middleware/auth_test.go
// auth middleware 單元測試：有效權杖放行、缺失／格式錯誤／無效（過期）回 401

package middleware

import (
	"encoding/json"
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

// performOptionalAuth 掛上 OptionalAuth 後發送請求，回傳狀態碼與 handler 看到的管理員身分（無則空字串）
func performOptionalAuth(svc auth.Service, header string) (int, string) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(ErrorHandler(zap.NewNop()))
	r.GET("/public", OptionalAuth(svc), func(c *gin.Context) {
		subject, _ := c.Get(ContextKeyAdmin)
		s, _ := subject.(string)
		c.JSON(http.StatusOK, gin.H{"subject": s})
	})
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/public", nil)
	if header != "" {
		req.Header.Set("Authorization", header)
	}
	r.ServeHTTP(rec, req)
	var body struct {
		Subject string `json:"subject"`
	}
	_ = json.Unmarshal(rec.Body.Bytes(), &body)
	return rec.Code, body.Subject
}

func TestOptionalAuth(t *testing.T) {
	svc := &stubAuthService{valid: "good-token", subject: "admin"}
	tests := []struct {
		name        string
		header      string
		wantSubject string
	}{
		{"有效權杖設定管理員身分", "Bearer good-token", "admin"},
		{"無標頭視同訪客", "", ""},
		{"格式錯誤視同訪客", "Token good-token", ""},
		{"只有 Bearer 無權杖視同訪客", "Bearer ", ""},
		{"無效或過期權杖視同訪客", "Bearer expired-token", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			code, subject := performOptionalAuth(svc, tt.header)

			// Assert：任何情況都放行（200），只有有效權杖才有身分
			if code != http.StatusOK {
				t.Errorf("OptionalAuth 不得中止請求，status = %d", code)
			}
			if subject != tt.wantSubject {
				t.Errorf("subject = %q, want %q", subject, tt.wantSubject)
			}
		})
	}
}

func TestAuth_StillRejectsWhatOptionalAuthLetsThrough(t *testing.T) {
	// Arrange：Auth 與 OptionalAuth 共用解析函式，確認 Auth 行為不受影響
	svc := &stubAuthService{valid: "good-token", subject: "admin"}

	// Act / Assert
	for _, h := range []string{"", "Token good-token", "Bearer expired-token"} {
		if rec := performAuth(svc, h); rec.Code != http.StatusUnauthorized {
			t.Errorf("header %q：Auth 應回 401，得到 %d", h, rec.Code)
		}
	}
}
