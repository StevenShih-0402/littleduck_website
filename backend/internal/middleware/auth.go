// backend/internal/middleware/auth.go
// Auth middleware：驗證 Authorization: Bearer <jwt>，掛在受保護的後台路由群組

package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"

	"littleduck-backend/internal/pkg/apperror"
	"littleduck-backend/internal/service/auth"
)

// ContextKeyAdmin gin.Context 中存放通過驗證之管理員帳號的鍵
const ContextKeyAdmin = "admin_subject"

// 標頭解析失敗的原因（Auth 轉為 401 訊息，OptionalAuth 直接當訪客放行）
const (
	msgMissingHeader = "缺少 Authorization 標頭"
	msgBadHeader     = "Authorization 標頭格式錯誤"
)

// bearerToken 解析 Authorization 標頭取出 Bearer 權杖；失敗回傳對應的 401 訊息
func bearerToken(header string) (token, failMsg string) {
	if header == "" {
		return "", msgMissingHeader
	}
	parts := strings.SplitN(header, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") || strings.TrimSpace(parts[1]) == "" {
		return "", msgBadHeader
	}
	return strings.TrimSpace(parts[1]), ""
}

// Auth 以 auth.Service 驗證 Bearer 權杖；缺失／格式錯誤／過期一律 401 並中止
func Auth(svc auth.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		token, failMsg := bearerToken(c.GetHeader("Authorization"))
		if failMsg != "" {
			abortUnauthorized(c, failMsg)
			return
		}

		subject, err := svc.Parse(token)
		if err != nil {
			_ = c.Error(err)
			c.Abort()
			return
		}
		c.Set(ContextKeyAdmin, subject)
		c.Next()
	}
}

// OptionalAuth 選用式驗證，供公開路由辨識管理員：
// 權杖有效才設定 ContextKeyAdmin；標頭缺失、格式錯誤、權杖無效或過期一律視同訪客放行，不回 401。
func OptionalAuth(svc auth.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		if token, failMsg := bearerToken(c.GetHeader("Authorization")); failMsg == "" {
			if subject, err := svc.Parse(token); err == nil {
				c.Set(ContextKeyAdmin, subject)
			}
		}
		c.Next()
	}
}

// abortUnauthorized 以統一錯誤格式回 401 並中止
func abortUnauthorized(c *gin.Context, message string) {
	_ = c.Error(apperror.Unauthorized(message))
	c.Abort()
}
